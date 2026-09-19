package handlers

import (
	"IAM-server/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CreateCustomerNoteRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type UpdateCustomerNoteRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

// resolveNoteID extracts the note ID from the route params.
func resolveNoteID(c *gin.Context) (string, bool) {
	noteID := c.Param("id")

	if noteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "note ID parameter is required"})
		return "", false
	}

	return noteID, true
}

// GetCustomerNotesHandler retrieves all notes for a customer.
// @Summary Get customer notes
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param customer_id path string false "Customer ID (optional if authenticated)"
// @Success 200 {array} models.Note
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes [get]
func GetCustomerNotesHandler(c *gin.Context) {
	customer, _ := resolveUser(c)

	notes, err := services.GetCustomerNotes(c.Request.Context(), customer.ID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid customer ID") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notes": notes,
	})
}

// GetCustomerNoteHandler retrieves a single note by ID for a customer.
// @Summary Get customer note by ID
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path string true "Note ID"
// @Param customer_id path string false "Customer ID (optional if authenticated)"
// @Success 200 {object} models.Note
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/{id} [get]
func GetCustomerNoteHandler(c *gin.Context) {
	customer, _ := resolveUser(c)

	noteID, ok := resolveNoteID(c)
	if !ok {
		return
	}

	note, err := services.GetCustomerNote(c.Request.Context(), customer.ID, noteID)
	if err != nil {
		if err.Error() == "note not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"note": note,
	})
}

// CreateCustomerNoteHandler creates a new note for a customer.
// @Summary Create customer note
// @Tags Notes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateCustomerNoteRequest true "Note content"
// @Param customer_id path string false "Customer ID (optional if authenticated)"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes [post]
func CreateCustomerNoteHandler(c *gin.Context) {
	customer, _ := resolveUser(c)

	var req CreateCustomerNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := services.CreateCustomerNote(c.Request.Context(), customer.ID, req.Title, req.Content)
	if err != nil {
		if strings.Contains(err.Error(), "invalid customer ID") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Note created successfully",
		"note":    note,
	})
}

// UpdateCustomerNoteHandler updates an existing note for a customer.
// @Summary Update customer note
// @Tags Notes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Note ID"
// @Param request body UpdateCustomerNoteRequest true "Fields to update"
// @Param customer_id path string false "Customer ID (optional if authenticated)"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/{id} [put]
func UpdateCustomerNoteHandler(c *gin.Context) {
	customer, _ := resolveUser(c)

	noteID, ok := resolveNoteID(c)
	if !ok {
		return
	}

	var req UpdateCustomerNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := services.UpdateCustomerNote(c.Request.Context(), customer.ID, noteID, req.Title, req.Content)
	if err != nil {
		if err.Error() == "note not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note updated successfully",
		"note":    note,
	})
}

// DeleteCustomerNoteHandler removes a note for a customer.
// @Summary Delete customer note
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path string true "Note ID"
// @Param customer_id path string false "Customer ID (optional if authenticated)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/{id} [delete]
func DeleteCustomerNoteHandler(c *gin.Context) {
	customer, _ := resolveUser(c)

	noteID, ok := resolveNoteID(c)
	if !ok {
		return
	}

	err := services.DeleteCustomerNote(c.Request.Context(), customer.ID, noteID)
	if err != nil {
		if err.Error() == "note not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note deleted successfully",
	})
}
