-- Create schema_migrations table to track applied migrations
CREATE TABLE IF NOT EXISTS `studious-union-470122-v7.finance.schema_migrations` (
  version       INT64 NOT NULL,
  name          STRING NOT NULL,
  applied_at    TIMESTAMP NOT NULL,
  checksum      STRING,
  applied_by    STRING
);
