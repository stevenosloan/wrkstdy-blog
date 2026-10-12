---
title: "Hello, world"
date: 2026-10-11
description: "The first post, and a quick tour of what a post can hold."
tags: ["dev-log", "meta"]
---

This site is built with [Hugo](https://gohugo.io) from plain markdown files
and deployed on Netlify. Posts are tagged, so writing and dev-log entries can
live side by side.

## Code

Fenced code blocks are highlighted in the browser by
[MicroLighter](https://github.com/davatron5000/microlighter):

```go
package main

import "fmt"

func main() {
	fmt.Println("hello from wrkstdy")
}
```

## Diagrams

A `mermaid` code block becomes a diagram:

```mermaid
flowchart LR
  md[Markdown files] --> hugo[Hugo build]
  hugo --> public[public/]
  public --> netlify[Netlify]
```

That's it for now.
