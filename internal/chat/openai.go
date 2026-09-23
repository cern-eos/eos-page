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
	openAIResponsesURL = "https://api.openai.com/v1/responses"
	maxToolRounds      = 4
)

var openaiHTTP = &http.Client{
	Timeout: 70 * time.Second,
	Transport: &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	},
}

type responsesReq struct {
	Model            string          `json:"model"`
	Instructions     string          `json:"instructions,omitempty"`
	Input            []responsesItem `json:"input"`
	Tools            []responsesTool `json:"tools,omitempty"`
	ToolChoice       string          `json:"tool_choice,omitempty"`
	Include          []string        `json:"include,omitempty"`
	Temperature      *float64        `json:"temperature,omitempty"`
	MaxOutputTokens  int             `json:"max_output_tokens,omitempty"`
	PreviousResponse string          `json:"previous_response_id,omitempty"`
	Reasoning        *struct {
		Effort string `json:"effort"`
	} `json:"reasoning,omitempty"`
}

type responsesTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	Strict      *bool           `json:"strict,omitempty"`
}

type responsesItem struct {
	Type      string `json:"type,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	ID        string `json:"id,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Output    string `json:"output,omitempty"`
}

type responsesAction struct {
	Type    string            `json:"type,omitempty"`
	Query   string            `json:"query,omitempty"`
	Queries []string          `json:"queries,omitempty"`
	Sources []responsesSource `json:"sources,omitempty"`
}

type responsesSource struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

type responsesContent struct {
	Type        string                `json:"type"`
	Text        string                `json:"text,omitempty"`
	Annotations []responsesAnnotation `json:"annotations,omitempty"`
}

type responsesAnnotation struct {
	Type        string `json:"type"`
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	URLCitation *struct {
		URL   string `json:"url"`
		Title string `json:"title,omitempty"`
	} `json:"url_citation,omitempty"`
}

type responsesResp struct {
	ID    string `json:"id"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Output []responsesOutputItem `json:"output"`
}

type responsesOutputItem struct {
	Type      string             `json:"type"`
	ID        string             `json:"id,omitempty"`
	CallID    string             `json:"call_id,omitempty"`
	Name      string             `json:"name,omitempty"`
	Arguments string             `json:"arguments,omitempty"`
	Role      string             `json:"role,omitempty"`
	Status    string             `json:"status,omitempty"`
	Action    *responsesAction   `json:"action,omitempty"`
	Content   []responsesContent `json:"content,omitempty"`
}

func openAITools() []responsesTool {
	strict := false
	return []responsesTool{
		{Type: "web_search"},
		{
			Type:        "function",
			Name:        "search_docs",
			Description: "Search EOS, XRootD, or CTA documentation sites and return titles with URLs.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"},"site":{"type":"string","enum":["all","eos-docs","xrootd","cta"]}},"required":["query"]}`),
			Strict:      &strict,
		},
		{
			Type:        "function",
			Name:        "fetch_doc",
			Description: "Fetch one allowed documentation page and return it as formatted Markdown (headings, lists, code).",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`),
			Strict:      &strict,
		},
		{
			Type:        "function",
			Name:        "search_ops_docs",
			Description: "Search the EOS operator documentation (all_docs.md) and return matching Markdown sections.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`),
			Strict:      &strict,
		},
	}
}

func isGPT5(model string) bool {
	return strings.Contains(strings.ToLower(model), "gpt-5")
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

	var input []responsesItem
	for _, m := range ClipHistory(history) {
		input = append(input, responsesItem{Type: "message", Role: m.Role, Content: m.Text})
	}
	input = append(input, responsesItem{Type: "message", Role: "user", Content: frameUserQuestion(question)})

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

	var prevID string
	for round := 0; round <= maxToolRounds; round++ {
		req := responsesReq{
			Model:            model,
			Instructions:     systemPromptOpenAI(siteContext, ops),
			Input:            input,
			Tools:            openAITools(),
			ToolChoice:       "auto",
			Include:          []string{"web_search_call.action.sources"},
			MaxOutputTokens:  2200,
			PreviousResponse: prevID,
		}
		if isGPT5(model) {
			req.Reasoning = &struct {
				Effort string `json:"effort"`
			}{Effort: "low"}
		} else {
			temp := 0.2
			req.Temperature = &temp
		}

		raw, err := c.completeResponses(ctx, req)
		if err != nil {
			return Reply{}, err
		}
		text, extra, calls := parseResponsesOutput(raw)
		addSources(extra)
		if len(calls) == 0 {
			if strings.TrimSpace(text) == "" {
				return Reply{}, fmt.Errorf("empty openai answer")
			}
			return Reply{Text: text, Sources: sources}, nil
		}

		prevID = raw.ID
		input = input[:0]
		for _, call := range calls {
			result, extraSrc := c.runTool(ctx, call.Name, call.Arguments)
			addSources(extraSrc)
			input = append(input, responsesItem{
				Type:   "function_call_output",
				CallID: call.ID,
				Output: result,
			})
		}
	}
	return Reply{}, fmt.Errorf("openai tool loop exceeded")
}

func parseResponsesOutput(raw responsesResp) (text string, sources []Source, calls []openaiToolCall) {
	var b strings.Builder
	seen := map[string]bool{}
	add := func(title, url string) {
		url = strings.TrimSpace(url)
		if url == "" || seen[url] {
			return
		}
		seen[url] = true
		if strings.TrimSpace(title) == "" {
			title = url
		}
		sources = append(sources, Source{Title: title, URL: url})
	}

	for _, item := range raw.Output {
		switch item.Type {
		case "web_search_call":
			if item.Action != nil {
				q := strings.TrimSpace(item.Action.Query)
				if q == "" && len(item.Action.Queries) > 0 {
					q = strings.Join(item.Action.Queries, "; ")
				}
				if q != "" {
					log.Printf("ask eos: model web_search %q", q)
				}
				for _, s := range item.Action.Sources {
					add(s.Title, s.URL)
				}
			}
		case "function_call":
			id := item.CallID
			if id == "" {
				id = item.ID
			}
			calls = append(calls, openaiToolCall{
				ID:        id,
				Name:      item.Name,
				Arguments: item.Arguments,
			})
		case "message":
			for _, part := range item.Content {
				if part.Type == "output_text" || part.Type == "text" {
					if strings.TrimSpace(part.Text) != "" {
						if b.Len() > 0 {
							b.WriteByte('\n')
						}
						b.WriteString(part.Text)
					}
				}
				for _, a := range part.Annotations {
					url, title := a.URL, a.Title
					if a.URLCitation != nil {
						if url == "" {
							url = a.URLCitation.URL
						}
						if title == "" {
							title = a.URLCitation.Title
						}
					}
					add(title, url)
				}
			}
		}
	}
	return strings.TrimSpace(b.String()), sources, calls
}

type openaiToolCall struct {
	ID        string
	Name      string
	Arguments string
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

func (c *Client) completeResponses(ctx context.Context, req responsesReq) (responsesResp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return responsesResp{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIResponsesURL, bytes.NewReader(body))
	if err != nil {
		return responsesResp{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.openaiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := openaiHTTP.Do(httpReq)
	if err != nil {
		return responsesResp{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return responsesResp{}, err
	}
	var out responsesResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return responsesResp{}, fmt.Errorf("openai decode: %w", err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return responsesResp{}, fmt.Errorf("openai: %s", out.Error.Message)
	}
	if res.StatusCode >= 300 {
		return responsesResp{}, fmt.Errorf("openai HTTP %d", res.StatusCode)
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
