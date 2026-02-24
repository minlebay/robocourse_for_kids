-- Limit chat message text to 100 KB.
-- Gemini responses are already truncated at 2 MB at the application layer,
-- but an explicit DB-level constraint prevents unbounded rows.
ALTER TABLE chat_messages
  ALTER COLUMN text TYPE VARCHAR(102400);
