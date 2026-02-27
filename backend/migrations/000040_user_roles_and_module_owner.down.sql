DROP INDEX IF EXISTS idx_modules_owner_id;
ALTER TABLE modules DROP COLUMN IF EXISTS owner_id;

DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP TABLE IF EXISTS user_roles;
