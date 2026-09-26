package main

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/routes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title IAM Server
// @version 1.0
// @description Backend API for IAM server.
// @host localhost:3030
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	connections.ConnectDB()
	connections.ConnectRedis()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		connections.Migrate()
		return
	}

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Device-Type", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	api := r.Group("/api")

	api.GET("/hi", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"Message": "Hi there",
		})
	})

	// auth routes
	routes.SetAuthRoutes(api)

	// note routes
	routes.SetNoteRoutes(api)

	r.Run(":3030")
}
