# Shortcodes

`{{< name >}}` drops in the template `templates/shortcodes/name.html`. It's the same syntax Hugo uses, but simpler: no arguments, and no closing tag.

```markdown
{{< swatches >}}
```

Put it on a line of its own. It works inline too, but the template's HTML ends up inside a paragraph.

## The shortcodes

| Name | What it shows | Used in |
| --- | --- | --- |
| `swatches` | The palette, as live swatches that follow light and dark mode | `/blog/styleguide/` |
| `specimen` | A type specimen for both font families, across their weight ranges | `/blog/styleguide/` |

## Adding one

Add `templates/shortcodes/<name>.html`. It's a Go `html/template`, run with no data, with the same template functions as the page templates (like `list` and `dict`). An unknown name stops the build.

Shortcodes are the way to get HTML into a post, since [raw HTML is dropped](README.md#standard-but-worth-knowing). When a shortcode needs input, that's probably a sign it should be an extension instead, like [figures](figures.md) or [collapsed code blocks](code-blocks.md#collapsed-blocks).

Code: `renderMarkdown` and `renderShortcode` in `internal/site/markdown.go`.
