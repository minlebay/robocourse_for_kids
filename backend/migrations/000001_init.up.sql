-- users
CREATE TYPE user_role AS ENUM ('student', 'teacher');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'student',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_login ON users(login);

-- modules
CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- lessons
CREATE TYPE lesson_type AS ENUM ('theory', 'practice', 'project');

CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    lesson_type lesson_type NOT NULL DEFAULT 'theory',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lessons_module_id ON lessons(module_id);

-- lesson_steps
CREATE TABLE lesson_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_lesson_steps_lesson_id ON lesson_steps(lesson_id);

-- lesson_materials
CREATE TYPE material_kind AS ENUM ('link', 'file');

CREATE TABLE lesson_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    kind material_kind NOT NULL,
    url_or_path TEXT NOT NULL,
    title TEXT
);

-- lesson_tags (many-to-many: lesson <-> tags as text)
CREATE TABLE lesson_tags (
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (lesson_id, tag)
);

-- checklist_items
CREATE TABLE checklist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0
);

-- user_lesson_progress
CREATE TYPE progress_status AS ENUM ('not_started', 'in_progress', 'completed');

CREATE TABLE user_lesson_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    status progress_status NOT NULL DEFAULT 'not_started',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lesson_id)
);

CREATE INDEX idx_user_lesson_progress_user_id ON user_lesson_progress(user_id);
CREATE INDEX idx_user_lesson_progress_lesson_id ON user_lesson_progress(lesson_id);

-- user_checklist_progress
CREATE TABLE user_checklist_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checklist_item_id UUID NOT NULL REFERENCES checklist_items(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, checklist_item_id)
);
