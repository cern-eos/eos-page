package chat

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"google.golang.org/genai"
)

const Model = "gemini-2.5-flash"

const maxQuestion = 1200
const maxHistory = 8
const maxTurn = 1600

type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type Source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Reply struct {
	Text    string   `json:"text"`
	Sources []Source `json:"sources,omitempty"`
}

type Client struct {
	api *genai.Client
}

func New(ctx context.Context, apiKey string) (*Client, error) {
	c := &Client{}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return c, nil
	}
	api, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return c, err
	}
	c.api = api
	return c, nil
}

func ClipQuestion(q string) (string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "", fmt.Errorf("empty question")
	}
	if utf8.RuneCountInString(q) > maxQuestion {
		return "", fmt.Errorf("question too long")
	}
	return q, nil
}

func ClipHistory(in []Message) []Message {
	out := make([]Message, 0, maxHistory)
	for _, m := range in {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		text := strings.TrimSpace(m.Text)
		if text == "" {
			continue
		}
		if utf8.RuneCountInString(text) > maxTurn {
			text = string([]rune(text)[:maxTurn])
		}
		out = append(out, Message{Role: role, Text: text})
	}
	if len(out) > maxHistory {
		out = out[len(out)-maxHistory:]
	}
	return out
}

func (c *Client) Ask(ctx context.Context, question, siteContext string, history []Message) (Reply, error) {
	if c == nil {
		return Reply{}, fmt.Errorf("chat is not configured")
	}
	q, err := ClipQuestion(question)
	if err != nil {
		return Reply{}, err
	}
	ai, webText, webSources, searchErr := liveSearch(ctx, q)
	if c.api == nil {
		if searchErr == nil && strings.TrimSpace(webText) != "" {
			if ai {
				return Reply{Text: webText, Sources: webSources}, nil
			}
			return fallbackReply(q, siteContext, webText, webSources, nil), nil
		}
		return fallbackReply(q, siteContext, webText, webSources, searchErr), nil
	}
	contents := make([]*genai.Content, 0, len(history)+1)
	for _, m := range ClipHistory(history) {
		var role genai.Role = genai.RoleUser
		if m.Role == "assistant" {
			role = genai.RoleModel
		}
		contents = append(contents, genai.NewContentFromText(m.Text, role))
	}
	user := q
	if searchErr == nil && webText != "" {
		label := "Google search extract"
		if ai {
			label = "Google AI Mode answer"
		}
		user = q + "\n\n" + label + ":\n" + webText
	}
	contents = append(contents, genai.NewContentFromText(user, genai.RoleUser))

	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: systemPrompt(siteContext, searchErr)}}},
	}
	if searchErr != nil {
		cfg.Tools = []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}}
	}
	resp, err := c.api.Models.GenerateContent(ctx, Model, contents, cfg)
	if err != nil {
		return fallbackReply(q, siteContext, webText, webSources, err), nil
	}
	text := strings.TrimSpace(responseText(resp))
	if text == "" {
		return fallbackReply(q, siteContext, webText, webSources, searchErr), nil
	}
	sources := webSources
	if len(sources) == 0 {
		sources = groundingSources(resp)
	}
	return Reply{Text: text, Sources: sources}, nil
}

func fallbackReply(q, siteContext, webText string, sources []Source, searchErr error) Reply {
	var b strings.Builder
	b.WriteString("Here is what I found for “")
	b.WriteString(q)
	b.WriteString("”.\n\n")
	if strings.TrimSpace(siteContext) != "" {
		b.WriteString(strings.TrimSpace(siteContext))
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(webText) != "" {
		b.WriteString(clipSearchText(webText))
	} else if searchErr != nil {
		b.WriteString("Live Google search did not return a page this time. See https://eos-docs.web.cern.ch/diopside/ or write to eos-support@cern.ch.")
	}
	return Reply{Text: strings.TrimSpace(b.String()), Sources: sources}
}

func systemPrompt(siteContext string, searchErr error) string {
	var b strings.Builder
	b.WriteString("You are the Ask EOS assistant on the EOS Open Storage website (CERN).\n")
	b.WriteString("Answer questions about EOS disk storage, XRootD, QuarkDB, FST/MGM, eosxd, CERNBox, CTA, workshops, docs and operations.\n")
	b.WriteString("Prefer facts from the Google AI Mode answer (or search extract) and the local catalogue. Cite URLs when they appear.\n")
	b.WriteString("Be concise. Prefer concrete commands, URLs and version names. If you are unsure, say so and point to https://eos-docs.web.cern.ch/diopside/ or eos-support@cern.ch.\n")
	b.WriteString("Do not invent APIs, people or workshop dates. Refuse requests unrelated to EOS or storage.\n")
	if searchErr != nil {
		b.WriteString("Google search was unavailable for this turn; answer from the catalogue and known EOS docs only.\n")
	}
	if strings.TrimSpace(siteContext) != "" {
		b.WriteString("\nLocal catalogue snippets (may be incomplete):\n")
		b.WriteString(siteContext)
	}
	return b.String()
}

func responseText(resp *genai.GenerateContentResponse) string {
	if resp == nil {
		return ""
	}
	if t := strings.TrimSpace(resp.Text()); t != "" {
		return t
	}
	var b strings.Builder
	for _, c := range resp.Candidates {
		if c == nil || c.Content == nil {
			continue
		}
		for _, p := range c.Content.Parts {
			if p != nil && p.Text != "" {
				b.WriteString(p.Text)
			}
		}
	}
	return b.String()
}

func groundingSources(resp *genai.GenerateContentResponse) []Source {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0] == nil {
		return nil
	}
	meta := resp.Candidates[0].GroundingMetadata
	if meta == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []Source
	for _, ch := range meta.GroundingChunks {
		if ch == nil || ch.Web == nil {
			continue
		}
		url := strings.TrimSpace(ch.Web.URI)
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		title := strings.TrimSpace(ch.Web.Title)
		if title == "" {
			title = url
		}
		out = append(out, Source{Title: title, URL: url})
		if len(out) >= 6 {
			break
		}
	}
	return out
}
