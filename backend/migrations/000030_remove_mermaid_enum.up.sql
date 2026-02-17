-- PG 15 has no REMOVE VALUE; recreate enum without 'mermaid'
CREATE TYPE material_kind_new AS ENUM ('link', 'file', 'simulator');

ALTER TABLE lesson_materials
  ALTER COLUMN kind TYPE material_kind_new
  USING kind::text::material_kind_new;

DROP TYPE material_kind;
ALTER TYPE material_kind_new RENAME TO material_kind;
