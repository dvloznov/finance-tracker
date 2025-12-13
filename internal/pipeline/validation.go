package pipeline

import (
	"context"
	"fmt"
	"strings"

	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
)

// CategoryValidator validates transaction categories against the taxonomy.
type CategoryValidator struct {
	// Map of "CATEGORY|SUBCATEGORY" or "CATEGORY|" -> category_id
	validPairs   map[string]string
	categoryRows []infra.CategoryRow // Keep for other lookups if needed
}

// NewCategoryValidator creates a validator from the categories taxonomy.
func NewCategoryValidator(ctx context.Context, repo CategoryRepository) (*CategoryValidator, error) {
	rows, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewCategoryValidator: list categories: %w", err)
	}

	validator := &CategoryValidator{
		validPairs:   make(map[string]string),
		categoryRows: rows,
	}

	// Build lookup map of valid category-subcategory pairs
	for _, row := range rows {
		normCat := normalizeCategory(row.CategoryName)
		normSubcat := ""
		if row.SubcategoryName.Valid && row.SubcategoryName.StringVal != "" {
			normSubcat = normalizeCategory(row.SubcategoryName.StringVal)
		}

		// Store as "CATEGORY|SUBCATEGORY" or "CATEGORY|"
		key := normCat + "|" + normSubcat
		validator.validPairs[key] = row.CategoryID
	}

	return validator, nil
}

// ValidateCategory checks if a category and subcategory are valid.
// Returns the category_id if valid, error if invalid.
func (v *CategoryValidator) ValidateCategory(category, subcategory string) (string, error) {
	normCat := normalizeCategory(category)
	normSubcat := normalizeCategory(subcategory)

	// Try exact match first
	key := normCat + "|" + normSubcat
	if categoryID, ok := v.validPairs[key]; ok {
		return categoryID, nil
	}

	// If subcategory was provided but not found, try without it
	if normSubcat != "" {
		keyWithoutSub := normCat + "|"
		if categoryID, ok := v.validPairs[keyWithoutSub]; ok {
			return categoryID, nil
		}
	}

	return "", fmt.Errorf("invalid category/subcategory combination: %q / %q", category, subcategory)
}

// normalizeCategory normalizes a category name for comparison.
// Converts to uppercase and trims whitespace for case-insensitive comparison.
func normalizeCategory(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}
