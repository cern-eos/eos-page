package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOpenAIModel = "gpt-4o-mini"
	openAIURL          = "https://api.openai.com/v1/chat/completions"
	maxToolRounds      = 4
)

var openaiHTTP = &http.Client{
	Timeout: 70 * time.Second,
	Transport: &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	},
}

type openaiReq struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Tools       []openaiTool    `json:"tools,omitempty"`
	ToolChoice  string          `json:"tool_choice,omitempty"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiTool struct {
	Type     string             `json:"type"`
	Function openaiToolFunction `json:"function"`
}

type openaiToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openaiResp struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
}

func openAITools() []openaiTool {
	return []openaiTool{
		{
			Type: "function",
			Function: openaiToolFunction{
				Name:        "search_docs",
				Description: "Search EOS, XRootD, or CTA documentation sites and return titles with URLs.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"},"site":{"type":"string","enum":["all","eos-docs","xrootd","cta"]}},"required":["query"]}`),
			},
		},
		{
			Type: "function",
			Function: openaiToolFunction{
				Name:        "fetch_doc",
				Description: "Fetch one allowed documentation page and return it as formatted Markdown (headings, lists, code).",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`),
			},
		},
		{
			Type: "function",
			Function: openaiToolFunction{
				Name:        "search_ops_docs",
				Description: "Search the EOS operator documentation (all_docs.md) and return matching Markdown sections.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`),
			},
		},
	}
}

func (c *Client) askOpenAI(ctx context.Context, question, siteContext string, history []Message) (Reply, error) {
	if c == nil || c.openaiKey == "" {
		return Reply{}, fmt.Errorf("openai is not configured")
	}
	model := c.openaiModel
	if model == "" {
		model = defaultOpenAIModel
	}
	ops := ""
	if c.ops != nil {
		ops = c.ops.FormatHits(question, 2)
	}
	msgs := []openaiMessage{{Role: "system", Content: systemPromptOpenAI(siteContext, ops)}}
	for _, m := range ClipHistory(history) {
		msgs = append(msgs, openaiMessage{Role: m.Role, Content: m.Text})
	}
	msgs = append(msgs, openaiMessage{Role: "user", Content: question})

	var sources []Source
	seen := map[string]bool{}
	addSources := func(in []Source) {
		for _, s := range in {
			if s.URL == "" || seen[s.URL] {
				continue
			}
			seen[s.URL] = true
			sources = append(sources, s)
		}
	}

	for round := 0; round <= maxToolRounds; round++ {
		raw, err := c.completeOpenAI(ctx, openaiReq{
			Model:       model,
			Messages:    msgs,
			Tools:       openAITools(),
			ToolChoice:  "auto",
			Temperature: 0.2,
			MaxTokens:   1400,
		})
		if err != nil {
			return Reply{}, err
		}
		if len(raw.Choices) == 0 {
			return Reply{}, fmt.Errorf("empty openai response")
		}
		msg := raw.Choices[0].Message
		if len(msg.ToolCalls) == 0 {
			text := strings.TrimSpace(msg.Content)
			if text == "" {
				return Reply{}, fmt.Errorf("empty openai answer")
			}
			return Reply{Text: text, Sources: sources}, nil
		}
		msgs = append(msgs, msg)
		for _, call := range msg.ToolCalls {
			result, extra := c.runTool(ctx, call.Function.Name, call.Function.Arguments)
			addSources(extra)
			msgs = append(msgs, openaiMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    result,
			})
		}
	}
	return Reply{}, fmt.Errorf("openai tool loop exceeded")
}

func (c *Client) runTool(ctx context.Context, name, argsJSON string) (string, []Source) {
	var args struct {
		Query string `json:"query"`
		Site  string `json:"site"`
		URL   string `json:"url"`
	}
	_ = json.Unmarshal([]byte(argsJSON), &args)
	switch name {
	case "search_docs":
		text, src, err := searchAllowedSites(ctx, args.Query, args.Site)
		if err != nil {
			return "Search failed: " + err.Error(), nil
		}
		return text, src
	case "fetch_doc":
		text, err := fetchDoc(ctx, args.URL)
		if err != nil {
			return "Fetch failed: " + err.Error(), nil
		}
		src := []Source{}
		if u := strings.TrimSpace(args.URL); u != "" {
			src = []Source{{Title: extractDocTitle(text), URL: u}}
		}
		return text, src
	case "search_ops_docs":
		if c.ops == nil {
			return "Operator documentation is not loaded.", nil
		}
		text := c.ops.FormatHits(args.Query, 3)
		if text == "" {
			return "No matching operator documentation.", nil
		}
		return text, nil
	default:
		return "Unknown tool.", nil
	}
}

func extractDocTitle(formatted string) string {
	for _, line := range strings.Split(formatted, "\n") {
		if strings.HasPrefix(line, "### Document: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "### Document: "))
		}
	}
	return "Documentation"
}

func (c *Client) completeOpenAI(ctx context.Context, req openaiReq) (openaiResp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return openaiResp{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIURL, bytes.NewReader(body))
	if err != nil {
		return openaiResp{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.openaiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := openaiHTTP.Do(httpReq)
	if err != nil {
		return openaiResp{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return openaiResp{}, err
	}
	var out openaiResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return openaiResp{}, fmt.Errorf("openai decode: %w", err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return openaiResp{}, fmt.Errorf("openai: %s", out.Error.Message)
	}
	if res.StatusCode >= 300 {
		return openaiResp{}, fmt.Errorf("openai HTTP %d", res.StatusCode)
	}
	return out, nil
}

func (c *Client) configuredOpenAI() bool {
	return c != nil && strings.TrimSpace(c.openaiKey) != ""
}

func (c *Client) StartupProbe(ctx context.Context) {
	if !c.configuredOpenAI() {
		return
	}
	q := "In one sentence, what is CERN EOS storage?"
	log.Printf("ask eos test query: %s", q)
	reply, err := c.Ask(ctx, q, "", nil)
	if err != nil {
		log.Printf("ask eos test FAILED: %v", err)
		return
	}
	text := clipRunes(strings.Join(strings.Fields(reply.Text), " "), 400)
	log.Printf("ask eos test reply: %s", text)
	if len(reply.Sources) > 0 {
		log.Printf("ask eos test sources: %d", len(reply.Sources))
	}
}
