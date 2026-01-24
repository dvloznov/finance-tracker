-- Create institutions table
CREATE TABLE IF NOT EXISTS `finance.institutions` (
  institution_id  STRING NOT NULL,
  name            STRING NOT NULL,
  created_ts      TIMESTAMP,
  updated_ts      TIMESTAMP
);

-- Create accounts table
CREATE TABLE IF NOT EXISTS `finance.accounts` (
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
