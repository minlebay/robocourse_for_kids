-- Restore NOT NULL: set role from user_roles for any user with NULL role, then enforce NOT NULL.
UPDATE users u
SET role = COALESCE(
  (SELECT ur.role FROM user_roles ur WHERE ur.user_id = u.id ORDER BY CASE ur.role WHEN 'administrator' THEN 1 WHEN 'teacher' THEN 2 WHEN 'course_owner' THEN 3 WHEN 'student' THEN 4 END LIMIT 1),
  'student'
)::user_role
WHERE u.role IS NULL;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'student';
ALTER TABLE users ALTER COLUMN role SET NOT NULL;
