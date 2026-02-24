package server

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"learn_kids/backend/api"
	"learn_kids/backend/internal/domain/chat"
	"learn_kids/backend/internal/domain/comments"
	"learn_kids/backend/internal/domain/lessons"
	"learn_kids/backend/internal/domain/progress"
	"learn_kids/backend/internal/domain/reactions"
	"learn_kids/backend/internal/domain/users"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	engine   *gin.Engine
	pool     *pgxpool.Pool
	lessons  *lessons.Handler
	users    *users.Handler
	progress *progress.Handler
	chat      *chat.Handler
	comments  *comments.Handler
	reactions *reactions.Handler

	// shutdown context: when cancelled, rate limiter cleanup goroutines exit
	shutdownContext context.Context

	// cached middleware closures (avoid re-creation on every request)
	requireAuthMw    gin.HandlerFunc
	requireTeacherMw gin.HandlerFunc
	requireAdminMw   gin.HandlerFunc
}

type Deps struct {
	Pool             *pgxpool.Pool
	Lessons          *lessons.Handler
	Users            *users.Handler
	Progress         *progress.Handler
	Chat             *chat.Handler
	Comments         *comments.Handler
	Reactions        *reactions.Handler
	JWTSecret        string
	FrontendOrigin   string
	ShutdownContext  context.Context // optional: cancels on server shutdown so rate limiters stop cleanup
}

func New(d Deps) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Only trust loopback and private Docker/LAN subnets as reverse proxies.
	// This ensures c.ClientIP() returns the real client IP (not spoofed X-Forwarded-For).
	_ = engine.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	origin := d.FrontendOrigin
	if origin == "" {
		origin = "http://localhost:5173"
	}
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS(origin))
	engine.Use(middleware.RequestID())
	engine.Use(middleware.HTTPLog())

	// Max request body size: 1 MB (protection against OOM)
	engine.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})

	s := &Server{
		engine:           engine,
		pool:             d.Pool,
		lessons:          d.Lessons,
		users:            d.Users,
		progress:         d.Progress,
		chat:             d.Chat,
		comments:         d.Comments,
		reactions:        d.Reactions,
		shutdownContext:  d.ShutdownContext,
		requireAuthMw:    middleware.RequireAuth(),
		requireTeacherMw: middleware.RequireTeacher(),
		requireAdminMw:   middleware.RequireAdmin(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.engine.GET("/api/v1/health", s.health)
	s.engine.GET("/api/docs", func(c *gin.Context) {
		b, _ := api.FS.ReadFile("openapi.yaml")
		c.Data(http.StatusOK, "application/x-yaml", b)
	})

	// Rate limiters (shutdown context stops cleanup goroutines on server exit)
	apiLimiter  := middleware.NewRateLimiter(300, time.Minute, s.shutdownContext) // general: all API endpoints
	authLimiter := middleware.NewRateLimiter(10, time.Minute, s.shutdownContext)  // strict: auth endpoints
	chatLimiter := middleware.NewRateLimiter(20, time.Minute, s.shutdownContext)  // strict: AI chat

	apiGroup := s.engine.Group("/api/v1")
	apiGroup.Use(apiLimiter.Handler())
	apiGroup.Use(middleware.Auth(s.users))
	apiGroup.Use(middleware.RequireFreshPassword())

	// Public auth (rate limited)
	apiGroup.POST("/auth/register", authLimiter.Handler(), s.users.RegisterUser)
	apiGroup.POST("/auth/login", authLimiter.Handler(), s.users.Login)

	// Change password (auth required, exempt from RequireFreshPassword block)
	apiGroup.POST("/auth/change-password", s.requireAuth, s.users.ChangePassword)

	// Lessons (public read; teacher can create/update)
	apiGroup.GET("/modules", s.lessons.ListModules)
	apiGroup.GET("/modules/:id", s.lessons.GetModule)
	apiGroup.GET("/lessons/:id", s.lessons.GetLesson)
	apiGroup.PUT("/lessons/:id", s.requireTeacher, s.lessons.UpdateLesson)
	apiGroup.DELETE("/lessons/:id", s.requireTeacher, s.lessons.DeleteLesson)
	apiGroup.POST("/modules", s.requireTeacher, s.lessons.CreateModule)
	apiGroup.DELETE("/modules/:id", s.requireTeacher, s.lessons.DeleteModule)
	apiGroup.POST("/modules/:id/lessons", s.requireTeacher, s.lessons.CreateLesson)

	// Comments (public read; auth required to post)
	apiGroup.GET("/lessons/:id/comments", s.comments.List)
	apiGroup.POST("/lessons/:id/comments", s.requireAuthMw, s.comments.Create)
	apiGroup.DELETE("/lessons/:id/comments/:commentId", s.requireAuthMw, s.comments.Delete)

	// Reactions (likes/dislikes) for lessons and comments — JWT required to set/delete
	apiGroup.PUT("/lessons/:id/reaction", s.requireAuthMw, s.reactions.SetLessonReaction)
	apiGroup.DELETE("/lessons/:id/reaction", s.requireAuthMw, s.reactions.DeleteLessonReaction)
	apiGroup.PUT("/lessons/:id/comments/:commentId/reaction", s.requireAuthMw, s.reactions.SetCommentReaction)
	apiGroup.DELETE("/lessons/:id/comments/:commentId/reaction", s.requireAuthMw, s.reactions.DeleteCommentReaction)

	// Auth required from here for some routes
	apiGroup.GET("/auth/me", s.requireAuthMw, s.users.Me)
	apiGroup.PATCH("/auth/me", s.requireAuthMw, s.users.UpdateMe)
	apiGroup.GET("/progress", s.requireAuthMw, s.progress.GetProgress)
	apiGroup.PUT("/lessons/:id/progress", s.requireAuthMw, s.progress.SetLessonProgress)
	apiGroup.PUT("/lessons/:id/checklist/:itemId", s.requireAuthMw, s.progress.SetChecklistItem)

	// Chat with Gemini (auth required, rate limited)
	apiGroup.POST("/chat", s.requireAuthMw, chatLimiter.Handler(), s.chat.Chat)
	apiGroup.GET("/chat/:lessonId/history", s.requireAuthMw, s.chat.GetHistory)
	apiGroup.DELETE("/chat/:lessonId/history", s.requireAuthMw, s.chat.ClearHistory)

	// Teacher only
	apiGroup.GET("/users", s.requireTeacherMw, s.users.ListUsers)
	apiGroup.DELETE("/users/:id", s.requireTeacherMw, s.users.DeleteUser)
	apiGroup.GET("/users/:id/progress", s.requireTeacherMw, s.progress.GetUserProgress)

	// Admin only
	// Note: requireAuth is already applied by apiGroup.Use(middleware.Auth(...)) above;
	// we only need requireAdmin here.
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(s.requireAdmin)
	adminGroup.GET("/users", s.users.AdminListUsers)
	adminGroup.POST("/users", s.users.AdminCreateUser)
	adminGroup.DELETE("/users/:id", s.users.AdminDeleteUser)
	adminGroup.POST("/users/:id/block", s.users.AdminBlockUser)
	adminGroup.POST("/users/:id/reset-password", s.users.AdminResetPassword)
	adminGroup.GET("/stats", s.users.AdminGetStats)
	adminGroup.GET("/activity", s.users.AdminGetActivity)

	// Frontend static (optional: serve from ./web when present, e.g. in Docker)
	if _, err := os.Stat("./web/index.html"); err == nil {
		s.engine.Static("/assets", "./web/assets")
		s.engine.StaticFile("/vite.svg", "./web/vite.svg")
		s.engine.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File("./web/index.html")
		})
	}
}

func (s *Server) requireAuth(c *gin.Context) {
	s.requireAuthMw(c)
}

func (s *Server) requireTeacher(c *gin.Context) {
	s.requireTeacherMw(c)
}

func (s *Server) requireAdmin(c *gin.Context) {
	s.requireAdminMw(c)
}

func (s *Server) health(c *gin.Context) {
	if err := s.pool.Ping(c.Request.Context()); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) Handler() http.Handler {
	return s.engine
}
