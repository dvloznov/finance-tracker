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

	// Group by category name
	categoryMap := make(map[string][]string)
	for _, row := range rows {
		cat := row.CategoryName
		if row.SubcategoryName.Valid && row.SubcategoryName.StringVal != "" {
			categoryMap[cat] = append(categoryMap[cat], row.SubcategoryName.StringVal)
		} else {
			// Ensure category exists even with no subcategory
			if _, exists := categoryMap[cat]; !exists {
				categoryMap[cat] = []string{}
			}
		}
	}

	var b strings.Builder
	b.WriteString("Use ONLY the following Categories and Subcategories:\n\n")

	for cat, subs := range categoryMap {
		b.WriteString(cat + ":\n")
		if len(subs) == 0 {
			b.WriteString("  (no subcategories)\n\n")
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
	b.WriteString("If you are unsure, use category \"Uncategorized\" with an empty subcategory.\n")

	return b.String(), nil
}
