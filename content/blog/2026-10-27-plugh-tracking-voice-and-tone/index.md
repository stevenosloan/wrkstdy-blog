---
title: "plugh, tracking voice and tone"
description: "Why I split my writing rules into small chunks, routed by what's being written, so AI only reads the ones that apply. Plus the checks that catch my usual slips, and where they don't."
draft: false
---

Two weeks ago, alongside the first post here, I started a file called `notes/style-guide.md`. It was a scratchpad: a handful of rules about dashes, slashes, and the word "markdown", and two of them ended with "Common slip of mine." It wasn't a style guide yet, just a list of things I either got wrong or disagreed with in how AI wrote them.

Admittedly, most of what I write here now involves AI, either drafting from my notes or reviewing what I wrote. The rules matter in both directions: they keep me honest when AI reviews a draft, and they guide it when it writes off of my notes. That's where the scratchpad started to strain. A notes file is easy for me to keep in my head. For AI, it either gets pasted in whole every time, or it doesn't get read at all.

Two things pushed it from a notes file to a tool. I knew I'd have other projects needing copy soon, each with its own voice and readers, and a scratchpad in this repo wasn't going to follow me there. And I'd been reading Little Plains' [Agentic Systems Are the Future of Brands](https://agentic.littleplains.com/). It describes a brand kit written for people and agents alike, with a readme that points each job at the files it needs. The line that stuck with me: "It doesn't need every file every time, but it does need to know where to start." I wanted a practical way to apply that kind of focused prompting to my own writing, across this blog and whatever came next.

So I built [plugh](https://github.com/stevenosloan/plugh): a small tool for keeping voice and tone rules in pieces, and handing AI only the pieces that apply to what it's writing. This post covers why I split the guide up, how the pieces get routed, the checks that catch my usual slips, where those checks still get it wrong, and how a draft gets from AI to done.

## One file doesn't fit every page

The site has more than one kind of writing. A blog post is for a developer who found it through a search and skims the headings. A recipe is for someone standing at the counter with wet hands who wants succinct steps. A front matter description is one or two sentences on the homepage.

Some rules hold for all of them: write as "I", say the specific thing, explain the why. But "open on a concrete moment" is wrong for recipe steps, and "one step per line, amounts first" is wrong for a post. One long guide means every piece of writing carries rules for every other kind.

Splitting it up helps in two ways. For me, it means I can write a rule for the situation it's actually about, instead of hedging it to cover every kind of page. For AI, it keeps the context lean and focused: it reads the rules for what it's writing, and nothing else to weigh against them. A recipe rule can't pull a blog post off course if it never gets loaded.

I'll admit I never got as far as the one big file. The scratchpad was still small when I split it up. But between the Little Plains piece and knowing that every rule rides along with every request, paid for in tokens each time, it seemed worth tackling before the guide grew rather than after. (so no dead end here, just a cost I decided not to wait for)

## Chunks, and a router

In plugh, each rule set is a *chunk*: a short markdown file in `voice/`, like `voice/punctuation.md` or `voice/recipe-steps.md`. Each rule in a chunk says the rule, why it exists, and what to do when unsure. The "when unsure" line ended up being the part I like best, because it's what lets the rule stretch to a case I didn't think of.

A `voice.yaml` at the project root routes chunks to content by path. Here's the part of mine for blog posts:

```yaml
types:
  blog:
    match: content/blog/
    always:
      - voice-core
      - punctuation
      - word-choice
      - blog-post
    if_relevant:
      - photo-captions
    budget: 1500 # words
```

`always` chunks load every time. `if_relevant` chunks show up as a one-line summary, and only get loaded when they apply, like the captions chunk for a post that has photos. A type doesn't have to be a folder, either. The front matter description is a type of its own, routed by the job instead of a path.

Running `go tool plugh context` on this post prints exactly what AI reads before writing it: who's writing, who it's for, and the 4 chunks.

```markdown summary="The full output of go tool plugh context for this post"
# Voice for blog (content/blog/)

Follow these chunks. Where they're silent, follow their "When unsure" lines.

## Writing as me

Steven, a developer writing a personal log. Works things out in public, keeps the dead ends, and would rather be useful than impressive.

## Writing for curious-dev

A developer who found the post through a search or a link. Knows the basics, hasn't solved this exact problem, and wants the choices and the reasons more than a tutorial. Skims headings first.

<!-- chunk voice-core, from voice/voice-core.md -->
# Voice core

Everything on wrkstdy is me talking to one reader. These are the defaults every other chunk builds on.

### First person, plain words

Write as "I", the way I'd explain it to a friend across the table. Prefer the plain word over the clever one. "We" is for me and AI working on something together, and only when that's who did it.

Why: the site is a personal log, not a product. Polish that hides the person defeats the point. A "we" that means me and AI is honest about how the work happened; any other "we" blurs who's talking.
When unsure: read the sentence out loud. If I wouldn't say it, rewrite it. If "we" could mean the reader too, name who did it.

### Specific over general

Name the thing: the tool, the number, the place, the moment. "5 MB straight off the phone" beats "large files".

Why: specifics are what make a post mine instead of generic.
When unsure: ask "which one?" of every general noun, and answer it.

### Explain the why

When I made a choice, say why, including the options I didn't take.

Why: the reasoning is usually the useful part for a reader, and the part I want to remember later.
When unsure: if a paragraph only says *what* happened, it's probably missing a *because*.

### Candid, a little self-deprecating

It's fine to admit a mistake, a dead end, or that something took longer than it should have. Light humor at my own expense, not at anyone else's.

Why: honesty about the process is the reason to read a dev log at all.
When unsure: poke fun at me, never at a person or a team.

### Stay humble

Present what I did as what worked for me, not the one right way. Say where my experience runs out, and leave room for the reader to know something I don't.

Why: I'd rather be useful than impressive, and a post that sounds like it has every answer is harder to trust.
When unsure: if a sentence tells the reader what they *should* do, check whether "what I'd do" is closer to the truth.

<!-- chunk punctuation, from voice/punctuation.md -->
# Punctuation

### Em dashes are a spaced ---

Write `word --- word`. goldmark's typographer turns `---` into an em dash and `--` into an en dash, and an en dash is for ranges, not breaks in a sentence.

Why: it's a common slip of mine, and the two look nearly the same in the editor.
When unsure: if the break could be a comma or parentheses, it probably should be.

### Throwaway asides

All lowercase, no closing punctuation, in parentheses: `(spoiler: this still took just an afternoon, maybe just a longer one)`. If an aside runs to more than one sentence, use normal capitalization inside it. An aside can sit inside a sentence or stand on its own between two, in the same format either way.

Why: a throwaway aside should read like one, quieter than the sentence around it.
When unsure: if the aside is important, it isn't an aside. Make it its own sentence.

### Italics for stress

Use *italics* on the one word I'd lean on when speaking. Not bold. Italics also mark a term the first time it's defined, like *chunk*.

Why: it's how I emphasize out loud; bold reads as shouting. A new term gets the same lean, once.
When unsure: one stressed word per paragraph, at most. A term is italic only where it's defined, not after.

### Oxford commas

In a list of three or more, put a comma before the "and" or "or": "dashes, slashes, and the word markdown".

Why: it's how I punctuate, and it keeps the last two items from reading as a pair.
When unsure: a pair that's one thing, like "salt and pepper", counts as a single item. That's why this check is a warning.

<!-- chunk word-choice, from voice/word-choice.md -->
# Word choice

### "Or", not a slash

"Hugo or 11ty", not "Hugo/11ty".

Why: a slash makes the reader decide what it means (and, or, either).
When unsure: paths, URLs, and fractions are fine; this is about words.

### "markdown" is lowercase

Always, including at the start of a list item. (At the start of a sentence, rephrase.)

Why: markdown is an idea, plain text with a little punctuation for structure, more than any one spec. CommonMark, GitHub's flavor, and goldmark all disagree on the edges, and I'm writing to the idea.

### Project names keep their casing

A lowercase name like plugh or tributary stays lowercase, even at the start of a sentence or a title.

Why: it's the name as written; capitalizing it makes a different word.
When unsure: this is for names. A common word like markdown still gets rephrased.

### Numerals for counts

Anything counted or measured is a numeral: "6 warnings", "1,500 words". At the start of a sentence, rephrase.

Why: a numeral is easier to spot when skimming, and mixing "6" and "four" in one paragraph reads as a slip.
When unsure: if the number isn't the point ("one or two sentences", "a thing or two"), words are fine.

### "I have", not "I've got"

Why: a common slip of mine, and "I have" is plainer.

### "AI", not the tool's name

When describing how work was done, say "AI".

Why: the post is about the way of working, not the vendor, and tool names date fast.
When unsure: naming a tool is fine when the post is *about* that tool. That's why this check is a warning.

### Repetition on purpose

Deliberate repetition for emphasis is fine ("wasn't mine, not really"). Don't edit it out, but question it.

<!-- chunk blog-post, from voice/blog-post.md -->
# Blog posts

### Open on something concrete

Start with a specific moment or a candid admission ("After a recent layoff...", "Here's a little secret"), not a thesis statement or a definition.

Why: it's how my posts have always started, and it earns the reader's attention before asking for it.
When unsure: cut the first paragraph and see if the second one is the real opening.

### Start broad, get specific

After the opening, give a short overview of what the post covers ("This post is a reflection on that, and a list of things I want to keep in mind..."), then cover it, one piece at a time. Each main topic gets a `##` heading; when a topic has parts of its own, break it up with `###` headings underneath.

Why: the reader knows where the post is going before the details start, and can decide which parts they need.
When unsure: every section should trace back to something in the overview, and everything in the overview should get a section.

### Headings carry the story

Short, plain `##` headings that tell the reader where the post is going ("Keeping a post and its photos together", "Resizing").

Why: posts get skimmed; the headings should work as a summary on their own.

### Show the work, including dead ends

When something was tried and dropped, say so and why ("I had a 1920px size at first, but...").

Why: the dead ends are often the most useful part for someone doing the same thing.

### Callouts for side points

A `> [!NOTE]` is for a side point the reader can skip without losing the thread. A few a post, at most.

Why: it lets a skimmer jump past it, and a post full of callouts has no thread left to follow.
When unsure: if the post doesn't make sense without it, it's a paragraph, not a callout.

### Close on what the reader takes away

End on what I hope the reader does or thinks differently, not on a recap.

Why: a recap repeats the post; a takeaway gives it a point.
When unsure: if the last paragraph starts with "In summary", rewrite it.

## Load if relevant

If a summary applies to what you're writing, run `go tool plugh chunk <id>` and follow it too.

- photo-captions: Alt text vs captions on photos in any section.
```

### Who's writing, and for whom

The router also names *personas*. For now --- mine are "me", "curious-dev", and "home-cook", each a few sentences long. Blog posts are written as me, for curious-dev. Recipes swap the audience to home-cook and keep the rest.

Someone reading a post and someone following a recipe want different things from the same site. Naming the reader once, up front, means the rules don't each have to.

### A budget on purpose

Each type gets a word budget, and `plugh kit` fails if a type goes over. The budget isn't for the post. It's for the rules: the `always` chunks plus the persona descriptions, which is everything AI reads before it writes a word. Blog posts get 1,500 words, and right now they use a little over 1,000.

Every word of rules is context AI has to hold alongside the actual task, and a guide that grows without limit stops being followed, by people *and* by AI. So adding a rule means checking whether an old one still earns its place.

## Checks for the mechanical slips

Some rules are mechanical enough that a regex catches them. A chunk can carry checks in its front matter, and `go tool plugh check` runs them against every file a type covers. `I've got` is an error. `Markdown` with a capital M is an error. (the first draft of this post failed both, quoting them as examples) A slash between two words is a warning.

> [!NOTE]
> Lowercase "markdown" is a choice, not a typo. To me, markdown is an idea, plain text with a little punctuation for structure, more than any one spec. CommonMark, GitHub's flavor, and goldmark all disagree on the edges, and I'm writing to the idea, not one of them.

Warnings are questions, not orders. A tool name in a post is a warning because "say AI, not the tool" is the right default, but this post is *about* a tool, so naming it is fine.

### Where the checks get it wrong

Running it against the rest of the site gives 0 errors and 6 warnings, and all 6 are wrong. Of those, 4 are the slash check, catching link text like `[tdewolff/minify](https://github.com/tdewolff/minify)`, where the slash is part of a repo name, not a choice between two words. The other 2 are the Oxford comma check, which can't tell `a wire rack, salt and pepper them` from a list.

Misses like these have had me question the mechanical checks more than once. A regex doesn't know what a sentence means, and AI reviewing against the same rules does. But for rote checks, a regex is instant, costs nothing, and gives the same answer every time, which makes it far more efficient than asking AI. So I'm keeping them, as warnings wherever they're guessing. Each wrong warning is another example of where the line actually is, and as they pile up, I expect they'll point me toward something that works better.

## One format, many projects

plugh is a Go module and a Claude Code plugin in [one repo](https://github.com/stevenosloan/plugh). The plugin has two skills: one loads the routed chunks before writing, and one reviews a draft against them without rewriting it.

> [!NOTE]
> plugh is written for Claude Code, but the techniques aren't tied to it. Small chunks of rules, routed by what's being written, with checks for the mechanical slips, should carry over to other AI tools. `plugh context` just prints text, so anything that can run a command before it writes can read it.

plugh ships no rules of its own. The repo has a made-up example project to copy from, but the point is that the rules are yours.

I'm working on other projects with different voices and different readers, and none of this site's rules belong there. What they can share is the format: the same chunks, router, and checks in each project --- so the way I write rules down and hand them to AI carries over even when the rules themselves don't.

[tributary](https://tributary-site.fly.dev/) is one of them. It shows current conditions on the water for stream fishing in the US, so its readers are people using it, not developers looking to learn something. And even within tributary, a guide and the rest of the copy will want different rules, the same way a post and a recipe do here.

## Working through a draft, one point at a time

A post starts with the write skill. AI loads the chunks for the type, works from my notes, and writes a first draft. For this post, it left marked gaps wherever a part needed something only I know, like the moment that started it or a dead end I hit, instead of making one up.

After that first draft, I run the review skill and work through what it finds one point at a time. I pick a spot, say what's missing or what I'd say differently, and we change just that, then run `plugh check` before moving on. This post went that way: the point about keeping context lean, the budget, the Little Plains link, and this section were each their own pass.

> [!TIP]
> Working through a draft is the perfect time to update the rules. When I make the same fix twice, or the review makes a call no rule covers, that's a rule waiting to be written, and it's easiest to write while I'm still looking at the content that needed it.

That's just how I edit. Unless I'm making sweeping changes, I work on specific wording, a sentence or a paragraph at a time, and I didn't want AI to change that.

## Worth it, so far

It's working. Even this post took fewer edits from me after the first draft than I'm used to.

It also gets better as I go. Along with its findings, the review lists the calls it made that no rule covers yet. Those are candidates, not findings, and I decide which ones become rules. Working through this post added 6: when "we" is fine, numerals for counts, lowercase project names, italics for a new term, asides that stand on their own, and when a callout earns its place. The next post starts with all of them.

It isn't free, though. Every round of my workflow re-checks the draft with plugh, and that spends some tokens each time. Over a post's worth of small passes it adds up, but it's been worth it.

If you're writing with AI and the drafts don't sound like you, I'd start smaller than a style guide. Keep a list of the things you fix every time. That list is already most of a voice guide, and it's one nobody else can write for you.
