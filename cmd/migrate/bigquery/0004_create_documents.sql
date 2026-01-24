-- Create documents table
CREATE TABLE IF NOT EXISTS `documents` (
  document_id           STRING NOT NULL,
  user_id               STRING,
  gcs_uri               STRING NOT NULL,
  institution_id        STRING,
  account_id            STRING,
  upload_ts             TIMESTAMP NOT NULL,
  parsing_status        STRING,
  original_filename     STRING,
  file_mime_type        STRING
);
