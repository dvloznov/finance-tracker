package pipeline

import (
	"context"
	"fmt"
	"strings"

	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
)

// CategoryValidator validates transaction categories against the taxonomy.
type CategoryValidator struct {
	categories     map[string]bool            // Set of valid category names
	subcategories  map[string]map[string]bool // Map of category -> set of valid subcategories
	subcatToParent map[string]string          // Map of subcategory name -> parent category name
}

// NewCategoryValidator creates a validator from the categories taxonomy.
func NewCategoryValidator(ctx context.Context, repo CategoryRepository) (*CategoryValidator, error) {
	rows, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewCategoryValidator: list categories: %w", err)
	}

	validator := &CategoryValidator{
		categories:     make(map[string]bool),
		subcategories:  make(map[string]map[string]bool),
		subcatToParent: make(map[string]string),
	}

	// Build lookup maps
	for _, row := range rows {
		normalizedName := normalizeCategory(row.Name)

		if row.Depth == 0 {
			// Top-level category
			validator.categories[normalizedName] = true
			if validator.subcategories[normalizedName] == nil {
				validator.subcategories[normalizedName] = make(map[string]bool)
			}
		} else if row.Depth == 1 && row.ParentCategoryID.Valid {
			// Subcategory - need to find parent name
			parentName := findParentName(rows, row.ParentCategoryID.StringVal)
			if parentName != "" {
				normalizedParent := normalizeCategory(parentName)
				if validator.subcategories[normalizedParent] == nil {
					validator.subcategories[normalizedParent] = make(map[string]bool)
				}
				validator.subcategories[normalizedParent][normalizedName] = true
				validator.subcatToParent[normalizedName] = normalizedParent
			}
		}
	}

	return validator, nil
}

// ValidateCategory checks if a category and subcategory are valid.
// Returns nil if valid, error if invalid.
// If category is actually a subcategory name, it treats it as such.
func (v *CategoryValidator) ValidateCategory(category, subcategory string) error {
	normCat := normalizeCategory(category)
	normSubcat := normalizeCategory(subcategory)

	// Check if what was provided as "category" is actually a subcategory
	if parentCat, isSubcat := v.subcatToParent[normCat]; isSubcat {
		// The category field contains a subcategory name
		// This is valid - Gemini sometimes returns subcategory as the main category
		// Ignore subcategory field if it's empty or the same as category
		if normSubcat == "" || normSubcat == normCat {
			return nil
		}
		return fmt.Errorf("conflicting category/subcategory: category=%q is a subcategory of %q, but subcategory=%q was also provided",
			category, parentCat, subcategory)
	}

	// Normal case: category is a top-level category
	if !v.categories[normCat] {
		return fmt.Errorf("invalid category: %q (normalized: %q)", category, normCat)
	}

	// If the category has subcategories defined
	if subcats, ok := v.subcategories[normCat]; ok && len(subcats) > 0 {
		// Allow empty subcategory (means top-level category was selected)
		if normSubcat == "" {
			return nil
		}

		// Otherwise, validate the subcategory
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
