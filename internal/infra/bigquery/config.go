package bigquery

import "os"

// projectID and datasetID are the single source of truth for GCP project and
// BigQuery dataset configuration. Set BQ_PROJECT_ID and BQ_DATASET_ID env vars
// to override the defaults (useful for local dev vs. staging vs. production).
var (
	projectID = getEnvOrDefault("BQ_PROJECT_ID", "studious-union-470122-v7")
	datasetID = getEnvOrDefault("BQ_DATASET_ID", "finance")
)

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
