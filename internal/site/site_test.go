package site

import (
	"strings"
	"testing"
)

func TestSplitFrontMatter(t *testing.T) {
	tests := []struct {
		name, src, meta, body string
		wantErr               bool
	}{
		{name: "front matter", src: "---\ntitle: Hi\n---\nBody\n", meta: "title: Hi", body: "Body\n"},
		{name: "windows line endings", src: "---\r\ntitle: Hi\r\n---\r\nBody\r\n", meta: "title: Hi", body: "Body\n"},
		{name: "no front matter", src: "Just text\n", body: "Just text\n"},
		{name: "unclosed", src: "---\ntitle: Hi\nBody\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, body, err := splitFrontMatter([]byte(tt.src))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if string(meta) != tt.meta || string(body) != tt.body {
				t.Errorf("got meta %q body %q, want %q %q", meta, body, tt.meta, tt.body)
			}
		})
	}
}

func TestCodeBlocks(t *testing.T) {
	b := &builder{md: newMarkdown()}
	tests := []struct{ src, want string }{
		{"```go\nx := 1 < 2\n```", `<pre><code class="language-go">x := 1 &lt; 2`},
		{"```diff\n- a\n```", `<pre><code class="language-git-diff">`},
		{"```\nplain\n```", "<pre><code>plain"},
		{"```mermaid\na --> b\n```", `<pre class="mermaid">a --&gt; b`},
	}
	for _, tt := range tests {
		got, err := b.renderMarkdown([]byte(tt.src))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, tt.want) {
			t.Errorf("render(%q) = %q, want it to contain %q", tt.src, got, tt.want)
		}
	}
}
