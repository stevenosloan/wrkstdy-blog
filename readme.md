# wrkstdy.com

The homepage and blog for wrkstdy.com, built with [Hugo](https://gohugo.io) and deployed on Netlify.

## Run it locally

```sh
brew install hugo      # needs Hugo 0.146 or newer; Netlify uses the version in netlify.toml
hugo server            # http://localhost:1313
```

## Write a post

```sh
hugo new content blog/my-post.md
```

Add `tags` in the front matter (for example `dev-log`). Posts with `draft: true` only show up with `hugo server -D`.

- Fenced code blocks with a language (` ```go `) are highlighted in the browser by
  [MicroLighter](https://github.com/davatron5000/microlighter), vendored in `static/vendor/microlighter/`.
- ` ```mermaid ` blocks render as diagrams. Mermaid loads from jsDelivr only on pages that use it.

## Styleguide

`/styleguide/` (from `content/styleguide.md`) shows the palette, type, and every styled element. Check it after changing `main.css`.

## Layout

| Path | What it is |
| --- | --- |
| `content/_index.md` | Homepage intro text |
| `content/blog/` | Blog posts, served at `/blog/` |
| `data/projects.toml` | Project links listed on the homepage |
| `layouts/` | Page templates (no third-party theme) |
| `layouts/_markup/` | Code block render hooks (plain markup for MicroLighter, Mermaid) |
| `assets/css/main.css` | All styles, including the forest palette and syntax theme |
| `static/fonts/` | Fira Sans and Fira Code (OFL) |
| `netlify.toml` | Netlify build settings and pinned Hugo version |
