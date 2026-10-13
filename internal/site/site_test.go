package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestSplitDatePrefix(t *testing.T) {
	tests := []struct {
		name, date, slug string
	}{
		{"2026-10-12-the-long-way", "2026-10-12", "the-long-way"},
		{"the-long-way", "", "the-long-way"},
		{"2026-10-12", "", "2026-10-12"},
		{"2026-13-40-bad-date", "", "2026-13-40-bad-date"},
		{"2026-10-12x-not-a-prefix", "", "2026-10-12x-not-a-prefix"},
	}
	for _, tt := range tests {
		date, slug := splitDatePrefix(tt.name)
		var got string
		if !date.IsZero() {
			got = date.Format(time.DateOnly)
		}
		if got != tt.date || slug != tt.slug {
			t.Errorf("splitDatePrefix(%q) = %q, %q; want %q, %q", tt.name, got, slug, tt.date, tt.slug)
		}
	}
}

func TestFingerprintDirRewritesCSSURLs(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	write := func(rel, data string) {
		if err := writeFile(filepath.Join(src, rel), []byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	write("fonts/a.woff2", "font")
	write("css/main.css", `@font-face { src: url("/assets/fonts/a.woff2"); }`)

	assets, err := fingerprintDir(src, out, "assets")
	if err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(assets["css/main.css"].URL)))
	if err != nil {
		t.Fatal(err)
	}
	if want := `url("` + assets["fonts/a.woff2"].URL + `")`; !strings.Contains(string(css), want) {
		t.Errorf("css = %q, want it to contain %q", css, want)
	}

	write("css/main.css", `@font-face { src: url("/assets/fonts/missing.woff2"); }`)
	if _, err := fingerprintDir(src, t.TempDir(), "assets"); err == nil {
		t.Error("want an error for a url() to a missing asset")
	}
}

func TestBuildWritesNotFoundPage(t *testing.T) {
	out := t.TempDir()
	if _, err := Build(Options{Root: "../..", Out: out}); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(out, "404.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), `<meta name="robots" content="noindex">`) {
		t.Error("404.html should be noindex")
	}
	if strings.Contains(string(html), `rel="canonical"`) {
		t.Error("404.html shouldn't claim a canonical URL")
	}
	if _, err := os.Stat(filepath.Join(out, "404", "index.html")); err == nil {
		t.Error("404.md should only be written to /404.html, not /404/")
	}
}
