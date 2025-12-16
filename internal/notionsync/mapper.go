package notionsync

import (
	"fmt"
	"time"

	"github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/jomei/notionapi"
)

// AccountToNotionProperties converts a BigQuery AccountRow to Notion properties.
// Maps fields according to the NOTION_SETUP.md specification for Accounts database.
func AccountToNotionProperties(acc *bigquery.AccountRow) notionapi.Properties {
	props := notionapi.Properties{
		"Account ID": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: acc.AccountID,
					},
				},
			},
		},
	}

	// Account Name
	if acc.AccountName != "" {
		props["Account Name"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: acc.AccountName,
					},
				},
			},
		}
	}

	// Institution
	if acc.InstitutionID != "" {
		props["Institution"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: acc.InstitutionID,
			},
		}
	}

	// Account Type
	if acc.AccountType != "" {
		props["Account Type"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: acc.AccountType,
			},
		}
	}

	// Currency
	if acc.Currency != "" {
		props["Currency"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: acc.Currency,
			},
		}
	}

	// Account Number
	if acc.AccountNumber != "" {
		props["Account Number"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: acc.AccountNumber,
					},
				},
			},
		}
	}

	// IBAN
	if acc.IBAN != "" {
		props["IBAN"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: acc.IBAN,
					},
				},
			},
		}
	}

	// Is Primary
	if acc.IsPrimary.Valid {
		props["Is Primary"] = notionapi.CheckboxProperty{
			Checkbox: acc.IsPrimary.Bool,
		}
	}

	// Opened Date
	if acc.OpenedDate.Valid {
		props["Opened Date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: func() *notionapi.Date {
					d := notionapi.Date(time.Date(
						acc.OpenedDate.Date.Year,
						time.Month(acc.OpenedDate.Date.Month),
						acc.OpenedDate.Date.Day,
						0, 0, 0, 0, time.UTC,
					))
					return &d
				}(),
			},
		}
	}

	// Closed Date
	if acc.ClosedDate.Valid {
		props["Closed Date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: func() *notionapi.Date {
					d := notionapi.Date(time.Date(
						acc.ClosedDate.Date.Year,
						time.Month(acc.ClosedDate.Date.Month),
						acc.ClosedDate.Date.Day,
						0, 0, 0, 0, time.UTC,
					))
					return &d
				}(),
			},
		}
	}

	return props
}

// CategoryToNotionProperties converts a BigQuery CategoryRow to Notion properties.
// Maps fields according to the NOTION_SETUP.md specification for Categories database.
func CategoryToNotionProperties(cat *bigquery.CategoryRow) notionapi.Properties {
	// Build title based on whether subcategory exists
	titleText := cat.CategoryName
	if cat.SubcategoryName.Valid && cat.SubcategoryName.StringVal != "" {
		titleText = cat.CategoryName + " → " + cat.SubcategoryName.StringVal
	}

	props := notionapi.Properties{
		"Category": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: titleText,
					},
				},
			},
		},
		"Slug": notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: cat.Slug,
					},
				},
			},
		},
	}

	// Category Name (parent category)
	props["Category Name"] = notionapi.RichTextProperty{
		RichText: []notionapi.RichText{
			{
				Type: notionapi.ObjectTypeText,
				Text: &notionapi.Text{
					Content: cat.CategoryName,
				},
			},
		},
	}

	// Subcategory Name (if exists)
	if cat.SubcategoryName.Valid && cat.SubcategoryName.StringVal != "" {
		props["Subcategory Name"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: cat.SubcategoryName.StringVal,
					},
				},
			},
		}
	}

	// Description
	if cat.Description.Valid {
		props["Description"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: cat.Description.StringVal,
					},
				},
			},
		}
	}

	// Is Active
	if cat.IsActive.Valid {
		props["Is Active"] = notionapi.CheckboxProperty{
			Checkbox: cat.IsActive.Bool,
		}
	} else {
		// Default to true if not specified
		props["Is Active"] = notionapi.CheckboxProperty{
			Checkbox: true,
		}
	}

	// Retired date
	if cat.RetiredTS.Valid {
		props["Retired"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: (*notionapi.Date)(&cat.RetiredTS.Timestamp),
			},
		}
	}

	return props
}

// DocumentToNotionProperties converts a BigQuery DocumentRow to Notion properties.
// Maps fields according to the NOTION_SETUP.md specification for Documents database.
func DocumentToNotionProperties(doc *bigquery.DocumentRow) notionapi.Properties {
	props := notionapi.Properties{
		"Document ID": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: doc.DocumentID,
					},
				},
			},
		},
		"Document Type": notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: doc.DocumentType,
			},
		},
		"Upload Date": notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: func() *notionapi.Date {
					d := notionapi.Date(doc.UploadTS)
					return &d
				}(),
			},
		},
	}

	// Original Filename
	if doc.OriginalFilename != "" {
		props["Original Filename"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: doc.OriginalFilename,
					},
				},
			},
		}
	}

	// Statement Start Date
	if doc.StatementStartDate.Valid {
		props["Statement Start"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: func() *notionapi.Date {
					d := notionapi.Date(time.Date(
						doc.StatementStartDate.Date.Year,
						time.Month(doc.StatementStartDate.Date.Month),
						doc.StatementStartDate.Date.Day,
						0, 0, 0, 0, time.UTC,
					))
					return &d
				}(),
			},
		}
	}

	// Statement End Date
	if doc.StatementEndDate.Valid {
		props["Statement End"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: func() *notionapi.Date {
					d := notionapi.Date(time.Date(
						doc.StatementEndDate.Date.Year,
						time.Month(doc.StatementEndDate.Date.Month),
						doc.StatementEndDate.Date.Day,
						0, 0, 0, 0, time.UTC,
					))
					return &d
				}(),
			},
		}
	}

	// Statement Period (formatted)
	if doc.StatementStartDate.Valid && doc.StatementEndDate.Valid {
		startTime := time.Date(
			doc.StatementStartDate.Date.Year,
			time.Month(doc.StatementStartDate.Date.Month),
			doc.StatementStartDate.Date.Day,
			0, 0, 0, 0, time.UTC,
		)
		endTime := time.Date(
			doc.StatementEndDate.Date.Year,
			time.Month(doc.StatementEndDate.Date.Month),
			doc.StatementEndDate.Date.Day,
			0, 0, 0, 0, time.UTC,
		)
		period := fmt.Sprintf("%s - %s",
			startTime.Format("Jan 2006"),
			endTime.Format("Jan 2006"))
		props["Statement Period"] = notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: period,
					},
				},
			},
		}
	}

	// Processing Status
	if doc.ParsingStatus != "" {
		props["Processing Status"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: doc.ParsingStatus,
			},
		}
	}

	// Processed Date
	if doc.ProcessedTS.Valid {
		props["Processed Date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: (*notionapi.Date)(&doc.ProcessedTS.Timestamp),
			},
		}
	}

	// File Type
	if doc.FileMimeType != "" {
		props["File Type"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: doc.FileMimeType,
			},
		}
	}

	// GCS Link
	if doc.GCSURI != "" {
		props["GCS Link"] = notionapi.URLProperty{
			URL: doc.GCSURI,
		}
	}

	// Note: Account relation will need to be handled separately
	// as it requires looking up the Notion page ID of the account

	return props
}

// TransactionToNotionProperties converts a BigQuery TransactionRow to Notion properties.
// Maps fields according to the Notion transaction database schema:
// Description, Date, Amount, Currency, Balance After, Account, Category, Subcategory,
// Source Document, Parsing Run ID, Document ID, Transaction ID, Imported At, Notes, Is Corrected
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
		"Transaction ID": notionapi.RichTextProperty{
			RichText: []notionapi.RichText{
				{
					Type: notionapi.ObjectTypeText,
					Text: &notionapi.Text{
						Content: tx.TransactionID,
					},
				},
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

	// Category - use CategoryName if available
	if tx.CategoryName.Valid && tx.CategoryName.StringVal != "" {
		props["Category"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: tx.CategoryName.StringVal,
			},
		}
	}

	// Subcategory - use SubcategoryName if available
	if tx.SubcategoryName.Valid && tx.SubcategoryName.StringVal != "" {
		props["Subcategory"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: tx.SubcategoryName.StringVal,
			},
		}
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

// TransactionToNotionPropertiesWithCategories converts a BigQuery TransactionRow to Notion properties
// with category relations. If categoryPageIDs is provided and the transaction has a category_id,
// it will create a relation to the Categories database instead of a Select property.
func TransactionToNotionPropertiesWithCategories(tx *bigquery.TransactionRow, categoryPageIDs map[string]string) notionapi.Properties {
	// Start with the base properties
	props := TransactionToNotionProperties(tx)

	// Override Category with relation if we have category_id and the mapping
	if tx.CategoryID.Valid && tx.CategoryID.StringVal != "" && categoryPageIDs != nil {
		if pageID, ok := categoryPageIDs[tx.CategoryID.StringVal]; ok {
			props["Category"] = notionapi.RelationProperty{
				Relation: []notionapi.Relation{
					{
						ID: notionapi.PageID(pageID),
					},
				},
			}
		}
	}

	// Remove Subcategory property since we're using denormalized categories
	delete(props, "Subcategory")

	return props
}
