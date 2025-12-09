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
	// Setup test categories
	categories := []infra.CategoryRow{
		{CategoryID: "cat1", Name: "HOUSING", Depth: 1},
		{CategoryID: "cat1-sub1", Name: "Rent", Depth: 2, ParentCategoryID: bigquerylib.NullString{StringVal: "cat1", Valid: true}},
		{CategoryID: "cat1-sub2", Name: "Utilities", Depth: 2, ParentCategoryID: bigquerylib.NullString{StringVal: "cat1", Valid: true}},
		{CategoryID: "cat2", Name: "FOOD", Depth: 1},
		{CategoryID: "cat2-sub1", Name: "Groceries", Depth: 2, ParentCategoryID: bigquerylib.NullString{StringVal: "cat2", Valid: true}},
		{CategoryID: "cat2-sub2", Name: "Restaurants", Depth: 2, ParentCategoryID: bigquerylib.NullString{StringVal: "cat2", Valid: true}},
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
			category:    "  FOOD  ",
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
			category:    "FOOD",
			subcategory: "Restaurants",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCategory(tt.category, tt.subcategory)
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
