package http

import (
	"crypto/rsa"
	"net/http"

	"github.com/gin-gonic/gin"

	"careergps/internal/interfaces/http/handlers"
	"careergps/internal/interfaces/http/middleware"
)

// SetupRouter wires all routes and returns the configured Gin engine.
func SetupRouter(
	authHandler *handlers.AuthHandler,
	assessmentHandler *handlers.AssessmentHandler,
	coachHandler *handlers.CoachHandler,
	publicKey *rsa.PublicKey,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// Auth — unauthenticated
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// Authenticated routes
	authed := v1.Group("")
	authed.Use(middleware.JWT(publicKey))
	{
		// Assessments
		authed.GET("/assessments/:id", assessmentHandler.Get)
		authed.GET("/assessments/:id/readiness", assessmentHandler.GetReadiness)

		// Coach
		authed.POST("/coach/sessions", coachHandler.CreateSession)
		authed.GET("/coach/sessions/:id", coachHandler.GetSession)
		authed.POST("/coach/sessions/:id/messages", coachHandler.SendMessage)
		authed.GET("/coach/sessions/:id/stream", coachHandler.Stream)
	}

	return r
}
