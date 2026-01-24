-- Seed initial categories (denormalized: one row per category-subcategory combination)
INSERT INTO `finance.categories`
  (category_id, category_name, subcategory_name, slug)
VALUES
  -- Income categories
  ('cat_income_salary', 'Income', 'Salary', 'income-salary'),
  ('cat_income_freelance', 'Income', 'Freelance', 'income-freelance'),
  ('cat_income_investment', 'Income', 'Investment Income', 'income-investment'),
  
  -- Housing categories
  ('cat_housing_rent', 'Housing', 'Rent/Mortgage', 'housing-rent'),
  ('cat_housing_utilities', 'Housing', 'Utilities', 'housing-utilities'),
  ('cat_housing_maintenance', 'Housing', 'Maintenance', 'housing-maintenance'),
  
  -- Transportation categories
  ('cat_transport_transit', 'Transportation', 'Public Transit', 'transport-transit'),
  ('cat_transport_fuel', 'Transportation', 'Fuel', 'transport-fuel'),
  ('cat_transport_parking', 'Transportation', 'Parking', 'transport-parking'),
  
  -- Food & Dining categories
  ('cat_food_groceries', 'Food & Dining', 'Groceries', 'food-groceries'),
  ('cat_food_restaurants', 'Food & Dining', 'Restaurants', 'food-restaurants'),
  ('cat_food_coffee', 'Food & Dining', 'Coffee Shops', 'food-coffee'),
  
  -- Shopping categories
  ('cat_shopping_clothing', 'Shopping', 'Clothing', 'shopping-clothing'),
  ('cat_shopping_electronics', 'Shopping', 'Electronics', 'shopping-electronics'),
  ('cat_shopping_home', 'Shopping', 'Home Goods', 'shopping-home'),
  
  -- Parent-only categories (no subcategory)
  ('cat_healthcare', 'Healthcare', NULL, 'healthcare'),
  ('cat_entertainment', 'Entertainment', NULL, 'entertainment'),
  ('cat_travel', 'Travel', NULL, 'travel'),
  ('cat_subscriptions', 'Subscriptions', NULL, 'subscriptions'),
  ('cat_transfers', 'Transfers', NULL, 'transfers'),
  ('cat_uncategorized', 'Uncategorized', NULL, 'uncategorized');
