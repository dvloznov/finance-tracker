package pipeline

import (
	"context"
	"fmt"
	"strings"
)

// buildCategoriesPromptWithRepo constructs a prompt string containing all active categories
// and subcategories from BigQuery, formatted for LLM consumption.
func buildCategoriesPromptWithRepo(ctx context.Context, repo CategoryRepository) (string, error) {
	rows, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return "", fmt.Errorf("buildCategoriesPrompt: list categories: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("buildCategoriesPrompt: no active categories found")
	}

	// Separate parents and children.
	type parentInfo struct {
		ID   string
		Name string
	}
	var parentsOrder []parentInfo
	childrenByParent := make(map[string][]string)

	for _, r := range rows {
		if r.Depth == 1 {
			parentsOrder = append(parentsOrder, parentInfo{ID: r.CategoryID, Name: r.Name})
			if _, ok := childrenByParent[r.CategoryID]; !ok {
				childrenByParent[r.CategoryID] = []string{}
			}
		}
	}

	for _, r := range rows {
		if r.Depth == 2 && r.ParentCategoryID.Valid {
			parentID := r.ParentCategoryID.StringVal
			childrenByParent[parentID] = append(childrenByParent[parentID], r.Name)
		}
	}

	var b strings.Builder
	b.WriteString("Use ONLY the following Categories and Subcategories:\n\n")

	for _, p := range parentsOrder {
		b.WriteString(p.Name + ":\n")
		subs := childrenByParent[p.ID]
		if len(subs) == 0 {
			// no subcategories defined – still list a placeholder so the model knows.
			b.WriteString("  - Other\n\n")
			continue
		}
		for _, s := range subs {
			b.WriteString("  - " + s + "\n")
		}
		b.WriteString("\n")
	}

	// Additionally, constrain what the model is allowed to output.
	b.WriteString("Category must be exactly one of the category names shown above.\n")
	b.WriteString("Subcategory must be exactly one of the subcategory names listed under that category.\n")
	b.WriteString("If you are unsure, default to category \"OTHER\" with subcategory \"Other\" if it exists.\n")

	return b.String(), nil
}
