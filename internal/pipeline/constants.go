package pipeline

import "os"

// Default values for document processing and parsing.
const (
	// DefaultUserID is a placeholder until real authentication is implemented.
	// TODO: Replace with user ID extracted from auth context.
	DefaultUserID = "denis"

	// DefaultModelName is the default Gemini model used for parsing.
	DefaultModelName = "gemini-2.5-flash"

	// DefaultCategoryID is used when categorization is deferred.
	DefaultCategoryID = "cat_uncategorized_other"

	// DefaultCategoryName and DefaultSubcategoryName align with the default category ID.
	DefaultCategoryName    = "Uncategorized"
	DefaultSubcategoryName = "Other"
)

// DefaultBankName is the institution name used in the statement-parsing prompt.
// Override with the BANK_NAME environment variable.
var DefaultBankName = func() string {
	if v := os.Getenv("BANK_NAME"); v != "" {
		return v
	}
	return "Barclays UK"
}()

// DefaultSourceSystem is the default institution identifier stored in BigQuery.
// Override with the SOURCE_SYSTEM environment variable.
var DefaultSourceSystem = func() string {
	if v := os.Getenv("SOURCE_SYSTEM"); v != "" {
		return v
	}
	return "BARCLAYS"
}()
