# Finance Tracker

A Go-based financial statement processing system that automatically extracts, categorizes, and stores transaction data from bank statements using AI.

## Overview

Finance Tracker is an intelligent document processing pipeline that:
1. Uploads bank statement PDFs to Google Cloud Storage (GCS)
2. Parses PDFs using Google's Gemini AI model to extract transaction data
3. Automatically categorizes transactions using a predefined taxonomy
4. Stores structured transaction data in Google BigQuery for analysis

## Architecture

The system consists of two main command-line tools and a processing pipeline:

### Commands

#### 1. `upload-pdf`
Uploads a local PDF file to Google Cloud Storage.

**Usage:**
```bash
go run cmd/upload-pdf/main.go \
  -bucket BUCKET_NAME \
  -file /path/to/statement.pdf \
  -object optional-object-name
```

**Parameters:**
- `-bucket`: GCS bucket name (required)
- `-file`: Path to local PDF file (required)
- `-object`: GCS object name (optional, defaults to filename)

#### 2. `ingest`
Processes a bank statement PDF stored in GCS through the full ingestion pipeline.

**Usage:**
```bash
go run cmd/ingest/main.go \
  -gcs-uri gs://bucket/path/to/statement.pdf
```

**Parameters:**
- `-gcs-uri`: Full GCS URI of the statement PDF (required)

### Processing Pipeline

The ingestion pipeline (`internal/pipeline/pipeline.go`) orchestrates the following steps:

1. **Document Registration**: Creates a document record in BigQuery with metadata
2. **Parsing Run Initialization**: Starts a new parsing run with status tracking
3. **PDF Retrieval**: Downloads the PDF from Google Cloud Storage
4. **AI-Powered Parsing**: Uses Google Gemini 2.5 Flash to:
   - Extract all transactions from the statement
   - Parse transaction details (date, description, amount, balance)
   - Automatically categorize each transaction
5. **Model Output Storage**: Stores raw AI output for audit/debugging
6. **Transaction Transformation**: Normalizes parsed data into structured format
7. **Data Persistence**: Inserts transactions into BigQuery
8. **Status Update**: Marks the parsing run as successful or failed

## Data Model

### BigQuery Schema

The system uses Google BigQuery with the following main tables:

#### `finance.documents`
Stores metadata about uploaded documents:
- Document ID, user ID, GCS URI
- Document type (e.g., BANK_STATEMENT)
- Source system (e.g., BARCLAYS)
- Upload and processing timestamps
- Parsing status

#### `finance.transactions`
Stores normalized transaction data:
- Transaction ID, date, amount, currency
- Account information
- Raw and normalized descriptions
- Category and subcategory
- Balance after transaction
- Direction (IN/OUT)
- Merchant information

#### `finance.parsing_runs`
Tracks each parsing attempt:
- Run ID, document ID
- Status (RUNNING, SUCCESS, FAILED)
- Start and end timestamps
- Error messages if failed

#### `finance.model_outputs`
Stores raw AI model outputs:
- Model name and version
- Raw JSON response
- Extracted text
- Metadata

#### `finance.categories`
Hierarchical transaction categories:
- Parent and child categories
- Category names and descriptions
- Active/inactive status

## Technology Stack

- **Language**: Go 1.24.2
- **Cloud Platform**: Google Cloud Platform
  - Cloud Storage (for PDF storage)
  - BigQuery (for data warehouse)
  - Vertex AI / Gemini API (for AI parsing)
- **AI Model**: Google Gemini 2.5 Flash
- **Key Dependencies**:
  - `cloud.google.com/go/bigquery`
  - `cloud.google.com/go/storage`
  - `google.golang.org/genai`

## How It Works

### Transaction Extraction

The system uses Gemini AI with a structured prompt that:
- Identifies all transactions in the PDF
- Extracts key fields: date, description, amount, currency, balance
- Handles different statement formats (separate credit/debit columns)
- Automatically classifies transactions into predefined categories

### Category System

Transactions are categorized using a hierarchical taxonomy stored in BigQuery:
- **Level 1**: Parent categories (e.g., GROCERIES, TRANSPORT, ENTERTAINMENT)
- **Level 2**: Subcategories (e.g., Supermarkets, Fuel, Streaming Services)

The AI model is instructed to use only predefined categories to ensure consistency.

### Transaction Normalization

Each transaction is normalized to include:
- **Signed Amount**: Positive for money IN, negative for money OUT
- **Direction**: IN or OUT based on amount sign
- **ISO Date Format**: YYYY-MM-DD
- **Category/Subcategory**: From predefined taxonomy
- **Optional Fields**: Account name/number, balance after transaction

## Current Capabilities

- ✅ Parses Barclays UK bank statements
- ✅ Extracts transaction data automatically
- ✅ Categorizes transactions using AI
- ✅ Stores data in BigQuery for analysis
- ✅ Tracks parsing runs and status
- ✅ Handles both credit and debit transactions
- ✅ Maintains running balance tracking

## Future Enhancements

The codebase includes placeholders for:
- Multi-bank support (currently hardcoded for Barclays)
- Multi-user support (currently hardcoded user "denis")
- Account mapping and reconciliation
- Merchant identification and tracking
- Receipt processing (tables exist but not implemented)
- Split transactions
- Internal transfers detection

## Setup Requirements

1. **Google Cloud Project**: Active GCP project with billing enabled
2. **Authentication**: Application Default Credentials configured
   ```bash
   gcloud auth application-default login
   ```
3. **BigQuery Dataset**: `finance` dataset with required tables
4. **GCS Bucket**: For storing statement PDFs
5. **Gemini API**: Access to Gemini AI models

## Project Structure

```
finance-tracker/
├── cmd/
│   ├── ingest/          # Main ingestion CLI
│   └── upload-pdf/      # PDF upload utility
├── internal/
│   ├── gcsuploader/     # GCS upload helpers
│   ├── infra/
│   │   └── bigquery/    # BigQuery data access layer
│   └── pipeline/        # Core processing pipeline
├── go.mod               # Go module definition
└── go.sum               # Dependency checksums
```

## Usage Example

### Complete Workflow

1. **Upload a bank statement**:
```bash
go run cmd/upload-pdf/main.go \
  -bucket my-finance-bucket \
  -file ~/Downloads/barclays-statement-jan-2024.pdf
```

2. **Process the statement**:
```bash
go run cmd/ingest/main.go \
  -gcs-uri gs://my-finance-bucket/barclays-statement-jan-2024.pdf
```

3. **Query results in BigQuery**:
```sql
SELECT 
  transaction_date,
  raw_description,
  amount,
  category_name,
  subcategory_name
FROM finance.transactions
WHERE document_id = 'your-document-id'
ORDER BY transaction_date;
```

## Error Handling

The pipeline includes comprehensive error handling:
- Failed parsing runs are tracked with error messages
- Raw model outputs are stored for debugging
- Each step is atomic with proper status tracking
- Context timeouts prevent hanging operations (5-minute default)

## Data Privacy

- User ID is currently hardcoded (intended for single-user or development)
- All financial data is stored in private BigQuery tables
- GCS buckets should be configured with appropriate access controls
- Sensitive data is only processed within Google Cloud infrastructure
