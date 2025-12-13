-- Create categories table (denormalized: one row per category-subcategory pair)
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.categories` (
  category_id        STRING NOT NULL,
  category_name      STRING NOT NULL,
  subcategory_name   STRING,
  slug               STRING NOT NULL,
  description        STRING,
  is_active          BOOL,
  created_ts         TIMESTAMP,
  retired_ts         TIMESTAMP,
  metadata           JSON
);
