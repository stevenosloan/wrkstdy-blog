# wrkstdy.com

The homepage and blog for wrkstdy.com, built by a small Go program in this repo and deployed on Netlify.

## Run it locally

```sh
brew install go                    # Go 1.24 or newer; Netlify uses the version in netlify.toml
go tool mage                       # list tasks
go tool mage serve                 # dev server on http://localhost:1313, rebuilds on save
go run ./cmd/wrkstdy serve -drafts # include posts with draft: true
go tool mage build                 # write the site to public/
go tool mage test                  # go vet + tests
go tool mage check                 # test, then build (what Netlify runs)
```

Tasks live in `magefile.go`, and Mage is pinned in `go.mod` as a Go tool, so there's nothing to install.

## Write a post

```sh
go run ./cmd/wrkstdy new blog/my-post
```

That creates `content/blog/my-post.md` as a draft. Remove `draft: true` to publish it.

- Fenced code blocks with a language (` ```go `) are highlighted in the browser by
  [MicroLighter](https://github.com/davatron5000/microlighter), vendored in `static/vendor/microlighter/`.
- ` ```mermaid ` blocks render as diagrams. Mermaid loads from jsDelivr only on pages that use it.
- `{{< name >}}` on its own line drops in `templates/shortcodes/name.html`.

## Styleguide

`/styleguide/` (from `content/styleguide.md`) shows the palette, type, and every styled element. Check it after changing `main.css`.

## How the build works

`site.Build` in `internal/site/build.go` runs these steps in order:

1. Read `site.toml` and `data/projects.toml`.
2. Parse `templates/`: `base.html` and `partials/` once, then one copy per page kind (`home`, `single`), each defining its own `main` block.
3. Walk `content/`. Each `.md` file has YAML front matter split off and its body rendered by [goldmark](https://github.com/yuin/goldmark). `_index.md` makes a folder a section (posts in `content/blog/` are served under `/blog/`).
4. Wipe `public/`, copy `static/` into it, and write each file in `assets/` under a content-hashed name.
5. Render every page.

## Layout

| Path | What it is |
| --- | --- |
| `cmd/wrkstdy/` | The command: `build`, `serve`, `new` |
| `magefile.go` | Dev tasks (serve, build, test, check) run with `go tool mage` |
| `internal/site/` | The generator: content loading, markdown, templates, assets |
| `site.toml` | Site title, base URL, and menu |
| `content/_index.md` | Homepage intro text |
| `content/blog/` | Blog posts, served under `/blog/` |
| `data/projects.toml` | Project links listed on the homepage |
| `templates/` | Page templates (Go `html/template`) |
| `assets/` | CSS and JS, fingerprinted at build time |
| `static/` | Copied as-is: fonts and MicroLighter |
| `netlify.toml` | Netlify build settings and pinned Go version |
