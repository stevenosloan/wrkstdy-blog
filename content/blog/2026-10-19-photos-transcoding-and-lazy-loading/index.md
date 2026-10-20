---
title: "Photos: transcoding and lazy loading"
description: "Teaching the generator to handle photos: page bundles, resizing, AVIF and WebP with a JPEG fallback, keeping location data out, and getting lazy loading and sizes=\"auto\" to work together."
draft: true
---

The site has two new sections: a [journal](/journal/), for photos and notes from trips, and [recipes](/recipes/), where the first one is [ssamjang wings](/recipes/ssamjang-wings/). Both lean on photos, but the generator didn't know what a photo was or how to handle them. Straight from the camera or phone they could be 5 MB or more, so dropping them in as-is wasn't an performant option.

Going in, I knew it needed two things:

- **Configurable eager or lazy loading.** Most photos are further down the page and can wait until you scroll to them, but one at the top shouldn't. Which is which depends on where the photo sits, so it has to be set per photo.
- **Resizing and transcoding.** Each photo in a few widths and a few formats, so the browser can pick the best format it supports at the size it needs.

## Keeping a post and its photos together

The first change is structural. A post used to be a single markdown file, and anything it pointed to lived somewhere else. Now a post can be a folder that holds the content and its resources together: the markdown as `index.md`, with the photos beside it. The folder takes the place of the filename (date, then slug):

```
content/recipes/2026-10-10-ssamjang-wings/
  index.md
  trail.jpg
  seasoned.jpg
  ...
```

In the markdown, a photo is referred to by its filename. A line right under it, with no blank line between, becomes its caption:

```markdown
![Seasoned wings on a wire rack](seasoned.jpg)
Salted, peppered, and drying under a fan while the oven preheats.
```

## Resizing

Every photo a page uses is resized to 640, 960 and 1280px wide. The text column is 640px, so those cover 1x and 2x screens, and phones at 3x, which are narrower than the column. I had a 1920px size at first, for 3x screens at desktop width, but it came to as much as the other three together, for screens that are rare. Widths wider than the photo are skipped, so nothing gets upscaled.

Phone cameras don't rotate the pixels when you hold the phone sideways. They store the rotation in the photo's EXIF data and leave it to whatever displays it. The resized copies don't carry EXIF, so the rotation gets applied before resizing.

## Transcoding

Each width is encoded three times: AVIF, WebP, and a JPEG to fall back on. They go in a `<picture>`, and the browser takes the first format it supports:

```html
<picture>
  <source type="image/avif" srcset="seasoned.640w.avif 640w, …" sizes="…">
  <source type="image/webp" srcset="seasoned.640w.webp 640w, …" sizes="…">
  <img src="seasoned.1280w.jpg" srcset="seasoned.640w.jpg 640w, …" sizes="…"
       width="1280" height="960" loading="lazy" alt="…">
</picture>
```

Go's standard library can't write WebP or AVIF. [gen2brain/avif](https://github.com/gen2brain/avif) and [gen2brain/webp](https://github.com/gen2brain/webp) wrap libavif and libwebp compiled to WebAssembly, so the build still needs no C toolchain, locally or on Netlify.

### Quality numbers don't compare

I started with AVIF at 60 and WebP at 70. On my first test photo, a smooth landscape, AVIF came out bigger. Since browsers pick AVIF first, most visitors would have been getting the larger file.

On a busier photo it goes the other way. A quality number only means something within one encoder, so how two formats compare depends on the photo. AVIF at 50 came out smallest on both, so that's where it landed.

Here's the [trail photo](/recipes/ssamjang-wings/) from the wings post at 1280px, the size most screens get, with the same 256 by 160 crop from each at full size:

| Format | Quality | Size | Crop |
| --- | --- | --- | --- |
| original | | 9.4 MB, before resizing | ![A small waterfall over mossy rock, from the photo resized without compression](crop-original.webp) |
| JPEG | 80 | 888 KB | ![The same crop as a JPEG at quality 80](crop-jpeg-80.webp) |
| WebP | 70 | 710 KB | ![The same crop as a WebP at quality 70](crop-webp-70.webp) |
| AVIF | 60 | 554 KB | ![The same crop as an AVIF at quality 60](crop-avif-60.webp) |
| AVIF | 50 | 394 KB | ![The same crop as an AVIF at quality 50](crop-avif-50.webp) |

The crops are saved losslessly, so each one shows exactly what its format did. WebP smooths out the speckle on the rock and lifts the dark shadows, so the rock goes a little gray. AVIF at 50 smooths the speckle a bit and softens the leaves at the top. JPEG at 80 and AVIF at 60 stay closest to the original.

### Build Artifact Caching

Encoding is slow, AVIF most of all. The wings post has eight photos, and encoding them the first time takes about 40 seconds. Every copy goes into `.cache/images/`, in a folder named by a hash of the photo and the encoding settings. A rebuild only encodes what changed, and changing a quality setting starts over on its own.

That cache lives on my machine, though. Netlify starts every deploy from a clean checkout, so each one would encode every photo again. Netlify build plugins can keep a folder between deploys, and that only takes a small local plugin, in `netlify/plugins/photo-cache`:

```js
const dir = ".cache/images";

module.exports = {
  async onPreBuild({ utils }) {
    await utils.cache.restore(dir);
  },
  async onPostBuild({ utils }) {
    await utils.cache.save(dir);
  },
};
```

It restores the cache before the build and saves it after, so a deploy only encodes photos that are new or changed. Since the whole folder is restored and saved every time, it shouldn't grow forever: the build marks each entry it uses, and prunes any it hasn't used in 30 days.

### Leaving location out

Photos can record where they were taken, so anything taken in my kitchen could record where I live. Only the encoded copies are published, never the originals, and re-encoding drops EXIF along with that location.

That covers the site, but not the repo. A folder post keeps its photos in `content/`, and the repo is public. So photos don't get copied in by hand. A mage task brings them in:

```sh
go tool mage photos content/recipes/2026-10-10-ssamjang-wings ~/Desktop/wings
```

It takes a photo or a folder of them, and writes each one into the post rotated upright, scaled to at most 2560px, and with no metadata at all. That also keeps the repo from filling up with 10 MB originals. It won't overwrite a photo that's already there unless it's run with `-replace`.

Dropping all the metadata drops the color profile too, and iPhone photos are in Display P3, a wider color space than sRGB. Without the profile, browsers read P3 colors as sRGB and show them muted, so the task converts them to sRGB first. The build does the same for any photo that didn't come through it.

## Lazy loading

Most photos are further down the page than anyone has scrolled yet, so they load lazily. That also unlocks `sizes="auto"`.

`sizes` tells the browser how wide an image will be drawn, so it can pick from the `srcset` before layout has happened. Normally you write that out by hand, and it's easy to get wrong. With `auto`, the browser waits for layout and uses the real width. It can only do that for a lazy image, since it has to wait for layout anyway.

`auto` [isn't Baseline yet](https://caniuse.com/wf-sizes-auto), so it can't be the whole value. It goes first, and the hand-written sizes follow as a fallback. A browser that supports `auto` uses it and ignores the rest. One that doesn't skips it and reads the list after it:

```
sizes="auto, (max-width: 42rem) 100vw, 40rem"
```

The catch is a photo at the top of the page. It's on screen straight away, and making it wait is just slower. My first take was to load the first photo on each page eagerly, automatically. That works until a post opens with a few paragraphs and the first photo is far down.

So it's declarative now. Photos are lazy unless marked, with an attribute block right after the image:

```markdown
![A friend walking a rocky trail](trail.jpg){loading=eager}
Out on the trail.
```

goldmark's attribute syntax only covers headings, so this is another small extension. It reads the block, removes it from the text, and gives that photo `sizes` without `auto`. Anything else in braces fails the build, so `{loading=eagre}` doesn't quietly do nothing.

## What's next

- **Covers for the journal.** `/journal/` is the same list of dates and titles as the blog. A photo per trip would suit it better, I'll tackle that when I get to actually writing an entry there.
