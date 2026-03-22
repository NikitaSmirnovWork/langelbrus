package main

import (
	"log"
	"net/http"
	"os"
	"wordbot/database"
	"wordbot/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Init DB
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./wordbot.db"
	}
	database.Init(dbPath)

	// Gin setup
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS — allow Telegram WebApp origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "X-Telegram-Init-Data", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Serve the Mini App HTML
	r.StaticFile("/", "./index.html")
	r.StaticFile("/index.html", "./index.html")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes (all protected by Telegram auth)
	api := r.Group("/api", handlers.TelegramAuth())
	{
		// Modules
		api.GET("/modules",      handlers.GetModules)
		api.POST("/modules",     handlers.CreateModule)
		api.PUT("/modules/:id",  handlers.UpdateModule)
		api.DELETE("/modules/:id", handlers.DeleteModule)

		// Topics (nested under module)
		api.GET("/modules/:moduleId/topics",       handlers.GetTopics)
		api.POST("/modules/:moduleId/topics",      handlers.CreateTopic)
		api.PUT("/modules/:moduleId/topics/:id",   handlers.UpdateTopic)
		api.DELETE("/modules/:moduleId/topics/:id",handlers.DeleteTopic)

		// Words (nested under topic)
		api.GET("/topics/:topicId/words",       handlers.GetWords)
		api.POST("/topics/:topicId/words",      handlers.CreateWord)
		api.PUT("/topics/:topicId/words/:id",   handlers.UpdateWord)
		api.DELETE("/topics/:topicId/words/:id",handlers.DeleteWord)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s", port)
	r.Run(":" + port)
}
