package pipeline

import (
	"context"
	"testing"

	bigquerylib "cloud.google.com/go/bigquery"
	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
)

// mockCategoryRepository is a mock for testing category validation
type mockCategoryRepository struct {
	categories []infra.CategoryRow
}

func (m *mockCategoryRepository) ListActiveCategories(ctx context.Context) ([]infra.CategoryRow, error) {
	return m.categories, nil
}

func TestCategoryValidator_ValidateCategory(t *testing.T) {
	// Setup test categories (denormalized)
	categories := []infra.CategoryRow{
		{CategoryID: "cat1-sub1", CategoryName: "Housing", SubcategoryName: bigquerylib.NullString{StringVal: "Rent", Valid: true}},
		{CategoryID: "cat1-sub2", CategoryName: "Housing", SubcategoryName: bigquerylib.NullString{StringVal: "Utilities", Valid: true}},
		{CategoryID: "cat2-sub1", CategoryName: "Food & Dining", SubcategoryName: bigquerylib.NullString{StringVal: "Groceries", Valid: true}},
		{CategoryID: "cat2-sub2", CategoryName: "Food & Dining", SubcategoryName: bigquerylib.NullString{StringVal: "Restaurants", Valid: true}},
		{CategoryID: "cat_healthcare", CategoryName: "Healthcare", SubcategoryName: bigquerylib.NullString{Valid: false}},
	}

	repo := &mockCategoryRepository{categories: categories}
	validator, err := NewCategoryValidator(context.Background(), repo)
	if err != nil {
		t.Fatalf("NewCategoryValidator failed: %v", err)
	}

	tests := []struct {
		name        string
		category    string
		subcategory string
		wantErr     bool
	}{
		{
			name:        "valid category and subcategory",
			category:    "HOUSING",
			subcategory: "Rent",
			wantErr:     false,
		},
		{
			name:        "valid with different case",
			category:    "housing",
			subcategory: "rent",
			wantErr:     false,
		},
		{
			name:        "valid with extra spaces",
			category:    "  Food & Dining  ",
			subcategory: "  Groceries  ",
			wantErr:     false,
		},
		{
			name:        "invalid category",
			category:    "INVALID",
			subcategory: "Rent",
			wantErr:     true,
		},
		{
			name:        "invalid subcategory for valid category",
			category:    "HOUSING",
			subcategory: "Groceries",
			wantErr:     true,
		},
		{
			name:        "valid food and restaurants",
			category:    "Food & Dining",
			subcategory: "Restaurants",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateCategory(tt.category, tt.subcategory)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HOUSING", "HOUSING"},
		{"housing", "HOUSING"},
		{"  Housing  ", "HOUSING"},
		{"FoOd", "FOOD"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeCategory(tt.input)
			if got != tt.want {
				t.Errorf("normalizeCategory(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
