-- Create documents table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.documents` (
  document_id           STRING NOT NULL,
  user_id               STRING,
  gcs_uri               STRING NOT NULL,
  document_type         STRING NOT NULL,
  source_system         STRING,
  institution_id        STRING,
  account_id            STRING,
  statement_start_date  DATE,
  statement_end_date    DATE,
  upload_ts             TIMESTAMP NOT NULL,
  processed_ts          TIMESTAMP,
  parsing_status        STRING,
  original_filename     STRING,
  file_mime_type        STRING,
  text_gcs_uri          STRING,
  checksum_sha256       STRING,
  metadata              JSON
);
