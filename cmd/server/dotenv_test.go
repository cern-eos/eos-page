package main

import (
	"os"
	"testing"
)

func TestApplyDotEnvFillsEmpty(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	n := applyDotEnv([]byte("OPENAI_API_KEY=sk-test-key\nOPENAI_MODEL=gpt-4o-mini\n"))
	if n != 2 {
		t.Fatalf("applied %d", n)
	}
	if os.Getenv("OPENAI_API_KEY") != "sk-test-key" {
		t.Fatalf("key not applied")
	}
	if os.Getenv("OPENAI_MODEL") != "gpt-4o-mini" {
		t.Fatalf("model not applied")
	}
}

func TestApplyDotEnvDoesNotOverrideSet(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "already")
	applyDotEnv([]byte("OPENAI_API_KEY=sk-other\n"))
	if os.Getenv("OPENAI_API_KEY") != "already" {
		t.Fatalf("should keep existing key")
	}
}

func TestApplyDotEnvExportAndQuotes(t *testing.T) {
	t.Setenv("FOO_DOTENV", "")
	applyDotEnv([]byte("export FOO_DOTENV=\"bar baz\"\n# comment\n"))
	if os.Getenv("FOO_DOTENV") != "bar baz" {
		t.Fatalf("got %q", os.Getenv("FOO_DOTENV"))
	}
}
