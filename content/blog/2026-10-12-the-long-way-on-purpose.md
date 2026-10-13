---
title: "Building a blog the long way, on purpose"
description: "Fresh off Mailchimp, I'm building a deliberately overengineered blog to learn Go and try out new ways of working with AI."
draft: false
---

I'm writing a blog.

> Narrator: you may be wondering how I found myself here.

After 14+ years at Mailchimp, I have something I haven't had in a long while --- time to poke at things just to see how they work. I want somewhere to keep track of that, so here we are: a blog.

A blog generator is a pretty solved problem though. I could have grabbed Hugo or 11ty, dropped in a theme, and been writing this post in an afternoon. Which is exactly why I'm not doing that. (spoiler: Based on how I started it, this still took just an afternoon. Maybe just a longer one)

## Overengineering on purpose

I'm treating this site as a practice field. I already know what a blog needs to do --- turn some markdown into HTML, list some posts, look decent in light and dark --- so I don't have to figure out *what* to build. That leaves all my attention for *how*, and the how is where I want to learn.

Go has been on my radar for a long time. I've dabbled, but never really had the hands-on time to learn it properly. I like [how small the language is](https://commandcenter.blogspot.com/2012/06/less-is-exponentially-more.html) and how that nudges everyone toward the same style --- one formatter, one loop keyword, not a lot of ways to be clever with the language itself.

So the site is built by a small generator I'm writing myself: [goldmark](https://github.com/yuin/goldmark) for markdown, the standard library's `html/template`, `fsnotify` watching files for the dev server, and [Mage](https://magefile.org) to run tasks. None of it is necessary. That's kind of the point.

Building a blogging engine isn't really new for me. I have deep experience with static site generators via [Middleman](https://middlemanapp.com), and I've tried this exact trick before. Digging through my old site, I found a draft from 2014 for "Middlewareman" (yes, the name was a pun), a static site generator I was going to write in Clojure to learn a bit more of the language. And now twelve years later --- same idea, different language.

## Starting from a vibe-coded skeleton

I'll be upfront: the first real commit of this site wasn't mine, not really. I described what I wanted (a forest color scheme was obviously a must) and had AI generate a Hugo skeleton --- layouts, syntax highlighting, the works. I didn't want the lack of a build system standing between me and getting started.

Since then, I've made a few "critical" edits to get closer to what I had in mind:

- **Swapped Hugo for my own Go generator.** Hugo does its job well, maybe too well. It's hard to learn Go from a tool that does everything for you.
- **Cut tags, the separate post archive, and RSS.** The skeleton came with all the things a "real" blog has. With one post, the home page list is archive enough, and I don't have anything to tag yet, so out they went. They can come back when they earn it.
- **Added Mage.** I know Rake well, so I went looking for something similar in Go, and Mage's Rake-style commands were a great fit. Now it's `go tool mage serve` instead of remembering a handful of `go run` incantations, and Netlify runs the same `check` task I do.

These changes move it more towards what I want than the starting point. I expect that to keep going, and I'll write about it as it does.

## Working with AI, differently

At Mailchimp I had a solid AI toolset, but it was about *doing work* --- getting a known task done faster. Here I want to explore the ways of working themselves: what's worth handing off, what I want to keep my hands on, and where the back-and-forth actually makes the result better.

Up to now I've been heavily invested in a plan → implement → review workflow: agree on the approach first, build it, then look the result over. It works, and I want to see what else is out there. Next I'd like to try playbooks and loops, where the steps are written down once and run repeatedly, instead of me walking each task through the same stages by hand.

This post is a small example of where things stand. I outlined my points and notes, then had AI help fill them in with prose, and steered it from there. The voice isn't fully mine yet --- putting together a solid voice and tone guide is on my to-do list.

## What's next

First, more adjustments to the blog itself, including a pass to check it's ready for production. Beyond that, I have a few projects in mind that I'll share here once they're started, along with how I worked through them --- dead ends included. Two I'm eyeing: a fishing notes app that ties your notes to a location and its watershed, and a CSS themer for raw markdown. No promises on either yet.
