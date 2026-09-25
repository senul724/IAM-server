package handlers

import (
	"IAM-server/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RemoveSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// RemoveSessionHandler removes a specific user session by its ID, deleting its tokens from Redis first.
// @Summary Remove session
// @Description Revokes tokens in Redis and deletes the session record from DB
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RemoveSessionRequest true "Remove session request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/session/remove [post]
func RemoveSessionHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req RemoveSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionUUID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id format"})
		return
	}

	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id in context"})
		return
	}

	if err := services.DeleteSessionByID(c.Request.Context(), sessionUUID, userUUID); err != nil {
		if err.Error() == "session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session removed successfully",
	})
}

// GetAllSessionsHandler gets all available sessions for the user.
// @Summary Get all sessions
// @Description Gets all available sessions for the user
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/sessions [get]
func GetAllSessionsHandler(c *gin.Context) {
	user, err := resolveUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id in context"})
		return
	}

	sessions, err := services.GetAvailableSessionsByUserID(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Sessions fetched successfully",
		"sessions": sessions,
	})
}
