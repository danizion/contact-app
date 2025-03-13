package main

import (
	"log/slog"

	"github.com/danizion/rise/internal/api"
	"github.com/danizion/rise/internal/logger"
	"github.com/danizion/rise/internal/middlewares"
	"github.com/danizion/rise/internal/storage/db"
	"github.com/danizion/rise/internal/storage/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the logger
	logger.Setup()
	slog.Info("Contact application starting up")

	// init db
	postgresDb := db.Init()
	defer postgresDb.Close()
	slog.Info("Database connection initialized")

	// init redis
	redisCache := redis.InitRedis()
	slog.Info("Redis cache connection initialized")

	// create handlers
	handler := api.NewHandler(postgresDb, redisCache)
	slog.Info("API handlers initialized")

	// routing
	router := gin.Default()

	// public endpoints
	router.POST("/users", handler.CreateUser)
	router.POST("/login", handler.Login)

	// protected endpoints (contacts)
	protectedRoutes := router.Group("/")
	protectedRoutes.Use(middlewares.AuthenticateJWT())
	{
		protectedRoutes.GET("/contacts", handler.GetContacts)
		protectedRoutes.POST("/contacts", handler.CreateContact)
		protectedRoutes.PATCH("/contacts/:id", handler.UpdateContact)
		protectedRoutes.DELETE("/contacts/:id", handler.DeleteContact)
	}

	slog.Info("Server starting on :8080")

	// start server
	if err := router.Run(); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}

// TODO: move this part to relevant place
