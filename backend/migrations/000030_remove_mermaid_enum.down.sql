-- Re-add mermaid to enum (recreate type)
CREATE TYPE material_kind_old AS ENUM ('link', 'file', 'simulator', 'mermaid');

ALTER TABLE lesson_materials
  ALTER COLUMN kind TYPE material_kind_old USING kind::text::material_kind_old;

DROP TYPE material_kind;
ALTER TYPE material_kind_old RENAME TO material_kind;
