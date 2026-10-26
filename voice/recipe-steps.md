---
id: recipe-steps
summary: "The process section of a recipe: second person, present tense, exact temperatures and times."
depends_on: [voice-core, punctuation]
status: draft
checks:
  - id: temperature-unit
    pattern: '\b\d{3}\s?(?:degrees|F\b)'
    message: write temperatures as 425°F
    severity: warn
---

# Recipe steps

### Second person, present tense

"Preheat the oven to 365°F on convection." Address the cook directly, one action per sentence where possible.

Why: the cook is reading this with their hands full.

### Exact numbers, written the same way every time

Temperatures as `425°F`, with "on convection" when it matters. Times as "15 minutes" in prose and "15 min" in the recipe table.

Why: these are the numbers someone actually follows, so they should be easy to spot.

### Say why a step matters

A short reason after the steps that aren't obvious ("The fan dries off the surface moisture, which helps them brown and crisp.").

Why: knowing why lets the cook adapt when their oven or ingredients differ.

### Loose amounts are fine

"A couple knobs of butter", "gochujang, to taste". This isn't a baking recipe.

Why: it's how I actually cook, and pretending to precision I don't use would be dishonest.
