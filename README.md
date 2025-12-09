# Finance Tracker

Backend service for a Notion dashboard that processes bank statements and receipts using Gemini AI. State is stored in Google Cloud (BigQuery + Cloud Storage).

## What it does

Uploads PDFs → Gemini AI extracts transactions → Stores in BigQuery → Powers Notion dashboard

## Usage

**Upload a PDF:**
```bash
go run cmd/upload-pdf/main.go -bucket BUCKET -file statement.pdf
```

**Process it:**
```bash
go run cmd/ingest/main.go -gcs-uri gs://bucket/statement.pdf
```

## Tech Stack

- **Go 1.24.2**
- **Google Cloud Storage** - PDF storage
- **Google BigQuery** - Transaction database
- **Gemini 2.5 Flash** - AI extraction & categorization

## BigQuery Schema

- `documents` - Uploaded PDFs metadata
- `transactions` - Extracted transactions with categories
- `parsing_runs` - Processing status tracking
- `model_outputs` - Raw AI responses
- `categories` - Hierarchical transaction taxonomy

## Setup

1. GCP project with BigQuery & Storage enabled
2. `gcloud auth application-default login`
3. Create `finance` dataset in BigQuery
