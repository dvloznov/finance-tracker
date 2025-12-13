-- Seed initial categories
-- Top-level categories (depth: 0)
INSERT INTO `{{PROJECT_ID}}.{{DATASET_ID}}.categories` 
  (category_id, parent_category_id, depth, slug, name, is_active, created_ts)
VALUES
  ('cat_income', NULL, 0, 'income', 'Income', TRUE, CURRENT_TIMESTAMP()),
  ('cat_housing', NULL, 0, 'housing', 'Housing', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transport', NULL, 0, 'transportation', 'Transportation', TRUE, CURRENT_TIMESTAMP()),
  ('cat_food', NULL, 0, 'food-dining', 'Food & Dining', TRUE, CURRENT_TIMESTAMP()),
  ('cat_shopping', NULL, 0, 'shopping', 'Shopping', TRUE, CURRENT_TIMESTAMP()),
  ('cat_healthcare', NULL, 0, 'healthcare', 'Healthcare', TRUE, CURRENT_TIMESTAMP()),
  ('cat_entertainment', NULL, 0, 'entertainment', 'Entertainment', TRUE, CURRENT_TIMESTAMP()),
  ('cat_travel', NULL, 0, 'travel', 'Travel', TRUE, CURRENT_TIMESTAMP()),
  ('cat_subscriptions', NULL, 0, 'subscriptions', 'Subscriptions', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transfers', NULL, 0, 'transfers', 'Transfers', TRUE, CURRENT_TIMESTAMP()),
  ('cat_uncategorized', NULL, 0, 'uncategorized', 'Uncategorized', TRUE, CURRENT_TIMESTAMP());

-- Subcategories (depth: 1)
INSERT INTO `{{PROJECT_ID}}.{{DATASET_ID}}.categories`
  (category_id, parent_category_id, depth, slug, name, is_active, created_ts)
VALUES
  -- Income subcategories
  ('cat_income_salary', 'cat_income', 1, 'salary', 'Salary', TRUE, CURRENT_TIMESTAMP()),
  ('cat_income_freelance', 'cat_income', 1, 'freelance', 'Freelance', TRUE, CURRENT_TIMESTAMP()),
  ('cat_income_investment', 'cat_income', 1, 'investment-income', 'Investment Income', TRUE, CURRENT_TIMESTAMP()),
  
  -- Housing subcategories
  ('cat_housing_rent', 'cat_housing', 1, 'rent-mortgage', 'Rent/Mortgage', TRUE, CURRENT_TIMESTAMP()),
  ('cat_housing_utilities', 'cat_housing', 1, 'utilities', 'Utilities', TRUE, CURRENT_TIMESTAMP()),
  ('cat_housing_maintenance', 'cat_housing', 1, 'maintenance', 'Maintenance', TRUE, CURRENT_TIMESTAMP()),
  
  -- Transportation subcategories
  ('cat_transport_transit', 'cat_transport', 1, 'public-transit', 'Public Transit', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transport_fuel', 'cat_transport', 1, 'fuel', 'Fuel', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transport_parking', 'cat_transport', 1, 'parking', 'Parking', TRUE, CURRENT_TIMESTAMP()),
  
  -- Food & Dining subcategories
  ('cat_food_groceries', 'cat_food', 1, 'groceries', 'Groceries', TRUE, CURRENT_TIMESTAMP()),
  ('cat_food_restaurants', 'cat_food', 1, 'restaurants', 'Restaurants', TRUE, CURRENT_TIMESTAMP()),
  ('cat_food_coffee', 'cat_food', 1, 'coffee-shops', 'Coffee Shops', TRUE, CURRENT_TIMESTAMP()),
  
  -- Shopping subcategories
  ('cat_shopping_clothing', 'cat_shopping', 1, 'clothing', 'Clothing', TRUE, CURRENT_TIMESTAMP()),
  ('cat_shopping_electronics', 'cat_shopping', 1, 'electronics', 'Electronics', TRUE, CURRENT_TIMESTAMP()),
  ('cat_shopping_home', 'cat_shopping', 1, 'home-goods', 'Home Goods', TRUE, CURRENT_TIMESTAMP());
