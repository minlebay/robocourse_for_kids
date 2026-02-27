-- Source of truth for roles is user_roles; users.role is deprecated (kept for backward compat, nullable).
-- New user creation no longer writes to users.role; API derives "primary" role from user_roles.
ALTER TABLE users ALTER COLUMN role DROP NOT NULL;
ALTER TABLE users ALTER COLUMN role SET DEFAULT NULL;
