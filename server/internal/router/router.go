package router

import (
	"database/sql"
	"net/http"

	"teacher-platform/server/internal/config"
	"teacher-platform/server/internal/middleware"
	"teacher-platform/server/internal/response"
	"teacher-platform/server/internal/service/auth"
	"teacher-platform/server/internal/service/profile"
	"teacher-platform/server/internal/service/system"
	"teacher-platform/server/internal/service/training"
	"teacher-platform/server/internal/service/treehole"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, db *sql.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Static("/uploads", "uploads")

	r.GET("/healthz", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			response.Fail(c, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		response.OK(c, gin.H{"status": "up"})
	})

	api := r.Group("/api/v1")
	auth.RegisterRoutes(api.Group("/auth"), cfg, db)

	protected := api.Group("")
	protected.Use(middleware.RequireToken(cfg))
	profile.RegisterRoutes(protected.Group("/profile"), db)
	treehole.RegisterRoutes(protected.Group("/treeholes"), db)
	training.RegisterRoutes(protected.Group("/trainings"), db)
	system.RegisterRoutes(protected.Group("/system"), db)

	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, http.StatusNotFound, "route not found")
	})

	return r
}
