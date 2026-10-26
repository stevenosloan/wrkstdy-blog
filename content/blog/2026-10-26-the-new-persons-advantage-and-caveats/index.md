---
title: "The new person's advantage, and caveats"
description: "Being new lets you see what a team has stopped seeing. Context is what helps the team hear you say it."
draft: false
---

After a recent layoff at #lastjob, a handful of people joined our team from another part of the company. They were smart, they were capable, and they showed up with strong opinions about how things should work.

What they didn't show up with was much curiosity about how things *already* worked, or more importantly *why* they worked the way they did. This created friction that made it difficult to receive, triage, and apply those ideas.

I was on the receiving end of this one, and it's stuck with me. It was also on my mind while reading Vikram Singh's [Increase your slope](https://www.vikramsingh.me/increase-your-slope), which is written for new grads, but holds for anyone new to a team. One of their points is to get comfortable having opinions, and to do your homework and check them with a trusted person or two before taking them wider.

This post is a reflection on that experience, and a list of things I want to keep in mind the next time I'm the new person on a team. Here, I want to look at what it looks like when that homework --- gaining context --- gets skipped, why being new is still an advantage, and what the homework looks like.

## Simplifying a cache key

Here's the example I keep coming back to. We had an edge cache, and the cache key admittedly had a lot of dimensions in it. To someone new, that looks like a smell: too much cardinality, simplify it, an easy win to *fix* the hit rate.

But those extra dimensions came from features we'd introduced over time, like an existing user's ecommerce plan or a visitor's device type. They had widely varying levels of usage, and each one *could* change what the origin would send back. Just simplifying the cache key would mean some visitors might end up with a page that was meant for someone else, or a stale variant of the right one. (spoiler: nobody enjoys finding that out in production)

Whether every one of those differentiators still *mattered* was a fair question, but making that decision would need to involve stakeholders. The fact that they *could* matter was the context the newcomers were missing. They could describe the change they wanted, but weren't telling us what the current setup was protecting and how that risk would be handled.

So the conversation never started from "is this protection still worth it?" It started from "why is this so complicated?", and we spent our energy re-explaining instead of improving anything. This pattern repeated and got personal and political faster than I'd like to admit.

## Being new is an advantage

I don't believe being new means you shouldn't have opinions or share ideas. Being new is a real advantage, and it's worth saying so plainly.

- You don't know the patterns yet, so you don't know which ones are "just how it's done" and which are stale constraints.
- You're more willing to ask the obvious question everyone else stopped asking.
- Teams go blind to their own friction. The new person still feels it.

Sometimes the question is as simple as "why is the hero image lazy loaded?" It was asked to understand, not to correct, and it was a real catch. The team that built the page had missed some of our requirements and conventions, and it took someone new to notice.

Sometimes it's more fraught, like the blocking request for our experiments JS file, which we didn't cache. It was posed as a mistake, but it exposed a lack of trying to understand how the system worked. That file had to vary on criteria our cache didn't know about, down to the individual user. Caching it would have meant serving someone else's experiments.

The difference wasn't the question, it was how it was asked. The same lack of familiarity that lets you see the friction also hides the reasons behind it.

## Context has to be earned

It helps to start by assuming good intent. Most weird-looking things are weird for a reason: an outage, a customer, a constraint, a person who left. Some of those reasons are still good. Some expired years ago and nobody noticed.

From the new seat you can't tell which is which, and proposing a rewrite in week two costs trust even if you turn out to be right. Sometimes the context a decision was made under no longer applies, but you have to seek the context to know that.

### Back to the cache key

The stakeholder conversation it needed took context *and* clout, and the newcomers had neither yet. So even a good idea couldn't land on its own. Earning the context is also how you earn the standing to have the bigger conversation.

Their idea wasn't wrong. A simpler cache key was worth wanting. But it came without the context of who it needed to go to, the stakeholders behind each of those features, or of the effort it would take to implement safely: checking each dimension, and rolling it out without serving anyone a page meant for someone else.

So the question was never whether they were right or wrong. It was who needed to be part of the decision, and what it would take to make the change safely. That made it a conversation with stakeholders, not solely an engineering cleanup.

## Ask, but ask to understand

None of this means sitting on questions. We had a saying to encourage that: "there are no bad questions." I want them asked, but framed as "help me understand why" instead of "why don't we just."

> [!NOTE]
> "Just" can generally be considered a four-letter word. It carries a lot, and if you find yourself thinking it, take a step back and consider what it really means.

For myself, I want to make sure I learn the history before I propose a change. If the reason is still valid, it saves everyone a debate. If it's expired, I'll have the context to make the case and the trust to be heard.

### Some pre-work

There's a tool I like for checking whether I'm ready. George Heilmeier ran DARPA in the 1970s and wrote up a short list of questions for vetting proposed projects, now known as [the Heilmeier Catechism](https://www.darpa.mil/about/heilmeier-catechism). The one I'd borrow before suggesting anything new:

> How is it done today, and what are the limits of current practice?

If I can't answer that, I'm not ready to propose a change yet. Answering the first half *is* gaining context. Naming the limits is what turns a pitch into a case other people can weigh. And it's exactly what was missing from the cache key conversations: the newcomers could describe their idea, but not "how it's done today."

Before I take an idea to anyone else, these are the questions I want to test it against myself:

- Assuming good intent, why does this exist? What could it be providing?
- What would it take for this to change? Who would need to agree, and what would have to be true first?

## Spend it on curiosity

Newness is an asset that runs out, and I intend to use it the next time around. I want to be deliberate with it, though: spend it on curiosity, not conclusions. The conclusions will still be there in a few months, and they'll be better than they would be otherwise.
