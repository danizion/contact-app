package main

import (
	"github.com/danizion/rise/internal/middlewares"
	"log"

	"github.com/danizion/rise/internal/storage/db"
	"github.com/danizion/rise/internal/storage/redis"

	"github.com/danizion/rise/internal/api"
	"github.com/gin-gonic/gin"
)

func main() {
	// init db
	postgresDb := db.Init()
	defer postgresDb.Close()

	// init redis
	redisCache := redis.InitRedis()

	// create handlers
	handler := api.NewHandler(postgresDb, redisCache)

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
		protectedRoutes.PATCH("/contacts", handler.UpdateContact)
		protectedRoutes.DELETE("/contacts", handler.DeleteContact)
	}

	// start server
	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}

// TODO: move this part to relevant place
