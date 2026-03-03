ALTER TABLE `finance.documents`
ADD COLUMN IF NOT EXISTS statement_start_date DATE,
ADD COLUMN IF NOT EXISTS statement_end_date DATE;
