-- Create parsing_runs table
CREATE TABLE IF NOT EXISTS `studious-union-470122-v7.finance.parsing_runs` (
  parsing_run_id STRING NOT NULL,
  document_id    STRING NOT NULL,
  started_ts     TIMESTAMP NOT NULL,
  finished_ts    TIMESTAMP,
  parser_type    STRING,
  parser_version STRING,
  status         STRING,
  error_message  STRING,
  tokens_input   INT64,
  tokens_output  INT64,
  metadata       JSON
);
