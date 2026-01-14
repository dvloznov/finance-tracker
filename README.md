# Finance Tracker

Full-stack application that processes bank statements and receipts using Gemini AI. Backend is built with Go, frontend with Next.js. State is stored in Google Cloud (BigQuery + Cloud Storage).

## What it does

Uploads PDFs → Gemini AI extracts transactions → Stores in BigQuery → View & manage through web UI

## Usage

**Initialize BigQuery tables:**
```bash
go run cmd/migrate/main.go -project studious-union-470122-v7
```

**Upload a PDF:**
```bash
go run cmd/upload-pdf/main.go -bucket BUCKET -file statement.pdf
```

**Process it:**
```bash
go run cmd/ingest/main.go -gcs-uri gs://bucket/statement.pdf
```

## Tech Stack

**Backend:**
- **Go 1.24.2**
- **Google Cloud Storage** - PDF storage
- **Google BigQuery** - Transaction database
- **Gemini 2.5 Flash** - AI extraction & categorization

**Frontend:**
- **Next.js 16** - React framework
- **TypeScript** - Type-safe development
- **TanStack Query** - Data fetching & caching
- **TanStack Table** - Advanced table features
- **Nivo** - Interactive charts & visualizations
- **Tailwind CSS** - Styling

## BigQuery Schema

The schema is managed through versioned SQL migrations in `migrations/bigquery/`:

- `schema_migrations` - Migration version tracking
- `institutions` - Financial institutions
- `accounts` - User accounts
- `categories` - Hierarchical transaction taxonomy
- `merchants` - Merchant information
- `documents` - Uploaded PDFs metadata
- `parsing_runs` - Processing status tracking
- `model_outputs` - Raw AI responses
- `transactions` - Extracted transactions with categories
- `receipts` - Receipt data
- `receipt_line_items` - Individual line items from receipts

## Setup

**Backend:**
1. GCP project with BigQuery & Storage enabled
2. `gcloud auth application-default login`
3. Create `finance` dataset in BigQuery
4. Run migrations: `go run cmd/migrate/main.go -project YOUR_PROJECT_ID`

**Frontend:**
```bash
cd web
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to access the web UI.

## Database Migrations

The project uses a migration system to manage BigQuery table schemas:

```bash
# Apply all pending migrations (project ID is required)
go run cmd/migrate/main.go -project studious-union-470122-v7

# Use custom project/dataset
go run cmd/migrate/main.go -project my-project -dataset my-dataset
```

Migrations are SQL files in `migrations/bigquery/` with format `NNNN_description.sql`.
The tool tracks applied migrations in the `schema_migrations` table and only applies new ones.

## Web UI

The frontend provides three main views:

- **Dashboard** - Financial overview with charts showing income/expenses trends, spending by category, and monthly comparisons
- **Documents** - Upload and manage PDFs (bank statements, receipts), trigger parsing jobs, view processing status
- **Transactions** - Browse, search, and filter all extracted transactions with sortable columns
