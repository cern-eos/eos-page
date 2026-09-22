package chat

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/apeters/eospage/internal/googleai"
)

const maxSearchText = 8000

func liveSearch(parent context.Context, question string) (ai bool, text string, sources []Source, err error) {
	q := searchQuery(question)

	if googleai.CSEConfigured() {
		if result, cseErr := googleai.CustomSearch(parent, q); cseErr == nil && strings.TrimSpace(result.Answer) != "" {
			return false, result.Answer, toChatSources(result.Sources), nil
		} else if cseErr != nil {
			log.Printf("ask eos google cse: %v", cseErr)
		}
	}

	var chromeErr error
	if os.Getenv("CHAT_SEARCH_CHROME") == "1" || os.Getenv("CHAT_SEARCH_HEADED") == "1" {
		var result googleai.AIResult
		result, chromeErr = googleai.Search(parent, q, googleai.Options{})
		if chromeErr == nil && strings.TrimSpace(result.Answer) != "" {
			return result.IsAI, result.Answer, toChatSources(result.Sources), nil
		}
	}

	web, webErr := googleai.HTTPSearch(parent, webQuery(question))
	if webErr == nil && strings.TrimSpace(web.Answer) != "" {
		if chromeErr != nil {
			log.Printf("ask eos google: chrome unavailable (%v); used web results", chromeErr)
		}
		return false, web.Answer, toChatSources(web.Sources), nil
	}

	if chromeErr != nil {
		err = chromeErr
	} else if webErr != nil {
		err = webErr
	} else {
		err = fmt.Errorf("empty search results")
	}
	log.Printf("ask eos google: search failed chrome=%v web=%v", chromeErr, webErr)
	return false, "", nil, err
}

func toChatSources(in []googleai.Source) []Source {
	out := make([]Source, 0, len(in))
	for _, s := range in {
		out = append(out, Source{Title: s.Title, URL: s.URL})
	}
	return out
}

func clipSearchText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= maxSearchText {
		return s
	}
	return string([]rune(s)[:maxSearchText]) + "…"
}

const eosQuestionTag = "This question is related to CERN's EOS Storage System."

func searchQuery(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		return q
	}
	low := strings.ToLower(q)
	if strings.Contains(low, "cern's eos storage system") || strings.Contains(low, "cerns eos storage system") {
		return q
	}
	return q + " — " + eosQuestionTag
}

func webQuery(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		return q
	}
	low := strings.ToLower(q)
	if strings.Contains(low, "eosxd") {
		return "eosxd FUSE mount CERN EOS"
	}
	if strings.Contains(low, "quarkdb") || strings.Contains(low, "xrootd") || strings.Contains(low, "cern") {
		return q
	}
	if strings.Contains(low, " eos") || strings.HasPrefix(low, "eos") {
		return q
	}
	return q + " CERN EOS storage"
}
