# TODO comments

An HTML comment on its own, starting with `TODO`, marks a gap in a draft.

```markdown
<!-- TODO(steven): the concrete opening. One or two sentences. -->
```

On a page with `draft: true`, it renders as an alert with a "TODO" title and a dashed border, so the gaps show while previewing with `go tool mage serve -drafts`. On a published page it stays a comment, which goldmark drops, leaving only a `<!-- raw HTML omitted -->` placeholder in the HTML. The TODO's text never reaches the page.

## Forms

All of these work; the name in parentheses and the colon are optional:

```markdown
<!-- TODO: the opening -->
<!-- TODO(steven): the opening -->
<!-- TODO the opening -->
```

A TODO can run over several lines inside the one comment.

## Rules

- The comment has to start its own line. A comment in the middle of a paragraph's text is dropped, draft or not.
- Any other comment, like `<!-- note to self -->`, is dropped too.
- Published pages never show a TODO, so removing `draft: true` with TODOs left in won't break anything, but it will leave gaps. Search the post for `TODO` before publishing.

Code: `internal/site/todo.go`. Styles: `.alert-todo` in `assets/css/base/alerts.css`.
