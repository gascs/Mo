package service

import (
	"strings"
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"english", "Hello World Post", "hello-world-post"},
		{"with punctuation", "What's New in Go?", "what-s-new-in-go"},
		{"empty title", "", ""},
		{"chinese only", "我的博客", ""},
		{"mixed", "Hello 世界", "hello-世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.title)
			if tt.want == "" {
				if got == "" {
					t.Error("expected non-empty slug")
				}
				if !strings.HasPrefix(got, "post-") {
					t.Errorf("expected post- prefix for fallback slug, got: %s", got)
				}
			} else if got != tt.want {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		maxChars int
		want     string
	}{
		{"plain text", "<p>Hello World</p>", 200, "Hello World"},
		{"strip tags", "<h1>Title</h1><p>Content here.</p>", 200, "TitleContent here."},
		{"truncate", "<p>This is a long text that should be truncated</p>", 10, "This is a ..."},
		{"empty", "", 200, ""},
		{"unicode", "<p>你好世界</p>", 3, "你好世..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSummary(tt.html, tt.maxChars)
			if got != tt.want {
				t.Errorf("ExtractSummary(%q, %d) = %q, want %q", tt.html, tt.maxChars, got, tt.want)
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		contain string
	}{
		{"heading", "# Hello", "<h1"},
		{"paragraph", "Hello world.", "<p>"},
		{"code block", "```go\nfmt.Println(1)\n```", "<pre"},
		{"bold", "**bold**", "<strong>"},
		{"link", "[text](url)", "<a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderMarkdown(tt.input)
			if err != nil {
				t.Fatalf("RenderMarkdown error: %v", err)
			}
			if !strings.Contains(got, tt.contain) {
				t.Errorf("RenderMarkdown(%q) = %q, expected to contain %q", tt.input, got, tt.contain)
			}
		})
	}
}
