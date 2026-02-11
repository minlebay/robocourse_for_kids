package server

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"learn_kids/backend/internal/domain/chat"
	"learn_kids/backend/internal/domain/comments"
	"learn_kids/backend/internal/domain/lessons"
	"learn_kids/backend/internal/domain/progress"
	"learn_kids/backend/internal/domain/users"
	"learn_kids/backend/internal/middleware"
	"learn_kids/backend/schematics"
)

type Server struct {
	engine   *gin.Engine
	pool     *pgxpool.Pool
	lessons  *lessons.Handler
	users    *users.Handler
	progress *progress.Handler
	chat     *chat.Handler
	comments *comments.Handler
}

type Deps struct {
	Pool      *pgxpool.Pool
	Lessons   *lessons.Handler
	Users     *users.Handler
	Progress  *progress.Handler
	Chat      *chat.Handler
	Comments  *comments.Handler
	JWTSecret string
}

func New(d Deps) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS())
	engine.Use(middleware.RequestID())
	engine.Use(gin.Logger())

	s := &Server{
		engine:   engine,
		pool:     d.Pool,
		lessons:  d.Lessons,
		users:    d.Users,
		progress: d.Progress,
		chat:     d.Chat,
		comments: d.Comments,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.engine.GET("/api/v1/health", s.health)
	s.engine.StaticFile("/api/docs", "./api/openapi.yaml")

	api := s.engine.Group("/api/v1")
	api.Use(middleware.Auth(s.users))

	// Public auth
	api.POST("/auth/register", s.users.RegisterUser)
	api.POST("/auth/login", s.users.Login)

	// Lessons (public read; teacher can update)
	s.lessons.Register(api)
	api.PUT("/lessons/:id", s.requireAuth, s.requireTeacher, s.lessons.UpdateLesson)

	// Comments (public read; auth required to post)
	api.GET("/lessons/:id/comments", s.comments.List)
	api.POST("/lessons/:id/comments", s.requireAuth, s.comments.Create)
	api.DELETE("/lessons/:id/comments/:commentId", s.requireAuth, s.comments.Delete)

	// Auth required from here for some routes
	api.GET("/auth/me", s.requireAuth, s.users.Me)
	api.GET("/progress", s.requireAuth, s.progress.GetProgress)
	api.PUT("/lessons/:id/progress", s.requireAuth, s.progress.SetLessonProgress)
	api.PUT("/lessons/:id/checklist/:itemId", s.requireAuth, s.progress.SetChecklistItem)

	// Chat with Gemini (optional auth: works without login for kids)
	api.POST("/chat", s.chat.Chat)

	// Teacher only
	api.GET("/users", s.requireTeacher, s.users.ListUsers)
	api.GET("/users/:id/progress", s.requireTeacher, s.progress.GetUserProgress)

	// Схемы всегда из встроенных в бинарник файлов (работает и в Docker, и локально)
	s.engine.StaticFS("/schematics", http.FS(schematics.FS))

	// Frontend static (optional: serve from ./web when present, e.g. in Docker)
	if _, err := os.Stat("./web/index.html"); err == nil {
		s.engine.Static("/assets", "./web/assets")
		s.engine.StaticFile("/vite.svg", "./web/vite.svg")
		s.engine.NoRoute(func(c *gin.Context) {
			c.File("./web/index.html")
		})
	}
}

func (s *Server) requireAuth(c *gin.Context) {
	middleware.RequireAuth()(c)
}

func (s *Server) requireTeacher(c *gin.Context) {
	s.users.RequireTeacher(c)
}

func (s *Server) health(c *gin.Context) {
	if err := s.pool.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) Handler() http.Handler {
	return s.engine
}
