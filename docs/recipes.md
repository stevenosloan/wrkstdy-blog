# Recipes

A ` ```recipe ` block is a recipe as a table, in [Michael Chu's notation](https://www.cookingforengineers.com/): ingredients down the left, each step in a cell to the right of everything that goes into it, so the table narrows to the finished dish on the far right.

Write it as an outline with the finished dish first. Each line is a step. Indented under a step are the ingredients and earlier steps it uses, and a line with nothing under it is an ingredient:

````markdown
```recipe
serve with rice
  stir together, simmer 5 min
    simmer until soft, 20 min
      1 cup red lentils, rinsed
      3 cups water
    fry until the seeds pop
      2 tbsp ghee
      1 tsp cumin seeds
```
````

Each ingredient gets a row. A step's cell spans the rows of the ingredients under it, and stretches across columns when its inputs are fewer steps deep than its neighbors'. The finished dish is marked with a `dish` class.

## Rules

- A line goes under the nearest line above it that's indented less. How far is up to you; a tab counts as 4 spaces.
- A block can have more than one top-level line, for parts served separately, like a sauce.
- Blank lines are ignored.
- An empty block stops the build.

For how to write the steps themselves, see `voice/recipe-steps.md`.

Write the talk-through around the block as normal markdown.

Code: `internal/site/recipe.go`. Styles: `assets/css/pages/recipe.css`.
