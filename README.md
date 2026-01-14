# Finance Tracker

AI-powered financial document processor. Upload bank statements/receipts → Gemini extracts transactions → Browse via web UI.

## Quick Start

**1. Setup**
```bash
# Authenticate with GCP
gcloud auth application-default login

# Initialize BigQuery tables
go run cmd/migrate/main.go -project YOUR_PROJECT_ID

# Install frontend dependencies
cd web && npm install && cd ..
```

**2. Run (both backend + frontend)**
```bash
./start-dev.sh
```

This starts:
- API server on [http://localhost:8080](http://localhost:8080)
- Frontend on [http://localhost:3000](http://localhost:3000)

Press Ctrl+C to stop both services.

**Or run separately:**
```bash
# Terminal 1 - API
go run cmd/api/main.go -port 8080 -bucket YOUR_BUCKET_NAME

# Terminal 2 - Frontend
cd web && npm run dev
```

## CLI Usage

Build the CLI:
```bash
go build -o cli cmd/cli/main.go
```

**Upload a PDF to GCS:**
```bash
./cli upload -bucket BUCKET_NAME -file statement.pdf
```

**Parse a document:**
```bash
./cli ingest -gcs-uri gs://bucket/statement.pdf
```

**Re-parse existing document:**
```bash
./cli reparse -id DOCUMENT_ID
```

**Inspect document details:**
```bash
./cli inspect -id DOCUMENT_ID
```

## Tech Stack

Go 1.24 · Next.js 16 · BigQuery · Cloud Storage · Gemini 2.5 Flash
