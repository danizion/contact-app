package main

import (
	"github.com/danizion/contact-app/internal/utils"
	"log/slog"

	"github.com/danizion/contact-app/internal/api"
	"github.com/danizion/contact-app/internal/logger"
	"github.com/danizion/contact-app/internal/middlewares"
	"github.com/danizion/contact-app/internal/storage/db"
	"github.com/danizion/contact-app/internal/storage/redis"
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

	port := utils.GetEnvOrDefault("PORT", "8080")
	router.Run(port)
	slog.Info("Server started on port", "port", port)
	// start server
	if err := router.Run(); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
