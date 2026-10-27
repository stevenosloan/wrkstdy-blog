package site

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
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
		got, err := b.renderMarkdown([]byte(tt.src), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, tt.want) {
			t.Errorf("render(%q) = %q, want it to contain %q", tt.src, got, tt.want)
		}
	}
}

func TestAlerts(t *testing.T) {
	b := &builder{md: newMarkdown()}
	tests := []struct{ name, src, want string }{
		{"note", "> [!NOTE]\n> Worth knowing.",
			"<div class=\"alert alert-note\">\n<p class=\"alert-title\">Note</p>\n<p>Worth knowing.</p>\n</div>\n"},
		{"marker on its own paragraph", "> [!WARNING]\n>\n> Careful.\n>\n> Twice.",
			"<div class=\"alert alert-warning\">\n<p class=\"alert-title\">Warning</p>\n<p>Careful.</p>\n<p>Twice.</p>\n</div>\n"},
		{"keeps inline markup", "> [!TIP]\n> Use `go tool mage`, **really**.",
			"<div class=\"alert alert-tip\">\n<p class=\"alert-title\">Tip</p>\n<p>Use <code>go tool mage</code>, <strong>really</strong>.</p>\n</div>\n"},
		{"unknown type stays a quote", "> [!FOO]\n> Text.", "<blockquote>\n<p>[!FOO]\nText.</p>\n</blockquote>\n"},
		{"marker not alone stays a quote", "> [!NOTE] inline\n> Text.", "<blockquote>\n<p>[!NOTE] inline\nText.</p>\n</blockquote>\n"},
		{"plain quote", "> Just a quote.", "<blockquote>\n<p>Just a quote.</p>\n</blockquote>\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.renderMarkdown([]byte(tt.src), nil)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestTodos(t *testing.T) {
	b := &builder{md: newMarkdown()}
	draft := &pageInfo{Draft: true}
	tests := []struct {
		name, src, want string
		page            *pageInfo
	}{
		{"draft", "<!-- TODO(steven): the opening. -->", "<div class=\"alert alert-todo\">\n<p class=\"alert-title\">TODO</p>\n<p>the opening.</p>\n</div>\n", draft},
		{"no name", "<!-- TODO: fix <this> -->", "<div class=\"alert alert-todo\">\n<p class=\"alert-title\">TODO</p>\n<p>fix &lt;this&gt;</p>\n</div>\n", draft},
		{"several lines", "<!-- TODO:\nfirst\nsecond -->", "<div class=\"alert alert-todo\">\n<p class=\"alert-title\">TODO</p>\n<p>first\nsecond</p>\n</div>\n", draft},
		{"other comment", "<!-- just a note -->", "<!-- raw HTML omitted -->\n", draft},
		{"published", "<!-- TODO: the opening. -->", "<!-- raw HTML omitted -->\n", &pageInfo{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.renderMarkdown([]byte(tt.src), tt.page)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
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

func TestFingerprintDirSkipsWhatEsbuildBuilds(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	for _, rel := range []string{"css/main.css", "js/main.js", "fonts/a.woff2", "icons/a.svg"} {
		if err := writeFile(filepath.Join(src, rel), []byte("x")); err != nil {
			t.Fatal(err)
		}
	}

	assets, err := fingerprintDir(src, out, "assets")
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets["icons/a.svg"].URL == "" {
		t.Errorf("assets = %v, want only icons/a.svg", assets)
	}
}

func TestBundleAssetsCSS(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	files := map[string]string{
		"assets/fonts/a.woff2": "font",
		"assets/css/main.css":  "/* comment */\n@font-face {\n  src: url(\"../fonts/a.woff2\");\n}\nbody {\n  color: red;\n}\n",
	}
	for name, body := range files {
		if err := writeFile(filepath.Join(root, filepath.FromSlash(name)), []byte(body)); err != nil {
			t.Fatal(err)
		}
	}

	assets, err := NewBundler(root, out, false).Build()
	if err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(assets["css/main.css"].URL)))
	if err != nil {
		t.Fatal(err)
	}
	font := assets["fonts/a.woff2"]
	if font.URL == "" {
		t.Fatalf("assets = %v, want one for fonts/a.woff2", assets)
	}
	if !strings.Contains(string(css), font.URL) || strings.Contains(string(css), "comment") {
		t.Errorf("css = %q, want it minified and pointing at %q", css, font.URL)
	}
	if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(font.URL))); err != nil {
		t.Error(err)
	}
	sum := sha256.Sum256(css)
	if integrity := "sha256-" + base64.StdEncoding.EncodeToString(sum[:]); assets["css/main.css"].Integrity != integrity {
		t.Errorf("integrity = %q, want the hash of the minified file, %q", assets["css/main.css"].Integrity, integrity)
	}

	if err := writeFile(filepath.Join(root, "assets/css/main.css"), []byte(`@font-face { src: url("../fonts/missing.woff2"); }`)); err != nil {
		t.Fatal(err)
	}
	if _, err := NewBundler(root, t.TempDir(), false).Build(); err == nil {
		t.Error("want an error for a url() to a missing font")
	}
}

func TestBundleAssetsJS(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	files := map[string]string{
		"assets/js/main.js":             "import { greet } from 'pkg';\ngreet('hi');\nimport('./lazy.js').then((m) => m.default());\n",
		"assets/js/lazy.js":             "export default () => console.log('lazy');\n",
		"node_modules/pkg/package.json": `{"name": "pkg", "main": "index.js"}`,
		"node_modules/pkg/index.js":     "export const greet = (message) => console.log(message);\n",
	}
	for name, body := range files {
		if err := writeFile(filepath.Join(root, filepath.FromSlash(name)), []byte(body)); err != nil {
			t.Fatal(err)
		}
	}

	assets, err := NewBundler(root, out, false).Build()
	if err != nil {
		t.Fatal(err)
	}
	// lazy.js is also a file in assets/js, so it's an entry point of its own.
	main, ok := assets["js/main.js"]
	if !ok {
		t.Fatalf("assets = %v, want one for js/main.js", assets)
	}
	got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(main.URL)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "console.log(") || strings.Contains(string(got), "from\"pkg\"") {
		t.Errorf("main.js = %q, want pkg bundled in", got)
	}
	sum := sha256.Sum256(got)
	if integrity := "sha256-" + base64.StdEncoding.EncodeToString(sum[:]); main.Integrity != integrity {
		t.Errorf("integrity = %q, want %q", main.Integrity, integrity)
	}
}

func TestBundlerDev(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := writeFile(filepath.Join(root, filepath.FromSlash(name)), []byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	write("assets/js/main.js", "console.log('one');\n")
	write("node_modules/.keep", "")

	b := NewBundler(root, out, true)
	defer b.Close()
	assets, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if got := assets["js/main.js"].URL; got != "/assets/js/main.js" {
		t.Fatalf("URL = %q, want a stable name", got)
	}
	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if got := read("assets/js/main.js"); !strings.Contains(got, `console.log("one");`) || !strings.Contains(got, "sourceMappingURL=/assets/js/main.js.map") {
		t.Errorf("main.js = %q, want it unminified and linked to a source map", got)
	}
	read("assets/js/main.js.map")

	// A rebuild sees edits, and an entry point added since the first build.
	write("assets/js/main.js", "console.log('two');\n")
	write("assets/js/extra.js", "console.log('extra');\n")
	assets, err = b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if got := read("assets/js/main.js"); !strings.Contains(got, "two") {
		t.Errorf("main.js = %q, want the edit", got)
	}
	if _, ok := assets["js/extra.js"]; !ok {
		t.Errorf("assets = %v, want one for the new js/extra.js", assets)
	}
}

func TestBundleAssetsJSNeedsNodeModules(t *testing.T) {
	root := t.TempDir()
	if err := writeFile(filepath.Join(root, "assets/js/main.js"), []byte("console.log(1)")); err != nil {
		t.Fatal(err)
	}
	if _, err := NewBundler(root, t.TempDir(), false).Build(); err == nil || !strings.Contains(err.Error(), "npm") {
		t.Errorf("err = %v, want one that says to run npm", err)
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

func TestBuildWritesArchiveAndFeed(t *testing.T) {
	out := t.TempDir()
	if _, err := Build(Options{Root: "../..", Out: out}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "blog", "index.html")); err != nil {
		t.Errorf("want an archive page at /blog/: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(out, "blog", "feed.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var feed rss
	if err := xml.Unmarshal(data, &feed); err != nil {
		t.Fatalf("feed.xml doesn't parse: %v", err)
	}
	if len(feed.Channel.Items) == 0 {
		t.Fatal("feed has no items")
	}
	for _, item := range feed.Channel.Items {
		if !strings.HasPrefix(item.Link, "https://") {
			t.Errorf("item link %q should be absolute", item.Link)
		}
	}
	// Checked on the raw XML: Unmarshal can't fill content:encoded, since it
	// matches the tag against the namespace URL rather than the prefix.
	if !strings.Contains(string(data), "<content:encoded>") {
		t.Error("feed items should carry their full content")
	}
	if strings.Contains(string(data), `href=&#34;/`) {
		t.Error("feed content has a site-relative link")
	}
}
