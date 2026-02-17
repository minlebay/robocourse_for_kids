-- Delete mermaid materials (REMOVE VALUE in next migration to avoid same-transaction use)
DELETE FROM lesson_materials WHERE kind = 'mermaid';
