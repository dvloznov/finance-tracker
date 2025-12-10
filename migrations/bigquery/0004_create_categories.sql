-- Create categories table
CREATE TABLE IF NOT EXISTS `{{PROJECT_ID}}.{{DATASET_ID}}.categories` (
  category_id        STRING NOT NULL,
  parent_category_id STRING,
  depth              INT64 NOT NULL,
  slug               STRING NOT NULL,
  name               STRING NOT NULL,
  description        STRING,
  is_active          BOOL,
  created_ts         TIMESTAMP,
  retired_ts         TIMESTAMP,
  metadata           JSON
);
