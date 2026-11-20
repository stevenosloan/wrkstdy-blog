package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Hello world": "hello-world",
		"Building a blog the long way, on purpose": "building-a-blog-the-long-way-on-purpose",
		"  I'm learning Go!  ":                     "im-learning-go",
		"Go 1.25 — what’s new?":                    "go-1-25-whats-new",
		"???":                                      "",
	}
	for title, want := range tests {
		if got := Slugify(title); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestNewPost(t *testing.T) {
	root := t.TempDir()
	date := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	path, err := NewPost(root, "blog", "Hello, world", date)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "content", "blog", "2026-10-12-hello-world", "index.md"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `title: "Hello, world"`) {
		t.Errorf("front matter missing title:\n%s", body)
	}
	if _, err := NewPost(root, "blog", "Hello world", date); err == nil {
		t.Error("second post with the same slug should fail, not overwrite")
	}

	// A plain-file post with the same name would publish to the same URL.
	if err := writeFile(filepath.Join(root, "content", "blog", "2026-10-13-old.md"), []byte("old")); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPost(root, "blog", "Old", date.AddDate(0, 0, 1)); err == nil {
		t.Error("a folder post beside a plain file of the same name should fail")
	}
}

func TestNewPostStarterBodies(t *testing.T) {
	root := t.TempDir()
	date := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	tests := map[string]string{
		"recipes": "```recipe\n",
		"journal": "go tool mage photos",
	}
	for section, want := range tests {
		path, err := NewPost(root, section, "Hello", date)
		if err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), want) {
			t.Errorf("%s post missing %q:\n%s", section, want, body)
		}
	}

	path, err := NewPost(root, "blog", "Hello", date)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(body), "---\n\n") {
		t.Errorf("blog post should be front matter only:\n%s", body)
	}
}

// A new recipe has to build as it is, and an empty recipe block doesn't.
func TestRecipeStarterParses(t *testing.T) {
	body := starterBodies["recipes"]
	start := strings.Index(body, "```recipe\n") + len("```recipe\n")
	end := strings.LastIndex(body, "```")
	if _, err := renderRecipe(body[start:end]); err != nil {
		t.Errorf("starter recipe doesn't render: %v", err)
	}
}
