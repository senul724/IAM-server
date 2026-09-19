package routes

import (
	"IAM-server/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetAuthRoutes(api *gin.RouterGroup) {
	// auth route group
	auth := api.Group("/auth")

	// register
	auth.POST("/register", handlers.RegisterWithOTPHandler)

	// login
	auth.POST("/login", handlers.LoginWithPasswordHandler)
	auth.POST("/login-otp", handlers.LoginWithOTPHandler)

	// refresh & session
	auth.POST("/refresh", handlers.RefreshTokenHandler)

	//logout
	auth.POST("/logout", handlers.LogoutHandler)

	// forgot password
	auth.POST("/forgot-password", handlers.SendResetPasswordOTPHandler)

	// reset password
	auth.POST("/reset-password", handlers.ResetPasswordHandler)

	// email & otp checks
	auth.POST("/check-email", handlers.CheckEmailHandler)
	auth.POST("/send-login-otp", handlers.SendLoginOTPHandler)
}
