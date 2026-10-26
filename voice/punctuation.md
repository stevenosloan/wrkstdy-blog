---
id: punctuation
summary: "Em dashes, throwaway asides, and italics for stress."
depends_on: [voice-core]
status: accepted
checks:
  - id: en-dash-as-break
    pattern: '\s--\s'
    message: a spaced -- renders as an en dash; use a spaced --- for a break
---

# Punctuation

### Em dashes are a spaced ---

Write `word --- word`. goldmark's typographer turns `---` into an em dash and `--` into an en dash, and an en dash is for ranges, not breaks in a sentence.

Why: it's a common slip of mine, and the two look nearly the same in the editor.
When unsure: if the break could be a comma or parentheses, it probably should be.

### Throwaway asides

All lowercase, no closing punctuation, in parentheses: `(spoiler: this still took just an afternoon, maybe just a longer one)`. If an aside runs to more than one sentence, use normal capitalization inside it.

Why: a throwaway aside should read like one, quieter than the sentence around it.
When unsure: if the aside is important, it isn't an aside. Make it its own sentence.

### Italics for stress

Use *italics* on the one word I'd lean on when speaking. Not bold.

Why: it's how I emphasize out loud; bold reads as shouting.
When unsure: one stressed word per paragraph, at most.
