-- Seed initial categories (denormalized: one row per category-subcategory combination)
INSERT INTO `{{PROJECT_ID}}.{{DATASET_ID}}.categories` 
  (category_id, category_name, subcategory_name, slug, is_active, created_ts)
VALUES
  -- Income categories
  ('cat_income_salary', 'Income', 'Salary', 'income-salary', TRUE, CURRENT_TIMESTAMP()),
  ('cat_income_freelance', 'Income', 'Freelance', 'income-freelance', TRUE, CURRENT_TIMESTAMP()),
  ('cat_income_investment', 'Income', 'Investment Income', 'income-investment', TRUE, CURRENT_TIMESTAMP()),
  
  -- Housing categories
  ('cat_housing_rent', 'Housing', 'Rent/Mortgage', 'housing-rent', TRUE, CURRENT_TIMESTAMP()),
  ('cat_housing_utilities', 'Housing', 'Utilities', 'housing-utilities', TRUE, CURRENT_TIMESTAMP()),
  ('cat_housing_maintenance', 'Housing', 'Maintenance', 'housing-maintenance', TRUE, CURRENT_TIMESTAMP()),
  
  -- Transportation categories
  ('cat_transport_transit', 'Transportation', 'Public Transit', 'transport-transit', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transport_fuel', 'Transportation', 'Fuel', 'transport-fuel', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transport_parking', 'Transportation', 'Parking', 'transport-parking', TRUE, CURRENT_TIMESTAMP()),
  
  -- Food & Dining categories
  ('cat_food_groceries', 'Food & Dining', 'Groceries', 'food-groceries', TRUE, CURRENT_TIMESTAMP()),
  ('cat_food_restaurants', 'Food & Dining', 'Restaurants', 'food-restaurants', TRUE, CURRENT_TIMESTAMP()),
  ('cat_food_coffee', 'Food & Dining', 'Coffee Shops', 'food-coffee', TRUE, CURRENT_TIMESTAMP()),
  
  -- Shopping categories
  ('cat_shopping_clothing', 'Shopping', 'Clothing', 'shopping-clothing', TRUE, CURRENT_TIMESTAMP()),
  ('cat_shopping_electronics', 'Shopping', 'Electronics', 'shopping-electronics', TRUE, CURRENT_TIMESTAMP()),
  ('cat_shopping_home', 'Shopping', 'Home Goods', 'shopping-home', TRUE, CURRENT_TIMESTAMP()),
  
  -- Parent-only categories (no subcategory)
  ('cat_healthcare', 'Healthcare', NULL, 'healthcare', TRUE, CURRENT_TIMESTAMP()),
  ('cat_entertainment', 'Entertainment', NULL, 'entertainment', TRUE, CURRENT_TIMESTAMP()),
  ('cat_travel', 'Travel', NULL, 'travel', TRUE, CURRENT_TIMESTAMP()),
  ('cat_subscriptions', 'Subscriptions', NULL, 'subscriptions', TRUE, CURRENT_TIMESTAMP()),
  ('cat_transfers', 'Transfers', NULL, 'transfers', TRUE, CURRENT_TIMESTAMP()),
  ('cat_uncategorized', 'Uncategorized', NULL, 'uncategorized', TRUE, CURRENT_TIMESTAMP());
