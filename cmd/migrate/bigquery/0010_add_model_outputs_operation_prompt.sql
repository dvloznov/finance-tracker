-- Add operation and prompt columns to model_outputs for full audit trail
ALTER TABLE `finance.model_outputs`
ADD COLUMN operation STRING,
ADD COLUMN prompt STRING;
