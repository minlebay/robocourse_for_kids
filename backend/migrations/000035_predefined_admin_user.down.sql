-- Revert administrator to teacher role (as originally set by migration 000002).
UPDATE users
SET role = 'teacher', must_change_password = false
WHERE id = '00000000-0000-0000-0000-000000000001';
