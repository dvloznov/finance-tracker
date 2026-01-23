-- Create institutions table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.institutions` (
  institution_id  STRING NOT NULL,
  name            STRING NOT NULL,
  country_code    STRING,
  logo_url        STRING,
  metadata        JSON,
  created_ts      TIMESTAMP,
  updated_ts      TIMESTAMP
);

-- Create accounts table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.accounts` (
  account_id      STRING NOT NULL,
  user_id         STRING,
  institution_id  STRING,
  account_name    STRING,
  account_number  STRING,
  sort_code       STRING,
  iban            STRING,
  currency        STRING,
  account_type    STRING
);
