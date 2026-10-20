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
	path, err := NewPost(root, "blog", "Hello, world", date, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "content", "blog", "2026-10-12-hello-world.md"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `title: "Hello, world"`) {
		t.Errorf("front matter missing title:\n%s", body)
	}
	if _, err := NewPost(root, "blog", "Hello world", date, false); err == nil {
		t.Error("second post with the same slug should fail, not overwrite")
	}
}

func TestNewPostBundle(t *testing.T) {
	root := t.TempDir()
	date := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	path, err := NewPost(root, "journal", "Lake day", date, true)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "content", "journal", "2026-10-15-lake-day", "index.md"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if _, err := NewPost(root, "journal", "Lake day", date, false); err == nil {
		t.Error("a plain post beside a bundle of the same name should fail, since both publish to one URL")
	}
}
