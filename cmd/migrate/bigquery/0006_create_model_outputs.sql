-- Create model_outputs table
CREATE TABLE IF NOT EXISTS `model_outputs` (
  output_id      STRING NOT NULL,
  parsing_run_id STRING NOT NULL,
  document_id    STRING NOT NULL,
  model_name     STRING NOT NULL,
  raw_json       JSON NOT NULL,
  created_ts     TIMESTAMP NOT NULL,
);
