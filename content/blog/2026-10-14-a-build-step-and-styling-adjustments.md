---
title: "A build step and styling adjustments"
description: "Adding an esbuild step to the Go generator, clearing out what was left of the Hugo skeleton, and a round of styling: a type scale, variable fonts, and GitHub-style alerts."
draft: false
---

The [production readiness pass](/blog/a-quick-production-readiness-pass/) ended with a list of things left to do, and one of them was a real build step for CSS and JavaScript. That's where today started. Once that was in, I finally got to the part I'd been putting off: the styles.

## A build step with esbuild

Up to now, assets went through a small pipeline of my own: hash each file, add an integrity attribute, minify the CSS and JS. That covered cache busting, but Mermaid still loaded from jsDelivr and MicroLighter was a vendored copy checked into `static/`.

[esbuild](https://esbuild.github.io) has a Go API, which made it an easy fit. The generator calls it directly, so there's no Node tool in the loop --- npm only installs the dependencies. Every stylesheet in `assets/css` and every script in `assets/js` is an entry point:

- **Mermaid and MicroLighter come from npm.** Both are code-split and content-hashed, so a page only fetches the chunks it uses. Each MicroLighter grammar is its own chunk.
- **Fonts go through esbuild too.** It hashes each font the CSS references and rewrites the `url()`, so the preload tags and the stylesheet always point at the same file.
- **The old minifier is gone.** esbuild minifies, so [tdewolff/minify](https://github.com/tdewolff/minify) didn't need to stay.

`mage serve` gets a dev mode. It keeps an esbuild context between rebuilds, so only what changed gets rebuilt, and it writes unminified files with source maps and stable names. Breakpoints in devtools survive a rebuild that way, and the dev server sends `no-store` so the browser doesn't hang on to a stale copy.

## Finishing off the skeleton

The site started as an AI-generated Hugo skeleton, and I swapped Hugo for my own generator a while ago. What I hadn't noticed is that the Hugo templates never left. `layouts/` and `hugo.toml` were still sitting there, and the generator reads from `templates/`.

I found out the honest way: I edited `layouts/single.html`, and nothing changed. Two folders of templates that look nearly the same is an easy trap, so out they went.

## Styling adjustments

With the build sorted, I spent the rest of the day on the styles. Most of it is groundwork, not anything you'd notice at a glance.

This is one of the downsides of starting from a vibe-coded base. The styles work, but they aren't organized the way I'd have organized them writing from scratch: one-off values, rules fighting each other, no system tying it together. What follows is a start at restructuring, not the finished version. I'll keep tightening it up over time.

### Splitting up the stylesheet

`main.css` was one long file with comment banners between sections. It's now a list of imports, with each section in its own file under `base/`, `layout/`, `pages/`, and `themes/`. esbuild bundles them back into one stylesheet, so the split costs nothing at runtime.

The palette got the same treatment. Each color is a named variable now (`--palette-moss`, `--palette-bark`, `--palette-forest-floor`), and the variables the rest of the CSS actually uses (`--accent`, `--bg`, `--rule`) pick a light and dark value from those with `light-dark()`. Components only reach for the semantic names.

### A font-family variable

The site switches between two families, Fira Sans and Fira Code, in a lot of places. Instead of setting `font-family` in each one, every element now reads its family from a variable:

```css
:root { --font-family: var(--sans); }
* { font-family: var(--font-family); }

code, pre, kbd { --font-family: var(--mono); }
```

The variable inherits, so setting it on an element switches everything inside it too. The header, footer, dates, and table headers all switch to mono with one line each.

### Fewer font files, then variable ones

The site was loading eight static font files. Trimming that down to the weights actually in use got me thinking about variable fonts instead.

Fira Code has a variable version on Fontsource. Fira Sans doesn't, but [Firava](https://github.com/hellogreg/firava) is a variable build of it. Its character set is much smaller than Fira Sans: about 450 characters instead of 1,700. Everything I write is in English, and checking every post turned up nothing it was missing.

Cut down to the latin range and the weights I use (300 to 500), the upright file is 37.6 KB, which is less than the two static weights it replaces. Italic was closer. Static italic files would have been a bit smaller on most pages, but one variable italic file means italic headings don't pull in a new font when you move between pages. I'll take consistency over 14 KB.

That's three font files instead of eight. A small script rebuilds them, since the cut-down Firava is a file I made, not one I downloaded.

### A type scale

Font sizes had drifted into a scatter of one-off values: `0.85rem`, `0.9em`, `0.95rem`, `1.1rem`, `1.15rem`. They all come from a scale now. The base is 18px, each step up multiplies by 1.25, and one step down covers the small text (dates, meta, footnotes, the nav).

Line heights come from the same scale. Body text is 1.25 squared, 1.5625, and headings are 1.25.

The one deliberate exception is monospace. Fira Code runs larger than Fira Sans at the same size, so code sits a little under whatever it's set in, not on a step of the scale.

### GitHub-style alerts

I wanted callouts, and the markdown already exists for them:

```markdown
> [!NOTE]
> Worth knowing.
```

goldmark doesn't render these, so I wrote a small extension for it. After parsing, it looks for blockquotes whose first line is just the marker, and swaps them for an alert with a title. Like GitHub, it only knows the five types, and a marker that isn't on its own line stays a normal quote.

> [!TIP]
> This is one. They come in note, tip, important, warning, and caution, each with its own color from the palette.

### Smaller fixes

- **Mermaid diagrams are flat.** Mermaid's newer default look adds drop shadows and gradients. The classic look with plain borders fits the rest of the site better.
- **The footer sits at the bottom.** Short pages used to end with the footer partway up the window. A flex column on the body fixes that, without a fixed minimum height on `main`.

## What's next

The build step comes off the list from last time. The rest is still there: an accessibility pass, a sitemap and `robots.txt`, and git hooks. Styles-wise, the groundwork is in place, so the next round can be about how things look instead of how they're put together.
