# wrkstdy.com

The homepage and blog for wrkstdy.com, built by a small Go program in this repo and deployed on Netlify.

[![Netlify Status](https://api.netlify.com/api/v1/badges/4162cdb1-453b-4d6e-bdca-660630b8a539/deploy-status)](https://app.netlify.com/projects/wrkstdy-blog/deploys)

## Run it locally

```sh
brew install go                    # Go 1.24 or newer; Netlify uses the version in netlify.toml
go tool mage                       # list tasks
go tool mage serve                 # dev server on http://localhost:1313, rebuilds on save
go tool mage serve -drafts         # include posts with draft: true
go tool mage build                 # write the site to public/
go tool mage build -production     # test first, then build (what Netlify runs on deploy)
go tool mage build -drafts         # include posts with draft: true (what PR deploy previews run)
go tool mage test                  # go vet + tests
go tool mage new                   # prompt for a title and start a draft post
```

Tasks live in `magefile.go`, and Mage is pinned in `go.mod` as a Go tool, so there's nothing to install.

## Write a post

```sh
go tool mage new
# Title: Building a blog, on purpose
# Date [2026-10-13]:              (enter for today, or type YYYY-MM-DD)
```

That creates `content/blog/2026-10-13-building-a-blog-on-purpose.md` as a draft, named with the date and a slug made from the title. Remove `draft: true` to publish it.

The date in the filename is the post's date, and it's left out of the URL, so that post is served at `/blog/building-a-blog-on-purpose/`. To re-date a post, rename the file, or set `date:` in its front matter, which wins over the filename.

Pass `-section` to put it somewhere other than `content/blog/`, e.g. `go tool mage new -section=notes`.

It won't overwrite an existing post, and the build fails if two posts end up with the same slug (say, the same title on different days). Change the title, or set `slug:` in the front matter.

- Fenced code blocks with a language (` ```go `) are highlighted in the browser by
  [MicroLighter](https://github.com/davatron5000/microlighter), vendored in `static/vendor/microlighter/`.
- ` ```mermaid ` blocks render as diagrams. Mermaid loads from jsDelivr only on pages that use it.
- `{{< name >}}` on its own line drops in `templates/shortcodes/name.html`.

## Styleguide

`/blog/styleguide/` (from `content/blog/2026-10-11-styleguide.md`) shows the palette, type, and every styled element. Check it after changing `main.css`. It's a draft, so run `go tool mage serve -drafts` to see it.

## How the build works

`site.Build` in `internal/site/build.go` runs these steps in order:

1. Read `site.toml` and `data/projects.toml`.
2. Parse `templates/`: `base.html` and `partials/` once, then one copy per page kind (`home`, `single`), each defining its own `main` block.
3. Walk `content/`. Each `.md` file has YAML front matter split off and its body rendered by [goldmark](https://github.com/yuin/goldmark). `_index.md` makes a folder a section (posts in `content/blog/` are served under `/blog/`). A leading `YYYY-MM-DD-` in a filename becomes the page's date and is dropped from its URL.
4. Wipe `public/`, copy `static/` into it, and write each file in `assets/` under a content-hashed name in `public/assets/`.
5. Render every page.

## Layout

| Path | What it is |
| --- | --- |
| `magefile.go` | Every task (serve, build, test, new), run with `go tool mage` |
| `internal/site/` | The generator: content loading, markdown, templates, assets, new posts |
| `site.toml` | Site title, base URL, and menu |
| `content/_index.md` | Homepage intro text |
| `content/blog/` | Blog posts, served under `/blog/` |
| `data/projects.toml` | Project links listed on the homepage |
| `templates/` | Page templates (Go `html/template`) |
| `assets/` | CSS, JS, and fonts, fingerprinted at build time |
| `static/` | Copied as-is: MicroLighter |
| `netlify.toml` | Netlify build settings and pinned Go version |
| `.github/workflows/test.yml` | Runs `mage test` and `mage build -drafts` on pushes to main and on pull requests |
