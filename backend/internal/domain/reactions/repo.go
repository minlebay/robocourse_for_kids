package reactions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Reaction is "like" or "dislike".
const (
	ReactionLike    = "like"
	ReactionDislike = "dislike"
)

var validReaction = map[string]bool{ReactionLike: true, ReactionDislike: true}

// Repository defines data access for lesson and comment reactions.
type Repository interface {
	// Lesson reactions
	GetLessonCounts(ctx context.Context, lessonID uuid.UUID) (likes, dislikes int, err error)
	GetLessonUserReaction(ctx context.Context, lessonID, userID uuid.UUID) (reaction string, ok bool, err error)
	SetLessonReaction(ctx context.Context, lessonID, userID uuid.UUID, reaction string) error
	DeleteLessonReaction(ctx context.Context, lessonID, userID uuid.UUID) (deleted bool, err error)

	// Comment reactions (batch for listing)
	GetCommentCounts(ctx context.Context, commentIDs []uuid.UUID) (map[uuid.UUID]struct{ Likes, Dislikes int }, error)
	GetCommentUserReactions(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]string, error)
	SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error
	DeleteCommentReaction(ctx context.Context, commentID, userID uuid.UUID) (deleted bool, err error)
}

var _ Repository = (*Repo)(nil)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) GetLessonCounts(ctx context.Context, lessonID uuid.UUID) (likes, dislikes int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN reaction = 'like' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN reaction = 'dislike' THEN 1 ELSE 0 END), 0)
		FROM lesson_reactions WHERE lesson_id = $1
	`, lessonID).Scan(&likes, &dislikes)
	return likes, dislikes, err
}

func (r *Repo) GetLessonUserReaction(ctx context.Context, lessonID, userID uuid.UUID) (reaction string, ok bool, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT reaction FROM lesson_reactions WHERE lesson_id = $1 AND user_id = $2
	`, lessonID, userID).Scan(&reaction)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return reaction, true, nil
}

func (r *Repo) SetLessonReaction(ctx context.Context, lessonID, userID uuid.UUID, reaction string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO lesson_reactions (lesson_id, user_id, reaction)
		VALUES ($1, $2, $3)
		ON CONFLICT (lesson_id, user_id) DO UPDATE SET reaction = $3
	`, lessonID, userID, reaction)
	return err
}

func (r *Repo) DeleteLessonReaction(ctx context.Context, lessonID, userID uuid.UUID) (deleted bool, err error) {
	cmd, err := r.pool.Exec(ctx, `
		DELETE FROM lesson_reactions WHERE lesson_id = $1 AND user_id = $2
	`, lessonID, userID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *Repo) GetCommentCounts(ctx context.Context, commentIDs []uuid.UUID) (map[uuid.UUID]struct{ Likes, Dislikes int }, error) {
	if len(commentIDs) == 0 {
		return map[uuid.UUID]struct{ Likes, Dislikes int }{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT comment_id,
			COALESCE(SUM(CASE WHEN reaction = 'like' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN reaction = 'dislike' THEN 1 ELSE 0 END), 0)
		FROM comment_reactions
		WHERE comment_id = ANY($1)
		GROUP BY comment_id
	`, commentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]struct{ Likes, Dislikes int })
	for rows.Next() {
		var id uuid.UUID
		var likes, dislikes int
		if err := rows.Scan(&id, &likes, &dislikes); err != nil {
			return nil, err
		}
		out[id] = struct{ Likes, Dislikes int }{Likes: likes, Dislikes: dislikes}
	}
	// Ensure every comment has an entry (0,0)
	for _, id := range commentIDs {
		if _, ok := out[id]; !ok {
			out[id] = struct{ Likes, Dislikes int }{}
		}
	}
	return out, rows.Err()
}

func (r *Repo) GetCommentUserReactions(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]string, error) {
	if len(commentIDs) == 0 || userID == uuid.Nil {
		return map[uuid.UUID]string{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT comment_id, reaction FROM comment_reactions
		WHERE comment_id = ANY($1) AND user_id = $2
	`, commentIDs, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]string)
	for rows.Next() {
		var id uuid.UUID
		var reaction string
		if err := rows.Scan(&id, &reaction); err != nil {
			return nil, err
		}
		out[id] = reaction
	}
	return out, rows.Err()
}

func (r *Repo) SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO comment_reactions (comment_id, user_id, reaction)
		VALUES ($1, $2, $3)
		ON CONFLICT (comment_id, user_id) DO UPDATE SET reaction = $3
	`, commentID, userID, reaction)
	return err
}

func (r *Repo) DeleteCommentReaction(ctx context.Context, commentID, userID uuid.UUID) (deleted bool, err error) {
	cmd, err := r.pool.Exec(ctx, `
		DELETE FROM comment_reactions WHERE comment_id = $1 AND user_id = $2
	`, commentID, userID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

// GetForLesson returns counts and the user's reaction for a lesson. Implements lessons.LessonReactionProvider.
func (r *Repo) GetForLesson(ctx context.Context, lessonID, userID uuid.UUID) (likes, dislikes int, userReaction string, err error) {
	likes, dislikes, err = r.GetLessonCounts(ctx, lessonID)
	if err != nil {
		return 0, 0, "", err
	}
	if userID != uuid.Nil {
		var ok bool
		userReaction, ok, err = r.GetLessonUserReaction(ctx, lessonID, userID)
		if err != nil {
			return 0, 0, "", err
		}
		if !ok {
			userReaction = ""
		}
	}
	return likes, dislikes, userReaction, nil
}

// GetForComments returns counts and user reactions for a list of comments. Implements comments.CommentReactionProvider.
func (r *Repo) GetForComments(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (
	counts map[uuid.UUID]struct{ Likes, Dislikes int },
	userReactions map[uuid.UUID]string,
	err error,
) {
	counts, err = r.GetCommentCounts(ctx, commentIDs)
	if err != nil {
		return nil, nil, err
	}
	userReactions, err = r.GetCommentUserReactions(ctx, commentIDs, userID)
	if err != nil {
		return nil, nil, err
	}
	return counts, userReactions, nil
}
