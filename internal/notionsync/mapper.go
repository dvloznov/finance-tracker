package notionsync

import (
	"fmt"
	"time"

	"github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/jomei/notionapi"
)

// TransactionToNotionProperties converts a BigQuery TransactionRow to Notion properties.
// Maps fields according to the Notion transaction database schema:
// Description, Date, Amount, Currency, Balance After, Account, Category, Subcategory,
// Source Document, Parsing Run ID, Document ID, Imported At, Notes, Is Corrected
func TransactionToNotionProperties(tx *bigquery.TransactionRow) notionapi.Properties {
	props := notionapi.Properties{
		"Description": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: tx.RawDescription,
					},
				},
			},
		},
		"Date": notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: func() *notionapi.Date {
					d := notionapi.Date(time.Date(
						tx.TransactionDate.Year,
						tx.TransactionDate.Month,
						tx.TransactionDate.Day,
						0, 0, 0, 0, time.UTC,
					))
					return &d
				}(),
			},
		},
		"Amount": notionapi.NumberProperty{
			Number: func() float64 {
				if tx.Amount != nil {
					f, _ := tx.Amount.Float64()
					return f
				}
				return 0
			}(),
		},
		"Currency": notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: func() string {
					if tx.Currency != "" {
						return tx.Currency
					}
					return "GBP"
				}(),
			},
		},
	}

	// Balance After (nullable)
	if tx.BalanceAfter != nil {
		props["Balance After"] = notionapi.NumberProperty{
			Number: func() float64 {
				f, _ := tx.BalanceAfter.Float64()
				return f
			}(),
		}
	}

	// Account - use AccountID if available
	if tx.AccountID != "" {
		props["Account"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: tx.AccountID,
					},
				},
			},
		}
	}

	// Category - leave empty for now (it's a relation property)
	// TODO: Implement proper relation handling when category database is set up

	// Subcategory - use SubcategoryName if available, default to 'Other'
	props["Subcategory"] = notionapi.SelectProperty{
		Select: notionapi.Option{
			Name: func() string {
				if tx.SubcategoryName.Valid && tx.SubcategoryName.StringVal != "" {
					return tx.SubcategoryName.StringVal
				}
				return "Other"
			}(),
		},
	}

	// Parsing Run ID
	if tx.ParsingRunID != "" {
		props["Parsing Run ID"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: tx.ParsingRunID,
					},
				},
			},
		}
	}

	// Document ID
	if tx.DocumentID != "" {
		props["Document ID"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: tx.DocumentID,
					},
				},
			},
		}
	}

	// Imported At - use CreatedTS
	props["Imported At"] = notionapi.DateProperty{
		Date: &notionapi.DateObject{
			Start: (*notionapi.Date)(&tx.CreatedTS),
		},
	}

	// Notes - use NormalizedDescription if available
	if tx.NormalizedDescription.Valid {
		props["Notes"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: tx.NormalizedDescription.StringVal,
					},
				},
			},
		}
	}

	// Is Corrected - we don't have this field in TransactionRow, default to false
	props["Is Corrected"] = notionapi.CheckboxProperty{
		Checkbox: false,
	}

	return props
}

// GetNotionPageIDFromTransaction extracts the Notion page ID from the external_reference field.
// Returns empty string if not set.
func GetNotionPageIDFromTransaction(tx *bigquery.TransactionRow) string {
	if tx.ExternalReference.Valid {
		return tx.ExternalReference.StringVal
	}
	return ""
}

// SetNotionPageIDOnTransaction creates a formatted external_reference string for storing Notion page ID.
func SetNotionPageIDOnTransaction(pageID string) string {
	return fmt.Sprintf("notion:%s", pageID)
}
