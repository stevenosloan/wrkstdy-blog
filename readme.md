# wrkstdy.com

The homepage and blog for wrkstdy.com, built by a small Go program in this repo and deployed on Netlify.

## Run it locally

```sh
brew install go                    # Go 1.24 or newer; Netlify uses the version in netlify.toml
go run ./cmd/wrkstdy serve         # http://localhost:1313, rebuilds when files change
go run ./cmd/wrkstdy serve -drafts # include posts with draft: true
go run ./cmd/wrkstdy build         # write the site to public/
go test ./...
```

## Write a post

```sh
go run ./cmd/wrkstdy new blog/my-post
```

That creates `content/blog/my-post.md` as a draft. Add `tags` in the front matter (for example `dev-log`) and remove `draft: true` to publish it.

- Fenced code blocks with a language (` ```go `) are highlighted in the browser by
  [MicroLighter](https://github.com/davatron5000/microlighter), vendored in `static/vendor/microlighter/`.
- ` ```mermaid ` blocks render as diagrams. Mermaid loads from jsDelivr only on pages that use it.
- `{{< name >}}` on its own line drops in `templates/shortcodes/name.html`.

## Styleguide

`/styleguide/` (from `content/styleguide.md`) shows the palette, type, and every styled element. Check it after changing `main.css`.

## How the build works

`site.Build` in `internal/site/build.go` runs these steps in order:

1. Read `site.toml` and `data/projects.toml`.
2. Parse `templates/`: `base.html` and `partials/` once, then one copy per page kind (`home`, `list`, `single`, `tags`, `tag`), each defining its own `main` block.
3. Walk `content/`. Each `.md` file has YAML front matter split off and its body rendered by [goldmark](https://github.com/yuin/goldmark). `_index.md` makes a folder a section (`content/blog/` becomes `/blog/`).
4. Wipe `public/`, copy `static/` into it, and write each file in `assets/` under a content-hashed name.
5. Render every page, the tag pages, and `/blog/index.xml`.

## Layout

| Path | What it is |
| --- | --- |
| `cmd/wrkstdy/` | The command: `build`, `serve`, `new` |
| `internal/site/` | The generator: content loading, markdown, templates, assets, RSS |
| `site.toml` | Site title, base URL, and menu |
| `content/_index.md` | Homepage intro text |
| `content/blog/` | Blog posts, served at `/blog/` |
| `data/projects.toml` | Project links listed on the homepage |
| `templates/` | Page templates (Go `html/template`) |
| `assets/` | CSS and JS, fingerprinted at build time |
| `static/` | Copied as-is: fonts and MicroLighter |
| `netlify.toml` | Netlify build settings and pinned Go version |
