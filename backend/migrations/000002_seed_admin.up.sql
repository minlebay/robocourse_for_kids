-- Предопределённый пользователь-администратор для первого входа.
-- Логин: admin, пароль: admin12. При первом входе потребуется сменить пароль.

INSERT INTO users (id, login, password_hash, name, role, must_change_password)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'admin',
  '$2a$10$7qdpyMuYJqjT6/1stY3pT.O30y0/isel/OSFp1epLJbL4idrrkz6e',
  'Администратор',
  'administrator',
  true
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_roles (user_id, role)
VALUES ('00000000-0000-0000-0000-000000000001', 'administrator')
ON CONFLICT (user_id, role) DO NOTHING;
