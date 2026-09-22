package googleai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// CustomSearch uses the official Google Programmable Search JSON API.
func CustomSearch(parent context.Context, query string) (AIResult, error) {
	query = strings.TrimSpace(query)
	key := strings.TrimSpace(os.Getenv("GOOGLE_CSE_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	}
	cx := strings.TrimSpace(os.Getenv("GOOGLE_CSE_ID"))
	if query == "" {
		return AIResult{}, fmt.Errorf("empty query")
	}
	if key == "" || cx == "" {
		return AIResult{}, fmt.Errorf("google cse not configured")
	}
	ctx := parent
	if ctx == nil {
		ctx = context.Background()
	}
	u := "https://www.googleapis.com/customsearch/v1?key=" + url.QueryEscape(key) +
		"&cx=" + url.QueryEscape(cx) + "&num=8&q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return AIResult{}, err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return AIResult{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return AIResult{}, fmt.Errorf("google cse http %s", res.Status)
	}
	var payload struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return AIResult{}, err
	}
	var sources []Source
	var b strings.Builder
	for _, it := range payload.Items {
		href := sanitizeURL(it.Link)
		if href == "" || skipURL(href) {
			continue
		}
		title := strings.TrimSpace(it.Title)
		if title == "" {
			title = hostTitle(href)
		}
		if snip := strings.TrimSpace(it.Snippet); snip != "" {
			title = title + " - " + clipRunes(snip, 180)
		}
		sources = append(sources, Source{Title: title, URL: href})
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(title)
		b.WriteByte('\n')
		b.WriteString(href)
		if len(sources) >= 8 {
			break
		}
	}
	if len(sources) == 0 {
		return AIResult{}, fmt.Errorf("empty search results")
	}
	return AIResult{Query: query, Answer: b.String(), Sources: sources, IsAI: false}, nil
}

func CSEConfigured() bool {
	cx := strings.TrimSpace(os.Getenv("GOOGLE_CSE_ID"))
	key := strings.TrimSpace(os.Getenv("GOOGLE_CSE_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	}
	return cx != "" && key != ""
}
