-- Seed initial categories (denormalized: one row per category-subcategory combination)
INSERT INTO `finance.categories`
  (category_id, category_name, subcategory_name, slug)
VALUES
  -- Uncategorized
  ('cat_uncategorized_cash_atm', 'Uncategorized', 'Cash & ATM', 'uncategorized-cash-atm'),
  ('cat_uncategorized_check', 'Uncategorized', 'Check', 'uncategorized-check'),
  ('cat_uncategorized_other', 'Uncategorized', 'Other', 'uncategorized-other'),

  -- Entertainment
  ('cat_entertainment_arts', 'Entertainment', 'Arts', 'entertainment-arts'),
  ('cat_entertainment_music', 'Entertainment', 'Music', 'entertainment-music'),
  ('cat_entertainment_dating', 'Entertainment', 'Dating', 'entertainment-dating'),
  ('cat_entertainment_movies_dvds', 'Entertainment', 'Movies & DVDs', 'entertainment-movies-dvds'),
  ('cat_entertainment_newspaper_magazines', 'Entertainment', 'Newspaper & Magazines', 'entertainment-newspaper-magazines'),
  ('cat_entertainment_social_club', 'Entertainment', 'Social Club', 'entertainment-social-club'),
  ('cat_entertainment_sport', 'Entertainment', 'Sport', 'entertainment-sport'),
  ('cat_entertainment_games', 'Entertainment', 'Games', 'entertainment-games'),

  -- Education
  ('cat_education_tuition', 'Education', 'Tuition', 'education-tuition'),
  ('cat_education_student_loan', 'Education', 'Student Loan', 'education-student-loan'),
  ('cat_education_books_supplies', 'Education', 'Books & Supplies', 'education-books-supplies'),

  -- Shopping
  ('cat_shopping_pets', 'Shopping', 'Pets', 'shopping-pets'),
  ('cat_shopping_groceries', 'Shopping', 'Groceries', 'shopping-groceries'),
  ('cat_shopping_general', 'Shopping', 'General', 'shopping-general'),
  ('cat_shopping_clothing', 'Shopping', 'Clothing', 'shopping-clothing'),
  ('cat_shopping_home', 'Shopping', 'Home', 'shopping-home'),
  ('cat_shopping_books', 'Shopping', 'Books', 'shopping-books'),
  ('cat_shopping_electronics_software', 'Shopping', 'Electronics & Software', 'shopping-electronics-software'),
  ('cat_shopping_hobbies', 'Shopping', 'Hobbies', 'shopping-hobbies'),
  ('cat_shopping_sporting_goods', 'Shopping', 'Sporting Goods', 'shopping-sporting-goods'),

  -- Personal Care
  ('cat_personal_care_hair', 'Personal Care', 'Hair', 'personal-care-hair'),
  ('cat_personal_care_laundry', 'Personal Care', 'Laundry', 'personal-care-laundry'),
  ('cat_personal_care_beauty', 'Personal Care', 'Beauty', 'personal-care-beauty'),
  ('cat_personal_care_spa_massage', 'Personal Care', 'Spa & Massage', 'personal-care-spa-massage'),

  -- Health & Fitness
  ('cat_health_fitness_dentist', 'Health & Fitness', 'Dentist', 'health-fitness-dentist'),
  ('cat_health_fitness_doctor', 'Health & Fitness', 'Doctor', 'health-fitness-doctor'),
  ('cat_health_fitness_eye_care', 'Health & Fitness', 'Eye care', 'health-fitness-eye-care'),
  ('cat_health_fitness_pharmacy', 'Health & Fitness', 'Pharmacy', 'health-fitness-pharmacy'),
  ('cat_health_fitness_gym', 'Health & Fitness', 'Gym', 'health-fitness-gym'),
  ('cat_health_fitness_pets', 'Health & Fitness', 'Pets', 'health-fitness-pets'),
  ('cat_health_fitness_sports', 'Health & Fitness', 'Sports', 'health-fitness-sports'),

  -- Food & Dining
  ('cat_food_dining_catering', 'Food & Dining', 'Catering', 'food-dining-catering'),
  ('cat_food_dining_coffee_shops', 'Food & Dining', 'Coffee shops', 'food-dining-coffee-shops'),
  ('cat_food_dining_delivery', 'Food & Dining', 'Delivery', 'food-dining-delivery'),
  ('cat_food_dining_fast_food', 'Food & Dining', 'Fast Food', 'food-dining-fast-food'),
  ('cat_food_dining_restaurants', 'Food & Dining', 'Restaurants', 'food-dining-restaurants'),
  ('cat_food_dining_bars', 'Food & Dining', 'Bars', 'food-dining-bars'),

  -- Gifts & Donations
  ('cat_gifts_donations_gift', 'Gifts & Donations', 'Gift', 'gifts-donations-gift'),
  ('cat_gifts_donations_charity', 'Gifts & Donations', 'Charity', 'gifts-donations-charity'),

  -- Investments
  ('cat_investments_equities', 'Investments', 'Equities', 'investments-equities'),
  ('cat_investments_bonds', 'Investments', 'Bonds', 'investments-bonds'),
  ('cat_investments_bank_products', 'Investments', 'Bank products', 'investments-bank-products'),
  ('cat_investments_retirement', 'Investments', 'Retirement', 'investments-retirement'),
  ('cat_investments_annuities', 'Investments', 'Annuities', 'investments-annuities'),
  ('cat_investments_real_estate', 'Investments', 'Real-estate', 'investments-real-estate'),

  -- Bills & Utilities
  ('cat_bills_utilities_television', 'Bills & Utilities', 'Television', 'bills-utilities-television'),
  ('cat_bills_utilities_home_phone', 'Bills & Utilities', 'Home Phone', 'bills-utilities-home-phone'),
  ('cat_bills_utilities_internet', 'Bills & Utilities', 'Internet', 'bills-utilities-internet'),
  ('cat_bills_utilities_mobile_phone', 'Bills & Utilities', 'Mobile Phone', 'bills-utilities-mobile-phone'),
  ('cat_bills_utilities_utilities', 'Bills & Utilities', 'Utilities', 'bills-utilities-utilities'),

  -- Auto & Transport
  ('cat_auto_transport_auto_insurance', 'Auto & Transport', 'Auto Insurance', 'auto-transport-auto-insurance'),
  ('cat_auto_transport_auto_payment', 'Auto & Transport', 'Auto Payment', 'auto-transport-auto-payment'),
  ('cat_auto_transport_parking', 'Auto & Transport', 'Parking', 'auto-transport-parking'),
  ('cat_auto_transport_public_transport', 'Auto & Transport', 'Public transport', 'auto-transport-public-transport'),
  ('cat_auto_transport_service_auto_parts', 'Auto & Transport', 'Service & Auto Parts', 'auto-transport-service-auto-parts'),
  ('cat_auto_transport_taxi', 'Auto & Transport', 'Taxi', 'auto-transport-taxi'),
  ('cat_auto_transport_gas_fuel', 'Auto & Transport', 'Gas & Fuel', 'auto-transport-gas-fuel'),

  -- Travel
  ('cat_travel_air_travel', 'Travel', 'Air Travel', 'travel-air-travel'),
  ('cat_travel_hotel', 'Travel', 'Hotel', 'travel-hotel'),
  ('cat_travel_rental_car_taxi', 'Travel', 'Rental Car & Taxi', 'travel-rental-car-taxi'),
  ('cat_travel_vacation', 'Travel', 'Vacation', 'travel-vacation'),

  -- Fees & Charges
  ('cat_fees_charges_service_fee', 'Fees & Charges', 'Service Fee', 'fees-charges-service-fee'),
  ('cat_fees_charges_late_fee', 'Fees & Charges', 'Late Fee', 'fees-charges-late-fee'),
  ('cat_fees_charges_finance_charge', 'Fees & Charges', 'Finance Charge', 'fees-charges-finance-charge'),
  ('cat_fees_charges_atm_fee', 'Fees & Charges', 'ATM Fee', 'fees-charges-atm-fee'),
  ('cat_fees_charges_bank_fee', 'Fees & Charges', 'Bank Fee', 'fees-charges-bank-fee'),
  ('cat_fees_charges_commissions', 'Fees & Charges', 'Commissions', 'fees-charges-commissions'),

  -- Business Services
  ('cat_business_services_advertising', 'Business Services', 'Advertising', 'business-services-advertising'),
  ('cat_business_services_financial_services', 'Business Services', 'Financial Services', 'business-services-financial-services'),
  ('cat_business_services_office_supplies', 'Business Services', 'Office Supplies', 'business-services-office-supplies'),
  ('cat_business_services_printing', 'Business Services', 'Printing', 'business-services-printing'),
  ('cat_business_services_shipping', 'Business Services', 'Shipping', 'business-services-shipping'),
  ('cat_business_services_legal', 'Business Services', 'Legal', 'business-services-legal'),

  -- Personal Services
  ('cat_personal_services_advisory_consulting', 'Personal Services', 'Advisory and Consulting', 'personal-services-advisory-and-consulting'),
  ('cat_personal_services_financial_services', 'Personal Services', 'Financial Services', 'personal-services-financial-services'),
  ('cat_personal_services_lawyer', 'Personal Services', 'Lawyer', 'personal-services-lawyer'),
  ('cat_personal_services_repairs_maintenance', 'Personal Services', 'Repairs & Maintenance', 'personal-services-repairs-maintenance'),

  -- Taxes
  ('cat_taxes_federal_tax', 'Taxes', 'Federal Tax', 'taxes-federal-tax'),
  ('cat_taxes_state_tax', 'Taxes', 'State Tax', 'taxes-state-tax'),
  ('cat_taxes_local_tax', 'Taxes', 'Local Tax', 'taxes-local-tax'),
  ('cat_taxes_sales_tax', 'Taxes', 'Sales Tax', 'taxes-sales-tax'),
  ('cat_taxes_property_tax', 'Taxes', 'Property Tax', 'taxes-property-tax'),

  -- Gambling
  ('cat_gambling_betting', 'Gambling', 'Betting', 'gambling-betting'),
  ('cat_gambling_lottery', 'Gambling', 'Lottery', 'gambling-lottery'),
  ('cat_gambling_casino', 'Gambling', 'Casino', 'gambling-casino'),

  -- Home
  ('cat_home_rent', 'Home', 'Rent', 'home-rent'),
  ('cat_home_mortgage', 'Home', 'Mortgage', 'home-mortgage'),
  ('cat_home_secured_loans', 'Home', 'Secured loans', 'home-secured-loans'),

  -- Pensions and Insurances
  ('cat_pensions_insurances_pension_payments', 'Pensions and Insurances', 'Pension payments', 'pensions-insurances-pension-payments'),
  ('cat_pensions_insurances_life_insurance', 'Pensions and Insurances', 'Life insurance', 'pensions-insurances-life-insurance'),
  ('cat_pensions_insurances_buildings_contents_insurance', 'Pensions and Insurances', 'Buildings and contents insurance', 'pensions-insurances-buildings-contents-insurance'),
  ('cat_pensions_insurances_health_insurance', 'Pensions and Insurances', 'Health insurance', 'pensions-insurances-health-insurance');
