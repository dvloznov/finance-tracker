-- Create merchants table
CREATE TABLE IF NOT EXISTS `finance.merchants` (
  merchant_id       STRING NOT NULL,
  merchant_name     STRING NOT NULL,
  normalized_name   STRING NOT NULL,
  category_id       STRING NOT NULL,
  created_ts        TIMESTAMP NOT NULL
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS `finance.transactions` (
  transaction_id         STRING NOT NULL,
  user_id                STRING,
  account_id             STRING,
  institution_id         STRING,
  document_id            STRING,
  parsing_run_id         STRING,
  transaction_date       DATE NOT NULL,
  statement_date         DATE NOT NULL,
  transaction_type       STRING,
  amount                 NUMERIC NOT NULL,
  currency               STRING NOT NULL,
  balance_after          NUMERIC,
  direction              STRING,
  raw_description        STRING NOT NULL,
  merchant_id            STRING NOT NULL,
  created_ts             TIMESTAMP NOT NULL
);
