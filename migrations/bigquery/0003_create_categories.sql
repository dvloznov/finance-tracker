-- Create categories table (denormalized: one row per category-subcategory pair)
CREATE TABLE IF NOT EXISTS `{{DATASET_ID}}.categories` (
  category_id        STRING NOT NULL,
  category_name      STRING NOT NULL,
  subcategory_name   STRING,
  slug               STRING NOT NULL
);
