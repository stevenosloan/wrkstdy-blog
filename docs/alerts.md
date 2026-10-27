# Alerts

GitHub-style alerts: a blockquote whose first line is only a marker.

```markdown
> [!NOTE]
> plugh is written for Claude Code, but the techniques aren't tied to it.
```

That renders as a `<div class="alert alert-note">` with a "Note" title, not a blockquote. Everything after the marker line is normal markdown, so it can hold links, lists, or more than one paragraph.

## Types

| Marker | Title | Color |
| --- | --- | --- |
| `[!NOTE]` | Note | `--info` |
| `[!TIP]` | Tip | `--accent` |
| `[!IMPORTANT]` | Important | `--important` |
| `[!WARNING]` | Warning | `--accent-2` |
| `[!CAUTION]` | Caution | `--danger` |

These are the 5 GitHub recognizes, so a post reads the same on GitHub. Anything else, like `[!FYI]`, stays a plain blockquote.

## Rules

- The marker has to be alone on the first line. `> [!NOTE] Worth knowing.` stays a blockquote.
- The marker is uppercase.

For when to use one in a post, see the "Callouts for side points" rule in `voice/blog-post.md`.

Code: `internal/site/alerts.go`. Styles: `assets/css/base/alerts.css`.
