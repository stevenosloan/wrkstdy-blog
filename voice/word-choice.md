---
id: word-choice
summary: "Small word rules: \"or\" not slashes, lowercase markdown and project names, numerals for counts, \"I have\", \"AI\"."
depends_on: [voice-core]
status: accepted
checks:
  - id: slash-alternatives
    pattern: '\b[A-Za-z]\w*/\w*[A-Za-z]\w*\b'
    message: name alternatives with "or", not a slash
    severity: warn
  - id: markdown-capitalized
    pattern: '\bMarkdown\b'
    message: '"markdown" is always lowercase'
  - id: ive-got
    pattern: "(?i)\\bI've got\\b"
    message: prefer "I have" over "I've got"
  - id: tool-names
    pattern: '\b(Claude|ChatGPT|Copilot|Cursor|Gemini)\b'
    message: say "AI", not a specific tool name, when describing how work was done
    severity: warn
---

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
