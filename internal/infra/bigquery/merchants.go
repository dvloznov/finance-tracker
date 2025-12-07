package bigquery

import "cloud.google.com/go/bigquery"

type MerchantRow struct {
	MerchantID    string `bigquery:"merchant_id"`    // REQUIRED
	CanonicalName string `bigquery:"canonical_name"` // REQUIRED

	DisplayName   string `bigquery:"display_name"`   // NULLABLE
	WebsiteDomain string `bigquery:"website_domain"` // NULLABLE
	MCCCode       string `bigquery:"mcc_code"`       // NULLABLE
	Country       string `bigquery:"country"`        // NULLABLE
	City          string `bigquery:"city"`           // NULLABLE

	Metadata  bigquery.NullJSON      `bigquery:"metadata"`   // NULLABLE (JSON)
	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"` // NULLABLE (default CURRENT_TIMESTAMP())
}
