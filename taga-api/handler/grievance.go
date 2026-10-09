package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/audit"
)

const grievanceFilePath = "data/grievance/grievanceg.json"
const oldGrievanceFilePath = "data/grievance/old_grievanceg.json"

var grievanceFileLock sync.RWMutex

func readGrievancesFromDisk() ([]model.Grievance, error) {
	if _, err := os.Stat(grievanceFilePath); os.IsNotExist(err) {
		_ = os.MkdirAll("data/grievance", 0755)
		_ = os.WriteFile(grievanceFilePath, []byte("[]"), 0644)
		return []model.Grievance{}, nil
	}

	data, err := os.ReadFile(grievanceFilePath)
	if err != nil {
		return []model.Grievance{}, err
	}

	var list []model.Grievance
	if err := json.Unmarshal(data, &list); err != nil {
		return []model.Grievance{}, err
	}
	return list, nil
}

func saveGrievancesToDisk(list []model.Grievance) error {
	_ = os.MkdirAll("data/grievance", 0755)
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(grievanceFilePath, data, 0644)
}

func isGrievanceAdmin(c *gin.Context) bool {
	if role, exists := c.Get("role"); exists && role != nil {
		if r, ok := role.(string); ok && (r == "admin" || r == "superadmin") {
			return true
		}
	}
	if username, exists := c.Get("username"); exists && username != nil {
		if u, ok := username.(string); ok && u != "" {
			return true
		}
	}
	return false
}

func getCallerEmail(c *gin.Context) string {
	if e := c.GetString("member_email"); e != "" {
		return e
	}
	if val, exists := c.Get("email"); exists && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

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

	// Resolve member identity from context and fallback to request
	memberName := c.GetString("member_name")
	memberEmail := getCallerEmail(c)
	if memberEmail == "" && newGrievance.MemberEmail != "" {
		memberEmail = newGrievance.MemberEmail
	}
	if memberName == "" && memberEmail != "" {
		memberName = getMemberNameByEmail(memberEmail)
	}
	if memberName == "" && newGrievance.MemberName != "" {
		memberName = newGrievance.MemberName
	}
	if memberName == "" {
		memberName = "Member"
	}

	// 2. Set ID + Dates + User Info
	newGrievance.ID = fmt.Sprintf("GRV-%d", time.Now().Unix())
	newGrievance.MemberName = memberName
	newGrievance.MemberEmail = memberEmail
	newGrievance.Status = "Pending"
	newGrievance.SubmittedDate = time.Now()
	newGrievance.LastUpdate = time.Now()

	// 3. Thread-safe append and save
	grievanceFileLock.Lock()
	currentGrievances, _ := readGrievancesFromDisk()
	currentGrievances = append(currentGrievances, newGrievance)
	err := saveGrievancesToDisk(currentGrievances)
	grievanceFileLock.Unlock()

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
	if isGrievanceAdmin(c) {
		actorID = "admin"
		actorName = c.GetString("username")
		if actorName == "" {
			actorName = "Admin"
		}
	} else {
		memberID := c.GetString("member_id")
		if memberID != "" {
			actorID = getMemberTagaIdByUUID(memberID)
		} else {
			actorID = "member"
		}
		actorName = memberEmail
		if actorName == "" {
			actorName = memberName
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
	grievanceFileLock.Lock()
	allGrievances, err := readGrievancesFromDisk()
	if err != nil {
		grievanceFileLock.Unlock()
		c.JSON(http.StatusOK, []model.Grievance{})
		return
	}

	// Auto-Archive logic (older than 6 months = ~4380 hours)
	var activeGrievances []model.Grievance
	var archivedGrievances []model.Grievance

	now := time.Now()
	for _, g := range allGrievances {
		if g.Status == "Read" && now.Sub(g.SubmittedDate).Hours() > 4380 {
			archivedGrievances = append(archivedGrievances, g)
		} else {
			activeGrievances = append(activeGrievances, g)
		}
	}

	// If we archived anything, save to both files
	if len(archivedGrievances) > 0 {
		var existingOld []model.Grievance
		oldFile, _ := os.ReadFile(oldGrievanceFilePath)
		if len(oldFile) > 0 {
			_ = json.Unmarshal(oldFile, &existingOld)
		}
		existingOld = append(existingOld, archivedGrievances...)

		oldData, _ := json.MarshalIndent(existingOld, "", "  ")
		_ = os.WriteFile(oldGrievanceFilePath, oldData, 0644)

		_ = saveGrievancesToDisk(activeGrievances)
		allGrievances = activeGrievances
	} else {
		allGrievances = activeGrievances
	}
	grievanceFileLock.Unlock()

	// IDOR / Privacy Protection:
	// Admins can see all grievances
	if isGrievanceAdmin(c) {
		c.JSON(http.StatusOK, allGrievances)
		return
	}

	// Regular members can only see their own grievances
	callerEmail := getCallerEmail(c)
	var memberGrievances []model.Grievance
	for _, g := range allGrievances {
		if callerEmail != "" && strings.EqualFold(g.MemberEmail, callerEmail) {
			memberGrievances = append(memberGrievances, g)
		}
	}
	if memberGrievances == nil {
		memberGrievances = []model.Grievance{}
	}
	c.JSON(http.StatusOK, memberGrievances)
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

	grievanceFileLock.RLock()
	allGrievances, err := readGrievancesFromDisk()
	grievanceFileLock.RUnlock()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read grievances"})
		return
	}

	for _, g := range allGrievances {
		if g.ID == id {
			// Authorization check: Admins can see any; members can only see their own
			if !isGrievanceAdmin(c) {
				callerEmail := getCallerEmail(c)
				if callerEmail == "" || !strings.EqualFold(g.MemberEmail, callerEmail) {
					c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this grievance"})
					return
				}
			}
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

	grievanceFileLock.Lock()
	defer grievanceFileLock.Unlock()

	allGrievances, err := readGrievancesFromDisk()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read grievances"})
		return
	}

	for i, g := range allGrievances {
		if g.ID == id {
			isAdmin := isGrievanceAdmin(c)
			callerEmail := getCallerEmail(c)

			// IDOR protection: only admin or the grievance author can update
			if !isAdmin && (callerEmail == "" || !strings.EqualFold(g.MemberEmail, callerEmail)) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this grievance"})
				return
			}

			// Admin-only fields (status transition, assignment)
			if isAdmin {
				if updated.Status != "" {
					allGrievances[i].Status = updated.Status
				}
				if updated.AssignedTo != "" {
					allGrievances[i].AssignedTo = updated.AssignedTo
				}
			}

			// General grievance content fields
			if updated.Subject != "" {
				allGrievances[i].Subject = updated.Subject
			}
			if updated.Category != "" {
				allGrievances[i].Category = updated.Category
			}
			if updated.Priority != "" {
				allGrievances[i].Priority = updated.Priority
			}
			if updated.Description != "" {
				allGrievances[i].Description = updated.Description
			}
			if updated.ContactPhone != "" {
				allGrievances[i].ContactPhone = updated.ContactPhone
			}
			if updated.PreferredResponse != "" {
				allGrievances[i].PreferredResponse = updated.PreferredResponse
			}
			allGrievances[i].LastUpdate = time.Now()

			if err := saveGrievancesToDisk(allGrievances); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save grievance"})
				return
			}

			actorID := ""
			actorName := ""
			if isAdmin {
				actorID = "admin"
				actorName = c.GetString("username")
				if actorName == "" {
					actorName = "Admin"
				}
			} else {
				memberID := c.GetString("member_id")
				if memberID != "" {
					actorID = getMemberTagaIdByUUID(memberID)
				} else {
					actorID = "member"
				}
				actorName = callerEmail
			}

			_ = audit.Log(c, actorID, actorName,
				audit.ActionUpdate, audit.ModuleGrievance,
				"grievance", allGrievances[i].ID,
				fmt.Sprintf("Updated grievance status to: %s", allGrievances[i].Status),
				g, allGrievances[i])

			c.JSON(http.StatusOK, allGrievances[i])
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

	grievanceFileLock.Lock()
	defer grievanceFileLock.Unlock()

	allGrievances, err := readGrievancesFromDisk()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read grievances"})
		return
	}

	for i, g := range allGrievances {
		if g.ID == id {
			isAdmin := isGrievanceAdmin(c)
			callerEmail := getCallerEmail(c)

			// IDOR protection: only admin or the grievance author can delete
			if !isAdmin && (callerEmail == "" || !strings.EqualFold(g.MemberEmail, callerEmail)) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this grievance"})
				return
			}

			// Remove from slice
			allGrievances = append(allGrievances[:i], allGrievances[i+1:]...)

			if err := saveGrievancesToDisk(allGrievances); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}

			actorID := ""
			actorName := ""
			if isAdmin {
				actorID = "admin"
				actorName = c.GetString("username")
				if actorName == "" {
					actorName = "Admin"
				}
			} else {
				memberID := c.GetString("member_id")
				if memberID != "" {
					actorID = getMemberTagaIdByUUID(memberID)
				} else {
					actorID = "member"
				}
				actorName = callerEmail
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
