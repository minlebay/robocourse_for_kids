-- Multi-role support: user_roles (user_id, role) with roles: administrator, teacher, course_owner, student
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('administrator', 'teacher', 'course_owner', 'student')),
    PRIMARY KEY (user_id, role)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);

-- Backfill from users.role (one role per user)
INSERT INTO user_roles (user_id, role)
SELECT id, role::TEXT FROM users
ON CONFLICT (user_id, role) DO NOTHING;

-- Module ownership: one owner per module (nullable for existing modules)
ALTER TABLE modules ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_modules_owner_id ON modules(owner_id);
