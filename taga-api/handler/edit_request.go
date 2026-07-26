package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"taga-api/model"
	"taga-api/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const filePath = "data/edit_requests.json"

func CreateEditRequest(c *gin.Context) {
	var (
		req      model.EditRequest
		requests []model.EditRequest
		err      error
		data     []byte
	)

	// Bind request
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	req.ID = uuid.New().String()
	req.Status = "pending"
	req.CreatedAt = time.Now().Format(time.RFC3339)

	// Read existing file
	file, _ := os.ReadFile(filePath)

	if len(file) > 0 {
		_ = json.Unmarshal(file, &requests)
	}

	// Append new request
	requests = append(requests, req)

	// Save JSON
	data, err = json.MarshalIndent(requests, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process request",
		})
		return
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save request",
		})
		return
	}

	// Send email
	// SEND EMAIL TO ADMIN
	err = service.SendEditRequestEmail(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Request saved but email failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Edit request submitted successfully",
	})
}
