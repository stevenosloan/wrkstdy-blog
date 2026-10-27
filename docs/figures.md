# Figures

A paragraph that starts with an image becomes a `<figure>`. Text on the lines right under the image, with no blank line between, becomes its `<figcaption>`.

```markdown
![Rowboat tied to the dock](dock.jpg)
First morning: fog didn't lift until *ten*.
```

The caption is normal inline markdown, so it can have links and emphasis. Leave it off for a figure with no caption.

The alt text and the caption do different jobs: alt text describes the photo for someone who can't see it, and the caption is what I'd say about it. See `voice/photo-captions.md`.

## Photos in a page bundle

A relative path, like `dock.jpg`, points at a photo in the post's folder. Those are resized to 640, 960, and 1280px wide, encoded as AVIF, WebP, and a JPEG fallback, and served in a `<picture>` with `srcset`, `sizes="auto"`, and `loading="lazy"`. The readme's "Journal entries and photos" section covers the pipeline, and how to import photos with `go tool mage photos`.

Absolute paths (`/static/x.png`) and URLs are left as they are, with no resizing.

## Loading a photo straight away

Photos load lazily. For one near the top of the page, add an attribute block right after the image:

```markdown
![A friend on the trail](trail.jpg){loading=eager}
Out on the trail, where these got talked about.
```

`loading=eager` and `loading=lazy` are the only attributes. Anything else in the braces, or a missing `}`, stops the build, so a typo doesn't go unnoticed.

## Errors

The build stops when an image:

- points outside the post's folder (`../other-post/photo.jpg`),
- can't be read or decoded,
- has a color profile other than sRGB or Display P3.

Code: `internal/site/figure.go`, with resizing in `internal/site/images.go`.
