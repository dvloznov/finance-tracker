-- Create institutions table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.institutions` (
  institution_id STRING NOT NULL,
  name           STRING NOT NULL,
  type           STRING,
  country        STRING,
  metadata       JSON,
  created_ts     TIMESTAMP
);
