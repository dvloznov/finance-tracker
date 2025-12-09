package pipeline

import (
	"context"
	"fmt"
	"strings"

	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
)

// CategoryValidator validates transaction categories against the taxonomy.
type CategoryValidator struct {
	categories    map[string]bool              // Set of valid category names
	subcategories map[string]map[string]bool   // Map of category -> set of valid subcategories
}

// NewCategoryValidator creates a validator from the categories taxonomy.
func NewCategoryValidator(ctx context.Context, repo CategoryRepository) (*CategoryValidator, error) {
	rows, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewCategoryValidator: list categories: %w", err)
	}

	validator := &CategoryValidator{
		categories:    make(map[string]bool),
		subcategories: make(map[string]map[string]bool),
	}

	// Build lookup maps
	for _, row := range rows {
		normalizedName := normalizeCategory(row.Name)
		
		if row.Depth == 1 {
			// Parent category
			validator.categories[normalizedName] = true
			if validator.subcategories[normalizedName] == nil {
				validator.subcategories[normalizedName] = make(map[string]bool)
			}
		} else if row.Depth == 2 && row.ParentCategoryID.Valid {
			// Subcategory - need to find parent name
			parentName := findParentName(rows, row.ParentCategoryID.StringVal)
			if parentName != "" {
				normalizedParent := normalizeCategory(parentName)
				if validator.subcategories[normalizedParent] == nil {
					validator.subcategories[normalizedParent] = make(map[string]bool)
				}
				validator.subcategories[normalizedParent][normalizedName] = true
			}
		}
	}

	return validator, nil
}

// ValidateCategory checks if a category and subcategory are valid.
// Returns nil if valid, error if invalid.
func (v *CategoryValidator) ValidateCategory(category, subcategory string) error {
	normCat := normalizeCategory(category)
	normSubcat := normalizeCategory(subcategory)

	if !v.categories[normCat] {
		return fmt.Errorf("invalid category: %q (normalized: %q)", category, normCat)
	}

	if subcats, ok := v.subcategories[normCat]; ok {
		if !subcats[normSubcat] {
			validSubs := make([]string, 0, len(subcats))
			for s := range subcats {
				validSubs = append(validSubs, s)
			}
			return fmt.Errorf("invalid subcategory %q for category %q. Valid subcategories: %v", 
				subcategory, category, validSubs)
		}
	}

	return nil
}

// normalizeCategory normalizes a category name for comparison.
// Converts to uppercase and trims whitespace for case-insensitive comparison.
func normalizeCategory(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

// findParentName finds the parent category name by ID.
func findParentName(rows []infra.CategoryRow, parentID string) string {
	for _, row := range rows {
		if row.CategoryID == parentID {
			return row.Name
		}
	}
	return ""
}
