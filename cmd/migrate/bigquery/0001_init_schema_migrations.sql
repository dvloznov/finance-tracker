-- Create schema_migrations table to track applied migrations
CREATE TABLE IF NOT EXISTS `finance.schema_migrations` (
  version       INT64 NOT NULL,
  name          STRING NOT NULL,
  applied_at    TIMESTAMP NOT NULL,
  checksum      STRING,
  applied_by    STRING
);
