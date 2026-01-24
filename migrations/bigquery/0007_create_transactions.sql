-- Create transactions table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.transactions` (
  transaction_id         STRING NOT NULL,
  user_id                STRING,
  account_id             STRING,
  institution_id         STRING,
  document_id            STRING,
  parsing_run_id         STRING,
  transaction_date       DATE NOT NULL,
  amount                 NUMERIC NOT NULL,
  currency               STRING NOT NULL,
  balance_after          NUMERIC,
  direction              STRING,
  raw_description        STRING NOT NULL,
  category_id            STRING,
  created_ts             TIMESTAMP NOT NULL
);
