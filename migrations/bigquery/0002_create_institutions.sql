-- Create institutions table
CREATE TABLE IF NOT EXISTS `studious-union-470122-v7.finance.institutions` (
  institution_id STRING NOT NULL,
  name           STRING NOT NULL,
  type           STRING,
  country        STRING,
  metadata       JSON,
  created_ts     TIMESTAMP
);
