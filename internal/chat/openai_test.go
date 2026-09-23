package chat

import (
	"encoding/json"
	"testing"
)

func TestParseResponsesOutput(t *testing.T) {
	var raw responsesResp
	if err := json.Unmarshal([]byte(`{
		"output": [
			{
				"type": "web_search_call",
				"status": "completed",
				"action": {
					"type": "search",
					"query": "CERN EOS storage volume",
					"sources": [{"url": "https://home.cern/news/eos", "title": "EOS at CERN"}]
				}
			},
			{
				"type": "message",
				"role": "assistant",
				"content": [
					{
						"type": "output_text",
						"text": "EOS stores about 1.1 EB at CERN.",
						"annotations": [
							{"type": "url_citation", "url": "https://eos.web.cern.ch/", "title": "EOS Open Storage"}
						]
					}
				]
			}
		]
	}`), &raw); err != nil {
		t.Fatal(err)
	}
	text, sources, calls := parseResponsesOutput(raw)
	if text != "EOS stores about 1.1 EB at CERN." {
		t.Fatalf("text %q", text)
	}
	if len(calls) != 0 {
		t.Fatalf("calls %d", len(calls))
	}
	if len(sources) != 2 {
		t.Fatalf("sources %d", len(sources))
	}
}

func TestIsGPT5(t *testing.T) {
	if !isGPT5("gpt-5-mini") || isGPT5("gpt-4o-mini") {
		t.Fatal("isGPT5 mismatch")
	}
}
