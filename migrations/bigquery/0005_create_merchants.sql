-- Create merchants table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.merchants` (
  merchant_id    STRING NOT NULL,
  canonical_name STRING NOT NULL,
  display_name   STRING,
  website_domain STRING,
  mcc_code       STRING,
  country        STRING,
  city           STRING,
  metadata       JSON,
  created_ts     TIMESTAMP
);
