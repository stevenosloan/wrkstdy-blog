package site

import (
	"strings"
	"testing"
)

func TestRenderRecipe(t *testing.T) {
	src := `serve with rice
  stir together, simmer 5 min
    simmer until soft, 20 min
      1 cup red lentils
      3 cups water
    2 tbsp ghee
`
	got, err := renderRecipe(src)
	if err != nil {
		t.Fatal(err)
	}
	// Three columns: ingredients, the lentil step, then "stir together", then the dish.
	// The ghee goes straight into "stir together", so it spans the lentil step's column.
	want := `<div class="recipe-table">
<table>
<tbody>
<tr><td class="ingredient">1 cup red lentils</td><td class="step" rowspan="2">simmer until soft, 20 min</td><td class="step" rowspan="3">stir together, simmer 5 min</td><td class="step dish" rowspan="3">serve with rice</td></tr>
<tr><td class="ingredient">3 cups water</td></tr>
<tr><td class="ingredient" colspan="2">2 tbsp ghee</td></tr>
</tbody>
</table>
</div>
`
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestRenderRecipeSeveralParts(t *testing.T) {
	got, err := renderRecipe("dal\n  simmer\n    lentils\n    water\nsalt to taste\n")
	if err != nil {
		t.Fatal(err)
	}
	// A shorter part stretches out to the last column, so each part ends on the right.
	if !strings.Contains(got, `<td class="ingredient" colspan="3">salt to taste</td>`) {
		t.Errorf("got\n%s\nwant salt to taste spanning every column", got)
	}
}

func TestRecipeBlock(t *testing.T) {
	b := &builder{md: newMarkdown()}
	got, err := b.renderMarkdown([]byte("```recipe\ntoast\n  bread\n```\n"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `<td class="step dish">toast</td>`) {
		t.Errorf("got %q, want a recipe table", got)
	}
	if _, err := b.renderMarkdown([]byte("```recipe\n```\n"), nil); err == nil {
		t.Error("want an error for an empty recipe block")
	}
}
