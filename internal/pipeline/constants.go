package pipeline

// Default values for document processing and parsing.
// These can be overridden via configuration or environment variables in the future.
const (
	// DefaultUserID is the default user identifier for documents and transactions.
	DefaultUserID = "denis"

	// DefaultSourceSystem is the default source system for documents.
	DefaultSourceSystem = "BARCLAYS"

	// DefaultDocumentType is the default document type for uploaded files.
	DefaultDocumentType = "BANK_STATEMENT"

	// DefaultModelName is the default Gemini model used for parsing.
	DefaultModelName = "gemini-2.5-flash"
)
