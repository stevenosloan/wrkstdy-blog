package site

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"
)

var funcs = template.FuncMap{
	"date":  func(t time.Time) string { return t.Format("2006-01-02") },
	"lower": strings.ToLower,
	// list and dict build literals inside templates, which have no syntax for them.
	"list": func(xs ...any) []any { return xs },
	"dict": func(kv ...any) (map[string]any, error) {
		if len(kv)%2 != 0 {
			return nil, fmt.Errorf("dict needs key/value pairs")
		}
		m := make(map[string]any, len(kv)/2)
		for i := 0; i < len(kv); i += 2 {
			m[fmt.Sprint(kv[i])] = kv[i+1]
		}
		return m, nil
	},
	// active reports whether a menu link should be marked as the current section.
	"active": func(pageURL, menuURL string) bool {
		return menuURL != "/" && strings.HasPrefix(pageURL, menuURL)
	},
	"tagURL": tagURL,
}

var pageKinds = []string{"home", "list", "single", "tags", "tag"}

// loadTemplates parses base.html and partials/ once, then clones that set for
// each page kind so every kind can define its own "main" block.
func (b *builder) loadTemplates(dir string) error {
	base, err := template.New("").Funcs(funcs).ParseFiles(filepath.Join(dir, "base.html"))
	if err != nil {
		return err
	}
	if base, err = base.ParseGlob(filepath.Join(dir, "partials", "*.html")); err != nil {
		return err
	}

	b.pages = map[string]*template.Template{}
	for _, kind := range pageKinds {
		t, err := base.Clone()
		if err != nil {
			return err
		}
		if t, err = t.ParseFiles(filepath.Join(dir, kind+".html")); err != nil {
			return err
		}
		b.pages[kind] = t
	}

	b.shortcodes, err = template.New("").Funcs(funcs).ParseGlob(filepath.Join(dir, "shortcodes", "*.html"))
	return err
}
