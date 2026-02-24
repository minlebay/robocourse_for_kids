-- Promote the seed admin user (created in 000002) to the administrator role
-- and force a password change on first login.
-- If migration 000002 was not applied (e.g. fresh test DB), insert the user directly.

UPDATE users
SET role = 'administrator', must_change_password = true
WHERE id = '00000000-0000-0000-0000-000000000001';

-- Fallback: seed user does not exist — create the administrator from scratch.
-- Password: admin12 (same as migration 000002). Must be changed on first login.
INSERT INTO users (id, login, password_hash, name, role, must_change_password)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'admin',
  '$2a$10$7qdpyMuYJqjT6/1stY3pT.O30y0/isel/OSFp1epLJbL4idrrkz6e',
  'Администратор',
  'administrator',
  true
)
ON CONFLICT DO NOTHING;
