-- Create receipt_line_items table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.receipt_line_items` (
  line_item_id     STRING NOT NULL,
  receipt_id       STRING NOT NULL,
  line_index       INT64,
  description      STRING NOT NULL,
  quantity         NUMERIC,
  unit_price       NUMERIC,
  total_price      NUMERIC,
  category_id      STRING,
  subcategory_id   STRING,
  category_name    STRING,
  subcategory_name STRING,
  sku              STRING,
  metadata         JSON
);
