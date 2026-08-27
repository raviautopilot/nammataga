package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/audit"
	"time"
)

var grievances []model.Grievance

// CreateGrievance godoc
// @Summary Submit a new grievance
// @Description Submit a grievance with subject, category, priority, etc.
// @Tags Grievances
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param grievance body model.Grievance true "Grievance Data"
// @Success 200 {object} model.Grievance
// @Failure 400 {object} map[string]string
// @Router /api/grievances [post]
func CreateGrievance(c *gin.Context) {
	config.Logger.Info("CreateGrievance called")
	var newGrievance model.Grievance

	// 1. Bind request JSON
	if err := c.ShouldBindJSON(&newGrievance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.Logger.Debug("Received grievance payload", zap.Any("grievance", newGrievance))
	// 2. Read existing file
	filePath := "data/grievance/grievanceg.json"

	var grievances []model.Grievance

	file, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(file, &grievances)
	}

	// 3. Add ID + Dates + User Info
	newGrievance.ID = fmt.Sprintf("GRV-%d", time.Now().Unix())
	newGrievance.MemberName = c.GetString("member_name")
	newGrievance.MemberEmail = c.GetString("member_email")
	newGrievance.Status = "Pending"
	newGrievance.SubmittedDate = time.Now()
	newGrievance.LastUpdate = time.Now()

	// 4. Append new grievance
	grievances = append(grievances, newGrievance)

	// 5. Write back to file
	updatedData, err := json.MarshalIndent(grievances, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to prepare grievance",
		})
		return
	}

	err = os.WriteFile(filePath, updatedData, 0644)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save grievance",
		})
		return
	}

	// Send email after save
	err = service.SendGrievanceEmail(
		newGrievance.Subject,
		newGrievance.Category,
		newGrievance.Priority,
		newGrievance.Description,
		newGrievance.ContactPhone,
		newGrievance.MemberName,
		newGrievance.MemberEmail,
		newGrievance.PreferredResponse,
	)

	if err != nil {
		config.Logger.Error(
			"Failed to send grievance email",
			zap.Error(err),
		)
	} else {
		config.Logger.Info("Grievance email sent successfully")
	}

	actorID := ""
	actorName := ""
	if val, ok := c.Get("username"); ok {
		actorID = "admin"
		actorName, _ = val.(string)
	} else {
		memberID := c.GetString("member_id")
		if memberID != "" {
			actorID = getMemberTagaIdByUUID(memberID)
		} else {
			actorID = "anonymous"
		}
		actorName = c.GetString("member_email")
		if actorName == "" {
			actorName = "Anonymous"
		}
	}

	_ = audit.Log(c, actorID, actorName,
		audit.ActionCreate, audit.ModuleGrievance,
		"grievance", newGrievance.ID,
		fmt.Sprintf("Submitted grievance: %s", newGrievance.Subject),
		nil, newGrievance)

	// Return response
	c.JSON(http.StatusOK, newGrievance)
}

// GetGrievances godoc
// @Summary Get all grievances
// @Description Fetch all submitted grievances
// @Tags Grievances
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Grievance
// @Router /api/grievances [get]
func GetGrievances(c *gin.Context) {
	filePath := "data/grievance/grievanceg.json"
	oldFilePath := "data/grievance/old_grievanceg.json"
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.WriteFile(filePath, []byte("[]"), 0644)
	}
	var grievances []model.Grievance
	file, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusOK, []model.Grievance{})
		return
	}
	json.Unmarshal(file, &grievances)

	// Auto-Archive logic (older than 6 months = ~4380 hours)
	var activeGrievances []model.Grievance
	var archivedGrievances []model.Grievance
	
	now := time.Now()
	for _, g := range grievances {
		if g.Status == "Read" && now.Sub(g.SubmittedDate).Hours() > 4380 {
			archivedGrievances = append(archivedGrievances, g)
		} else {
			activeGrievances = append(activeGrievances, g)
		}
	}

	// If we archived anything, save to both files
	if len(archivedGrievances) > 0 {
		var existingOld []model.Grievance
		oldFile, _ := os.ReadFile(oldFilePath)
		if len(oldFile) > 0 {
			json.Unmarshal(oldFile, &existingOld)
		}
		existingOld = append(existingOld, archivedGrievances...)
		
		oldData, _ := json.MarshalIndent(existingOld, "", "  ")
		os.WriteFile(oldFilePath, oldData, 0644)
		
		activeData, _ := json.MarshalIndent(activeGrievances, "", "  ")
		os.WriteFile(filePath, activeData, 0644)
		
		grievances = activeGrievances
	}

	c.JSON(http.StatusOK, grievances)
}

// GetGrievanceByID godoc
// @Summary Get grievance by ID
// @Tags Grievances
// @Produce json
// @Security BearerAuth
// @Param id path string true "Grievance ID"
// @Success 200 {object} model.Grievance
// @Failure 404 {object} map[string]string
// @Router /api/grievances/{id} [get]
func GetGrievanceByID(c *gin.Context) {
	id := c.Param("id")

	for _, g := range grievances {
		if g.ID == id {
			c.JSON(http.StatusOK, g)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Grievance not found",
	})
}

// UpdateGrievance godoc
// @Summary Update grievance
// @Tags Grievances
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Grievance ID"
// @Param grievance body model.Grievance true "Updated Data"
// @Success 200 {object} model.Grievance
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/grievances/{id} [put]
func UpdateGrievance(c *gin.Context) {
	id := c.Param("id")

	var updated model.Grievance
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	filePath := "data/grievance/grievanceg.json"
	var grievances []model.Grievance
	file, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(file, &grievances)
	}

	for i, g := range grievances {
		if g.ID == id {
			// Update fields
			grievances[i].MemberName = updated.MemberName
			grievances[i].MemberEmail = updated.MemberEmail
			grievances[i].Subject = updated.Subject
			grievances[i].Category = updated.Category
			grievances[i].Priority = updated.Priority
			grievances[i].Description = updated.Description
			grievances[i].ContactPhone = updated.ContactPhone
			grievances[i].PreferredResponse = updated.PreferredResponse
			grievances[i].Status = updated.Status
			grievances[i].AssignedTo = updated.AssignedTo
			grievances[i].LastUpdate = time.Now()

			updatedData, _ := json.MarshalIndent(grievances, "", "  ")
			os.WriteFile("data/grievance/grievanceg.json", updatedData, 0644)

			actorID := ""
			actorName := ""
			if val, ok := c.Get("username"); ok {
				actorID = "admin"
				actorName, _ = val.(string)
			} else {
				memberID := c.GetString("member_id")
				if memberID != "" {
					actorID = getMemberTagaIdByUUID(memberID)
				} else {
					actorID = "anonymous"
				}
				actorName = c.GetString("member_email")
				if actorName == "" {
					actorName = "Anonymous"
				}
			}

			_ = audit.Log(c, actorID, actorName,
				audit.ActionUpdate, audit.ModuleGrievance,
				"grievance", grievances[i].ID,
				fmt.Sprintf("Updated grievance status to: %s", grievances[i].Status),
				g, grievances[i])

			c.JSON(http.StatusOK, grievances[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Grievance not found",
	})
}

// DeleteGrievance godoc
// @Summary Delete grievance
// @Tags Grievances
// @Security BearerAuth
// @Param id path string true "Grievance ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/grievances/{id} [delete]
func DeleteGrievance(c *gin.Context) {
	id := c.Param("id")

	for i, g := range grievances {
		if g.ID == id {
			// Remove from slice
			grievances = append(grievances[:i], grievances[i+1:]...)

			// Write back to file
			updatedData, _ := json.MarshalIndent(grievances, "", "  ")
			_ = os.WriteFile("data/grievance/grievanceg.json", updatedData, 0644)

			actorID := ""
			actorName := ""
			if val, ok := c.Get("username"); ok {
				actorID = "admin"
				actorName, _ = val.(string)
			} else {
				memberID := c.GetString("member_id")
				if memberID != "" {
					actorID = getMemberTagaIdByUUID(memberID)
				} else {
					actorID = "anonymous"
				}
				actorName = c.GetString("member_email")
				if actorName == "" {
					actorName = "Anonymous"
				}
			}

			_ = audit.Log(c, actorID, actorName,
				audit.ActionDelete, audit.ModuleGrievance,
				"grievance", g.ID,
				fmt.Sprintf("Deleted grievance: %s", g.Subject),
				g, nil)

			c.JSON(http.StatusOK, gin.H{
				"message": "Grievance deleted successfully",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Grievance not found",
	})
}



// GetCategories godoc
// @Summary Get grievance categories
// @Description Fetch all grievance categories from JSON file
// @Tags Grievances
// @Produce json
// @Security BearerAuth
// @Success 200 {array} string
// @Router /api/categories [get]
func GetCategories(c *gin.Context) {
	filePath := "data/grievance/categories.json"

	file, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusOK, []string{})
		return
	}

	var categories []string
	json.Unmarshal(file, &categories)

	c.JSON(http.StatusOK, categories)
}

// GetPriorities godoc
// @Summary Get grievance priorities
// @Description Fetch all grievance priorities from JSON file
// @Tags Grievances
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]string
// @Router /api/priorities [get]
func GetPriorities(c *gin.Context) {
	filePath := "data/grievance/priorities.json"

	file, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusOK, []map[string]string{})
		return
	}

	var priorities []map[string]string
	json.Unmarshal(file, &priorities)

	c.JSON(http.StatusOK, priorities)
}

// GetGrievanceBanner godoc
// @Summary Get grievance banner image info
// @Description Get the relative path/url of the grievance banner image
// @Tags Grievances
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/grievance-banner [get]
func GetGrievanceBanner(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"image": "/api/images/grievance-banner.jpg",
	})
}
