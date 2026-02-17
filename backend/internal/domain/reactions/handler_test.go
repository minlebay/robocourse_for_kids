package reactions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

type mockRepo struct {
	SetLessonReactionFn    func(ctx context.Context, lessonID, userID uuid.UUID, reaction string) error
	DeleteLessonReactionFn func(ctx context.Context, lessonID, userID uuid.UUID) (bool, error)
	SetCommentReactionFn   func(ctx context.Context, commentID, userID uuid.UUID, reaction string) error
	DeleteCommentReactionFn func(ctx context.Context, commentID, userID uuid.UUID) (bool, error)
}

func (m *mockRepo) GetLessonCounts(ctx context.Context, lessonID uuid.UUID) (likes, dislikes int, err error) {
	return 0, 0, nil
}
func (m *mockRepo) GetLessonUserReaction(ctx context.Context, lessonID, userID uuid.UUID) (string, bool, error) {
	return "", false, nil
}
func (m *mockRepo) SetLessonReaction(ctx context.Context, lessonID, userID uuid.UUID, reaction string) error {
	if m.SetLessonReactionFn != nil {
		return m.SetLessonReactionFn(ctx, lessonID, userID, reaction)
	}
	return nil
}
func (m *mockRepo) DeleteLessonReaction(ctx context.Context, lessonID, userID uuid.UUID) (bool, error) {
	if m.DeleteLessonReactionFn != nil {
		return m.DeleteLessonReactionFn(ctx, lessonID, userID)
	}
	return true, nil
}
func (m *mockRepo) GetCommentCounts(ctx context.Context, commentIDs []uuid.UUID) (map[uuid.UUID]struct{ Likes, Dislikes int }, error) {
	return nil, nil
}
func (m *mockRepo) GetCommentUserReactions(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]string, error) {
	return nil, nil
}
func (m *mockRepo) SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	if m.SetCommentReactionFn != nil {
		return m.SetCommentReactionFn(ctx, commentID, userID, reaction)
	}
	return nil
}
func (m *mockRepo) DeleteCommentReaction(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	if m.DeleteCommentReactionFn != nil {
		return m.DeleteCommentReactionFn(ctx, commentID, userID)
	}
	return true, nil
}

func jsonRequest(method, target string, body interface{}, params map[string]string, userID uuid.UUID) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, target, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	c.Request = req
	if params != nil {
		for k, v := range params {
			c.Params = append(c.Params, gin.Param{Key: k, Value: v})
		}
	}
	if userID != uuid.Nil {
		c.Set("user_id", userID)
	}
	return w, c
}

func TestSetLessonReaction_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{})
	lessonID := uuid.New()
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/reaction", SetReactionRequest{Reaction: "like"}, map[string]string{"id": lessonID.String()}, uuid.Nil)
	h.SetLessonReaction(c)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want 401", w.Code)
	}
}

func TestSetLessonReaction_InvalidLessonID(t *testing.T) {
	userID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodPut, "/lessons/bad/reaction", SetReactionRequest{Reaction: "like"}, map[string]string{"id": "bad"}, userID)
	h.SetLessonReaction(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", w.Code)
	}
}

func TestSetLessonReaction_InvalidReaction(t *testing.T) {
	lessonID := uuid.New()
	userID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/reaction", SetReactionRequest{Reaction: "invalid"}, map[string]string{"id": lessonID.String()}, userID)
	h.SetLessonReaction(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", w.Code)
	}
}

func TestSetLessonReaction_Success(t *testing.T) {
	lessonID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{
		SetLessonReactionFn: func(ctx context.Context, lid, uid uuid.UUID, reaction string) error {
			if reaction != "like" {
				t.Errorf("reaction = %q; want like", reaction)
			}
			return nil
		},
	}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/reaction", SetReactionRequest{Reaction: "like"}, map[string]string{"id": lessonID.String()}, userID)
	h.SetLessonReaction(c)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("status = %d; want 204 or 200", w.Code)
	}
}

func TestDeleteLessonReaction_Success(t *testing.T) {
	lessonID := uuid.New()
	userID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+lessonID.String()+"/reaction", nil, map[string]string{"id": lessonID.String()}, userID)
	h.DeleteLessonReaction(c)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("status = %d; want 204 or 200", w.Code)
	}
}

func TestSetCommentReaction_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{})
	lessonID := uuid.New()
	commentID := uuid.New()
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/comments/"+commentID.String()+"/reaction", SetReactionRequest{Reaction: "like"}, map[string]string{"id": lessonID.String(), "commentId": commentID.String()}, uuid.Nil)
	h.SetCommentReaction(c)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want 401", w.Code)
	}
}

func TestSetCommentReaction_InvalidCommentID(t *testing.T) {
	lessonID := uuid.New()
	userID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/comments/bad/reaction", SetReactionRequest{Reaction: "dislike"}, map[string]string{"id": lessonID.String(), "commentId": "bad"}, userID)
	h.SetCommentReaction(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", w.Code)
	}
}

func TestSetCommentReaction_Success(t *testing.T) {
	lessonID := uuid.New()
	commentID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{
		SetCommentReactionFn: func(ctx context.Context, cid, uid uuid.UUID, reaction string) error {
			if reaction != "dislike" {
				t.Errorf("reaction = %q; want dislike", reaction)
			}
			return nil
		},
	}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/comments/"+commentID.String()+"/reaction", SetReactionRequest{Reaction: "dislike"}, map[string]string{"id": lessonID.String(), "commentId": commentID.String()}, userID)
	h.SetCommentReaction(c)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("status = %d; want 204 or 200", w.Code)
	}
}

func TestDeleteCommentReaction_Success(t *testing.T) {
	lessonID := uuid.New()
	commentID := uuid.New()
	userID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+lessonID.String()+"/comments/"+commentID.String()+"/reaction", nil, map[string]string{"id": lessonID.String(), "commentId": commentID.String()}, userID)
	h.DeleteCommentReaction(c)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("status = %d; want 204 or 200", w.Code)
	}
}
