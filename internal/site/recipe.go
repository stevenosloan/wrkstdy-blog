package site

import (
	"errors"
	"fmt"
	"html"
	"strings"
)

// A ```recipe block is a recipe as a table, in the notation Michael Chu
// introduced on Cooking for Engineers: ingredients down the left, and each step
// in a cell to the right of everything that goes into it, so the table narrows
// to the finished dish on the far right.
//
// It's written as an outline with the finished dish first. Each line is a step,
// indented under it are the ingredients and earlier steps it uses, and a line
// with nothing under it is an ingredient:
//
//	serve with rice
//	  stir together, simmer 5 min
//	    simmer until soft, 20 min
//	      1 cup red lentils, rinsed
//	      3 cups water
//	    fry until the seeds pop
//	      2 tbsp ghee
//	      1 tsp cumin seeds
//
// A block can have more than one top-level line, for parts served separately.
type recipeNode struct {
	text     string
	indent   int
	children []*recipeNode
}

// depth is how many steps deep a node is: 0 for an ingredient, and one more
// than its deepest input for a step. It's also the column the node's cell starts in.
func (n *recipeNode) depth() int {
	d := 0
	for _, c := range n.children {
		d = max(d, c.depth()+1)
	}
	return d
}

// leaves counts the ingredients under a node, which is how many rows its cell spans.
func (n *recipeNode) leaves() int {
	if len(n.children) == 0 {
		return 1
	}
	count := 0
	for _, c := range n.children {
		count += c.leaves()
	}
	return count
}

func parseRecipe(src string) ([]*recipeNode, error) {
	var roots, stack []*recipeNode
	for line := range strings.Lines(src) {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		indent := 0
		for _, r := range line {
			if r == ' ' {
				indent++
			} else if r == '\t' {
				indent += 4
			} else {
				break
			}
		}
		n := &recipeNode{text: text, indent: indent}
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, n)
		} else {
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, n)
		}
		stack = append(stack, n)
	}
	if len(roots) == 0 {
		return nil, errors.New("recipe block is empty")
	}
	return roots, nil
}

// renderRecipe turns a ```recipe block into a table. Each ingredient is a row;
// a step's cell spans the rows of the ingredients under it, and stretches across
// columns when its inputs are fewer steps deep than its neighbours'.
func renderRecipe(src string) (string, error) {
	roots, err := parseRecipe(src)
	if err != nil {
		return "", err
	}
	last := 0 // the rightmost column
	for _, r := range roots {
		last = max(last, r.depth())
	}

	// rows holds each row's cells in column order. A cell goes in the row of
	// the first ingredient under it, after the cells of its inputs, which sit
	// in columns to its left.
	var rows [][]string
	var place func(n *recipeNode, end int, top bool) int
	place = func(n *recipeNode, end int, top bool) int {
		start := n.depth()
		class := "step"
		if len(n.children) == 0 {
			class = "ingredient"
		}
		if top && len(n.children) > 0 {
			class += " dish"
		}
		cell := "<td class=\"" + class + "\""
		if span := n.leaves(); span > 1 {
			cell += fmt.Sprintf(" rowspan=\"%d\"", span)
		}
		if span := end - start + 1; span > 1 {
			cell += fmt.Sprintf(" colspan=\"%d\"", span)
		}
		cell += ">" + html.EscapeString(n.text) + "</td>"

		if len(n.children) == 0 {
			rows = append(rows, []string{cell})
			return len(rows) - 1
		}
		first := -1
		for _, c := range n.children {
			row := place(c, start-1, false)
			if first < 0 {
				first = row
			}
		}
		rows[first] = append(rows[first], cell)
		return first
	}
	for _, r := range roots {
		place(r, last, true)
	}

	var b strings.Builder
	b.WriteString("<div class=\"recipe-table\">\n<table>\n<tbody>\n")
	for _, cells := range rows {
		b.WriteString("<tr>" + strings.Join(cells, "") + "</tr>\n")
	}
	b.WriteString("</tbody>\n</table>\n</div>\n")
	return b.String(), nil
}
