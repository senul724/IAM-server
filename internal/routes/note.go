package routes

import (
	"IAM-server/internal/handlers"
	"IAM-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetNoteRoutes(api *gin.RouterGroup) {
	notes := api.Group("/notes")
	notes.Use(middleware.ProtectRoute())

	notes.GET("", handlers.GetCustomerNotesHandler)
	notes.GET("/:id", handlers.GetCustomerNoteHandler)
	notes.POST("", handlers.CreateCustomerNoteHandler)
	notes.PUT("/:id", handlers.UpdateCustomerNoteHandler)
	notes.DELETE("/:id", handlers.DeleteCustomerNoteHandler)
}
