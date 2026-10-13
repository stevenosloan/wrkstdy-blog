package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	path, err := NewPost(root, "blog", "Hello, world")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "content", "blog", "hello-world.md"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `title: "Hello, world"`) {
		t.Errorf("front matter missing title:\n%s", body)
	}
	if _, err := NewPost(root, "blog", "Hello world"); err == nil {
		t.Error("second post with the same slug should fail, not overwrite")
	}
}
