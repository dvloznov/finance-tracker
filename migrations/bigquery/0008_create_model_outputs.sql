-- Create model_outputs table
CREATE TABLE IF NOT EXISTS `studious-union-470122-v7.finance.model_outputs` (
  output_id      STRING NOT NULL,
  parsing_run_id STRING NOT NULL,
  document_id    STRING NOT NULL,
  model_name     STRING NOT NULL,
  model_version  STRING,
  raw_json       JSON NOT NULL,
  extracted_text STRING,
  created_ts     TIMESTAMP NOT NULL,
  notes          STRING,
  metadata       JSON
);
