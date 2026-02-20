DROP INDEX IF EXISTS users_email_unique;
ALTER TABLE users
  DROP COLUMN IF EXISTS is_blocked,
  DROP COLUMN IF EXISTS must_change_password,
  DROP COLUMN IF EXISTS email;
