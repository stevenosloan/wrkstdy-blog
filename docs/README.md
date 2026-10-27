# Writing markdown for wrkstdy

Posts are rendered by [goldmark](https://github.com/yuin/goldmark), set up in `internal/site/markdown.go`. Most of it is standard markdown. These docs cover what this site adds on top, and the few places it differs.

## Extensions

| Doc | Syntax | Renders as |
| --- | --- | --- |
| [Alerts](alerts.md) | `> [!NOTE]` | A callout box with a title |
| [TODO comments](todos.md) | `<!-- TODO: ... -->` | A dashed callout on drafts, nothing once published |
| [Figures](figures.md) | `![alt](photo.jpg)` with a line under it | A `<figure>` with a caption and resized photos |
| [Code blocks](code-blocks.md) | ` ```lang `, ` ```lang summary="..." ` | Highlighted code, optionally collapsed |
| [Mermaid](code-blocks.md#mermaid) | ` ```mermaid ` | A diagram |
| [Recipes](recipes.md) | ` ```recipe ` | A recipe table |
| [Shortcodes](shortcodes.md) | `{{< name >}}` | A template from `templates/shortcodes/` |

## Standard, but worth knowing

- **GitHub flavored markdown** is on: tables, ~~strikethrough~~, task lists, and bare URLs as links.
- **Footnotes** work: `a claim[^1]`, then `[^1]: the source` anywhere in the post.
- **Typographer** turns `---` into an em dash, `--` into an en dash, `...` into an ellipsis, and straight quotes into curly ones. Write `word --- word` for a break in a sentence; `--` is for ranges.
- **Heading IDs** are generated from the text, so `## Resizing` can be linked as `#resizing`.
- **Raw HTML is dropped.** A `<details>` or `<div>` in a post renders as nothing, just a `<!-- raw HTML omitted -->` placeholder in the source. That's on purpose: it keeps [TODO comments](todos.md) out of published pages. When a post needs something markdown can't do, add an extension or a [shortcode](shortcodes.md).

## Checking a post

`go tool mage serve -drafts` shows drafts, TODOs included. `/blog/styleguide/` shows every styled element in one place.
