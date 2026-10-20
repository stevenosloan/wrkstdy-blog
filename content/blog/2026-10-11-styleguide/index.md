---
title: "Styleguide"
description: "A full example of markdown formatting."
draft: true
---

## Palette

{{< swatches >}}

## Type

Both families are variable fonts: one file per style covers every weight in its range.

{{< specimen >}}

# Heading one
## Heading two
### Heading three
#### Heading four

Body text is set in Fira Sans. It carries **bold**, *italic*, ~~strikethrough~~,
`inline code`, and [links](/). A footnote looks like this.[^1]

> A blockquote. Good for pulling out a line from somewhere else, or an aside
> that deserves a little room.

## Alerts

> [!NOTE]
> Useful information that's worth knowing, even when skimming.

> [!TIP]
> A helpful suggestion for doing things better or more easily.

> [!IMPORTANT]
> Key information needed to get something right.

> [!WARNING]
> Something that needs attention right away to avoid problems.

> [!CAUTION]
> The risks or downsides of an action.

---

## Lists

- Unordered item
- Another item
  - A nested item
  - And one more
- Back to the top level

1. Ordered item
2. Second item
3. Third item

- [x] A finished task
- [ ] An open task

## Table

| Element | Font | Notes |
| --- | --- | --- |
| Body | Fira Sans | weight 300, 18px, 1.5625 line height |
| Headings | Fira Sans | weight 500 |
| Bold | Fira Sans | weight 500 |
| Code | Fira Code | weight 400, ligatures on |

## Recipe table

```recipe
serve with rice
  stir together, simmer 5 min
    simmer until soft, 20 min
      1 cup red lentils, rinsed
      3 cups water
      1 tsp turmeric
    fry until the seeds pop
      2 tbsp ghee
      1 tsp cumin seeds
  salt to taste
```

## Code

```go
// Go
func greet(name string) string {
	if name == "" {
		return "hello, stranger"
	}
	return fmt.Sprintf("hello, %s", name)
}
```

```js
// JavaScript
const posts = await fetch("/blog/index.xml").then((r) => r.text());
export const isDraft = (post) => post?.draft === true && !post.published;
```

```bash
# Shell
go tool mage new
go tool mage serve -drafts
```

```css
/* CSS */
:root { --accent: light-dark(#3f6b3a, #9cc283); }
a:hover { color: var(--accent-2); }
```

```json
{ "title": "Hello, world", "tags": ["dev-log", "meta"], "draft": false }
```

```diff
- font-family: "IBM Plex Sans";
+ font-family: "Fira Sans";
```

```
A code block with no language stays plain.
```

## Diagrams

```mermaid
flowchart LR
  write[Write markdown] --> build[mage build]
  build --> preview{Looks right?}
  preview -- yes --> ship[Push to Netlify]
  preview -- no --> write
```

```mermaid
sequenceDiagram
  participant You
  participant GitHub
  participant Netlify
  You->>GitHub: git push
  GitHub->>Netlify: webhook
  Netlify-->>You: deploy is live
```

[^1]: Footnotes collect at the bottom of the page.