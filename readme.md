# wrkstdy.com

The homepage and blog for wrkstdy.com, built by a small Go program in this repo and deployed on Netlify.

[![Netlify Status](https://api.netlify.com/api/v1/badges/4162cdb1-453b-4d6e-bdca-660630b8a539/deploy-status)](https://app.netlify.com/projects/wrkstdy-blog/deploys)

## Run it locally

```sh
brew install go node               # Go 1.27 or newer (Netlify uses the version in netlify.toml), and Node for the JS dependencies
npm install                        # mermaid and microlighter, which esbuild bundles into the site
go tool mage                       # list tasks
go tool mage serve                 # dev server on http://localhost:1313, rebuilds on save (JS and CSS unminified, with source maps)
go tool mage serve -drafts         # include posts with draft: true
go tool mage build                 # write the site to public/
go tool mage build -production     # test first, then build (what Netlify runs on deploy)
go tool mage build -drafts         # include posts with draft: true (what PR deploy previews run)
go tool mage test                  # go vet + tests
go tool mage new                   # prompt for a title and start a draft post
go tool mage new -section=journal  # same, as a folder for photos (see below)
go tool mage new -section=recipes
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
  [MicroLighter](https://github.com/davatron5000/microlighter), bundled from npm. Each language's grammar is its own chunk, fetched when a page needs it.
- ` ```mermaid ` blocks render as diagrams. Mermaid is bundled from npm and loads only on pages that use it.
- `{{< name >}}` on its own line drops in `templates/shortcodes/name.html`.
- ` ```recipe ` blocks render as recipe tables (see below).

## Journal entries and photos

`go tool mage new -section=journal` makes a folder instead of a file: `content/journal/2026-10-15-lake-day/index.md`. That's a page bundle. The folder's name works like a post's filename (date, then slug, so it's served at `/journal/lake-day/`), and its photos go next to `index.md`. Any section can have bundles; pass `-bundle` to make one elsewhere, or `-bundle=false` for a plain journal post.

Refer to a photo by its filename, and write its note on the line right under it. With no blank line between, the note becomes the photo's caption:

```markdown
![Rowboat tied to the dock](dock.jpg)
First morning: fog didn't lift until *ten*.
```

JPEGs and PNGs are resized to 640, 960, 1280 and 1920px wide (skipping any wider than the photo), each width encoded as AVIF (quality 50), WebP (quality 70) and a JPEG fallback (quality 80; PNGs stay PNG). They're served in a `<picture>`, so the browser takes the first format it supports, with `srcset`, `sizes="auto"` and `loading="lazy"`. The encoders ([gen2brain/avif](https://github.com/gen2brain/avif), [gen2brain/webp](https://github.com/gen2brain/webp)) are libavif and libwebp compiled to WebAssembly, so the build needs no C toolchain; they use a system copy instead when one is installed. Other formats are published as they are. Only photos a page refers to are published, and never the originals, so their EXIF data (including where a photo was taken) stays out of the site. The rotation phone cameras store in EXIF is applied before resizing.

Resizing is slow (about 9 seconds for a 4K photo, mostly AVIF), so the results are kept in `.cache/images/` (not committed) and redone only when a photo changes. Netlify starts without that cache, so a deploy resizes every photo again. Exporting photos at around 2000px rather than straight off the camera keeps that quick. iPhone HEIC photos need converting to JPEG first.

## Recipes

A ` ```recipe ` block is the recipe as a table, in [Michael Chu's notation](https://www.cookingforengineers.com/): ingredients down the left, each step to the right of what goes into it, narrowing to the finished dish. Write it as an outline with the dish first. Indented under each step are the ingredients and earlier steps it uses, and a line with nothing under it is an ingredient:

```recipe
serve with rice
  stir together, simmer 5 min
    simmer until soft, 20 min
      1 cup red lentils, rinsed
      3 cups water
    fry until the seeds pop
      2 tbsp ghee
      1 tsp cumin seeds
```

Write the talk-through around it as normal markdown.

## Styleguide

`/blog/styleguide/` (from `content/blog/2026-10-11-styleguide.md`) shows the palette, type, and every styled element. Check it after changing `main.css`. It's a draft, so run `go tool mage serve -drafts` to see it.

## How the build works

`site.Build` in `internal/site/build.go` runs these steps in order:

1. Read `site.toml` and `data/projects.toml`.
2. Parse `templates/`: `base.html` and `partials/` once, then one copy per page kind (`home`, `archive`, `single`), each defining its own `main` block.
3. Walk `content/`. Each `.md` file has YAML front matter split off and its body rendered by [goldmark](https://github.com/yuin/goldmark). `_index.md` makes a folder a section (posts in `content/blog/` are served under `/blog/`). A leading `YYYY-MM-DD-` in a filename becomes the page's date and is dropped from its URL. A folder holding `index.md` is a page bundle, named by the folder; the photos it refers to are resized (`internal/site/images.go`).
4. Wipe `public/`, copy `static/` and the resized photos into it, and write each file in `assets/` under a content-hashed name in `public/assets/`. [esbuild](https://esbuild.github.io/) (`internal/site/bundle.go`) builds the CSS and JS: each file in `assets/css/` and `assets/js/` is an entry point, minified, with JS bundled with what it imports from `node_modules/` and code-split. `serve` keeps one esbuild context between rebuilds and writes unminified files with source maps and unhashed names, so devtools breakpoints survive a rebuild. Fonts referenced by `url()` in the CSS are hashed by esbuild and the `url()` rewritten. Icons and anything else are hashed as they are.
5. Render every page, including each section's `_index.md` as an archive of its posts grouped by year (`/blog/`), and write each section's posts as an RSS feed (`/blog/feed.xml`, in `internal/site/feed.go`). The feed carries each post's full HTML, with site-relative links made absolute so they work in a feed reader.

## Layout

| Path | What it is |
| --- | --- |
| `magefile.go` | Every task (serve, build, test, new), run with `go tool mage` |
| `internal/site/` | The generator: content loading, markdown, templates, assets, new posts |
| `site.toml` | Site title, base URL, and menu |
| `content/_index.md` | Homepage intro text |
| `content/blog/` | Blog posts, served under `/blog/`, with an archive at `/blog/` and an RSS feed at `/blog/feed.xml` |
| `content/journal/` | Trip entries, each a folder with its photos, served under `/journal/` with its own archive and feed |
| `content/recipes/` | Recipes, served under `/recipes/` with its own archive and feed |
| `content/404.md` | The not-found page, written to `/404.html`, which Netlify serves for any missing path |
| `data/projects.toml` | Project links listed on the homepage |
| `templates/` | Page templates (Go `html/template`) |
| `assets/` | CSS, JS, fonts, and icons, fingerprinted at build time. esbuild builds `css/` and `js/` (and the fonts the CSS uses); the rest is hashed as is |
| `scripts/fonts.sh` | Rebuilds the variable fonts in `assets/fonts/` (needs `fonttools`); the output is committed |
| `package.json` | Browser JS dependencies (mermaid, microlighter), pinned by `package-lock.json` |
| `static/` | Copied as-is: `favicon.ico`, for clients that ask for it at the root |
| `netlify.toml` | Netlify build settings (`npm ci`, then the build) and pinned Go version |
| `.github/workflows/test.yml` | Runs `mage test` and `mage build -drafts` on pushes to main and on pull requests |
