package router

import (
	"time"

	"blogapp/backend/internal/config"
	"blogapp/backend/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, h *handler.Handler) *gin.Engine {
	router := gin.Default()
	router.Static("/uploads", cfg.UploadDir)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendOrigin, "http://127.0.0.1:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)
		api.GET("/articles", h.ListArticles)
		api.GET("/articles/popular", h.PopularArticles)
		api.GET("/articles/:slug", h.GetArticleBySlug)
		api.GET("/articles/id/:id/comments", h.ListComments)
		api.GET("/categories", h.ListCategories)

		auth := api.Group("")
		auth.Use(h.AuthRequired())
		{
			auth.GET("/auth/me", h.Me)
			auth.PATCH("/auth/me", h.UpdateMe)
			auth.POST("/auth/me/avatar", h.UploadAvatar)
			auth.POST("/articles/id/:id/comments", h.CreateComment)
			auth.POST("/articles/id/:id/reactions/:type", h.ToggleReaction)
			auth.GET("/me/library", h.Library)
		}

		admin := api.Group("/admin")
		admin.Use(h.AuthRequired(), h.AdminRequired())
		{
			admin.GET("/stats", h.AdminStats)
			admin.POST("/uploads", h.UploadFile)
			admin.GET("/articles", h.AdminListArticles)
			admin.POST("/articles", h.UpsertArticle)
			admin.PUT("/articles/:id", h.UpsertArticle)
			admin.DELETE("/articles/:id", h.DeleteArticle)
			admin.GET("/comments", h.AdminListComments)
			admin.PATCH("/comments/:id", h.ModerateComment)
			admin.GET("/users", h.ListUsers)
			admin.GET("/categories", h.AdminListCategories)
			admin.POST("/categories", h.CreateCategory)
			admin.PUT("/categories/:id", h.UpdateCategory)
			admin.DELETE("/categories/:id", h.DeleteCategory)
		}
	}

	return router
}
