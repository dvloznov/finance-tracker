-- Create receipts table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.receipts` (
  receipt_id            STRING NOT NULL,
  user_id               STRING,
  document_id           STRING NOT NULL,
  parsing_run_id        STRING,
  merchant_id           STRING,
  merchant_name         STRING,
  purchase_datetime     DATETIME,
  purchase_date         DATE,
  total_amount          NUMERIC NOT NULL,
  subtotal_amount       NUMERIC,
  tax_amount            NUMERIC,
  tip_amount            NUMERIC,
  currency              STRING NOT NULL,
  payment_method        STRING,
  card_last4            STRING,
  linked_transaction_id STRING,
  created_ts            TIMESTAMP NOT NULL,
  updated_ts            TIMESTAMP,
  metadata              JSON
);
