-- Create parsing_runs table
CREATE TABLE IF NOT EXISTS `parsing_runs` (
  parsing_run_id STRING NOT NULL,
  document_id    STRING NOT NULL,
  started_ts     TIMESTAMP NOT NULL,
  finished_ts    TIMESTAMP,
  parser_type    STRING,
  parser_version STRING,
  status         STRING,
  error_message  STRING
);
