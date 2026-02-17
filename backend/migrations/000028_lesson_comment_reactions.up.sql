-- Реакции (лайк/дизлайк) к урокам: один тип реакции на пользователя на урок.
CREATE TABLE lesson_reactions (
  lesson_id  uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
  user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reaction   text NOT NULL CHECK (reaction IN ('like', 'dislike')),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (lesson_id, user_id)
);

CREATE INDEX idx_lesson_reactions_lesson_id ON lesson_reactions(lesson_id);

-- Реакции к комментариям.
CREATE TABLE comment_reactions (
  comment_id uuid NOT NULL REFERENCES lesson_comments(id) ON DELETE CASCADE,
  user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reaction   text NOT NULL CHECK (reaction IN ('like', 'dislike')),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX idx_comment_reactions_comment_id ON comment_reactions(comment_id);
