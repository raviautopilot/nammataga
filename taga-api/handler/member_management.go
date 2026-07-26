package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service/member"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AddMemberRequest - Define this type
type AddMemberRequest struct {
	ID                       string `json:"id"`
	TagaID                   string `json:"tagaId" binding:"required"`
	Name                     string `json:"name" binding:"required"`
	Initial                  string `json:"initial" binding:"required"`
	Gender                   string `json:"gender" binding:"required"`
	FatherName               string `json:"fatherName"`
	MotherName               string `json:"motherName"`
	EducationalQualification string `json:"educationalQualification" binding:"required"`
	Designation              string `json:"designation" binding:"required"`
	WorkingDistrict          string `json:"workingDistrict" binding:"required"`
	NativeDistrict           string `json:"nativeDistrict" binding:"required"`
	RecruitmentBatch         string `json:"recruitmentBatch"`
	SeniorityNumber          string `json:"seniorityNumber"`
	ResidentialAddress       string `json:"residentialAddress"`
	PermanentAddress         string `json:"permanentAddress"`
	DateOfBirth              string `json:"dateOfBirth" binding:"required"`
	MobileNumber             string `json:"mobileNumber" binding:"required"`
	Email                    string `json:"email" binding:"required"`
	TbfNumber                string `json:"tbfNumber"`
	CpsGpfNumber             string `json:"cpsGpfNumber"`
}

// UpdateMemberRequest - fields editable by admin (no password, no TAGA ID)
type UpdateMemberRequest struct {
	Name                     string `json:"name"`
	Initial                  string `json:"initial"`
	Gender                   string `json:"gender"`
	FatherName               string `json:"father_name"`
	MotherName               string `json:"mother_name"`
	EducationalQualification string `json:"educational_qualification"`
	Designation              string `json:"designation"`
	WorkingDistrict          string `json:"working_district"`
	NativeDistrict           string `json:"native_district"`
	RecruitmentBatch         string `json:"recruitment_batch"`
	SeniorityNumber          string `json:"seniority_number"`
	ResidentialAddress       string `json:"residential_address"`
	PermanentAddress         string `json:"permanent_address"`
	DateOfBirth              string `json:"date_of_birth"`
	MobileNumber             string `json:"mobile_number"`
	EmailId                  string `json:"emailId"`
	TbfNumber                string `json:"tbf_number"`
	CpsGpfNumber             string `json:"cps_gpf_number"`
}

// AddMember godoc
// @Summary Add a new member
// @Description Admin adds a new member with a generated temporary password
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member body AddMemberRequest true "Member details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/members/add [post]
func AddMember(c *gin.Context) {
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Read existing members
	existingMembers, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	// Check if email already exists
	for _, m := range existingMembers {
		if email, ok := m["emailId"].(string); ok && email == req.Email {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
			return
		}
	}

	// Check if mobile number already exists
	for _, m := range existingMembers {
		if mobile, ok := m["mobile_number"].(string); ok && mobile == req.MobileNumber {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number already registered"})
			return
		}
	}

	// Generate temporary password
	tempPassword := generateTempPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

	// Create new member
	newMember := model.Member{
		ID:                       uuid.New().String(),
		TagaID:                   req.TagaID,
		Username:                 req.Email,
		Name:                     req.Name,
		Initial:                  req.Initial,
		Gender:                   req.Gender,
		FatherName:               req.FatherName,
		MotherName:               req.MotherName,
		EducationalQualification: req.EducationalQualification,
		Designation:              req.Designation,
		WorkingDistrict:          req.WorkingDistrict,
		NativeDistrict:           req.NativeDistrict,
		RecruitmentBatch:         req.RecruitmentBatch,
		SeniorityNumber:          req.SeniorityNumber,
		ResidentialAddress:       req.ResidentialAddress,
		PermanentAddress:         req.PermanentAddress,
		DateOfBirth:              req.DateOfBirth,
		MobileNumber:             req.MobileNumber,
		EmailId:                  req.Email,
		TbfNumber:                req.TbfNumber,
		CpsGpfNumber:             req.CpsGpfNumber,
		Password:                 string(hashedPassword),
		FirstLogin:               true,
		CreatedAt:                time.Now().Format(time.RFC3339),
		PaymentStatus:            "Unpaid",
		SubscriptionActive:       false,
	}

	if err := member.SaveMember(newMember); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save member"})
		return
	}

	go sendSuccessEmail(req.Email, tempPassword)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Member added successfully",
		"member_id":     newMember.ID,
		"temp_password": tempPassword,
	})
}

// UpdateMember godoc
// @Summary Update member details
// @Description Admin updates member details (all fields except password and TAGA ID)
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Param member body UpdateMemberRequest true "Updated member fields"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/members/{id} [put]
func UpdateMember(c *gin.Context) {
	memberID := c.Param("id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Member ID is required"})
		return
	}

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		config.Logger.Error("Failed to read members file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	// Unmarshal into generic map to preserve all existing fields
	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		config.Logger.Error("Failed to parse members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse members"})
		return
	}

	found := false
	for i, m := range members {
		id, _ := m["id"].(string)
		if id != memberID {
			continue
		}

		// Apply only non-empty updates (skip blank strings so existing data is preserved)
		if req.Name != "" {
			members[i]["name"] = req.Name
		}
		if req.Initial != "" {
			members[i]["initial"] = req.Initial
		}
		if req.Gender != "" {
			members[i]["gender"] = req.Gender
		}
		if req.FatherName != "" {
			members[i]["father_name"] = req.FatherName
		}
		if req.MotherName != "" {
			members[i]["mother_name"] = req.MotherName
		}
		if req.EducationalQualification != "" {
			members[i]["educational_qualification"] = req.EducationalQualification
		}
		if req.Designation != "" {
			members[i]["designation"] = req.Designation
		}
		if req.WorkingDistrict != "" {
			members[i]["working_district"] = req.WorkingDistrict
		}
		if req.NativeDistrict != "" {
			members[i]["native_district"] = req.NativeDistrict
		}
		if req.RecruitmentBatch != "" {
			members[i]["recruitment_batch"] = req.RecruitmentBatch
		}
		if req.SeniorityNumber != "" {
			members[i]["seniority_number"] = req.SeniorityNumber
		}
		if req.ResidentialAddress != "" {
			members[i]["residential_address"] = req.ResidentialAddress
		}
		if req.PermanentAddress != "" {
			members[i]["permanent_address"] = req.PermanentAddress
		}
		if req.DateOfBirth != "" {
			members[i]["date_of_birth"] = req.DateOfBirth
		}
		if req.MobileNumber != "" {
			members[i]["mobile_number"] = req.MobileNumber
		}
		if req.EmailId != "" {
			members[i]["emailId"] = req.EmailId
			// Keep username in sync with email
			members[i]["username"] = req.EmailId
		}
		if req.TbfNumber != "" {
			members[i]["tbf_number"] = req.TbfNumber
		}
		if req.CpsGpfNumber != "" {
			members[i]["cps_gpf_number"] = req.CpsGpfNumber
		}

		members[i]["updated_at"] = time.Now().Format(time.RFC3339)
		found = true
		break
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	updatedData, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		config.Logger.Error("Failed to marshal members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	if err := os.WriteFile(cfg.MembersFile, updatedData, 0644); err != nil {
		config.Logger.Error("Failed to write members file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	config.Logger.Info("Member updated by admin", zap.String("member_id", memberID))
	c.JSON(http.StatusOK, gin.H{"message": "Member updated successfully"})
}

// DeleteMember godoc
// @Summary Delete a member permanently
// @Description Admin hard-deletes a member by ID (removes from JSON)
// @Tags Admin Members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/members/{id} [delete]
func DeleteMember(c *gin.Context) {
	memberID := c.Param("id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Member ID is required"})
		return
	}

	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		config.Logger.Error("Failed to read members file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		config.Logger.Error("Failed to parse members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse members"})
		return
	}

	found := false
	var deletedName string
	filtered := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		id, _ := m["id"].(string)
		if id == memberID {
			found = true
			deletedName, _ = m["name"].(string)
			// Skip this member — hard delete
			continue
		}
		filtered = append(filtered, m)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	updatedData, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		config.Logger.Error("Failed to marshal members after delete", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	if err := os.WriteFile(cfg.MembersFile, updatedData, 0644); err != nil {
		config.Logger.Error("Failed to write members file after delete", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	// Also clean up related subscription records (best-effort, no failure on error)
	go cleanupMemberSubscriptions(memberID)

	config.Logger.Info("Member permanently deleted by admin",
		zap.String("member_id", memberID),
		zap.String("member_name", deletedName))

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Member '%s' deleted permanently", deletedName),
	})
}

// cleanupMemberSubscriptions removes subscription records for a deleted member (best-effort)
func cleanupMemberSubscriptions(memberID string) {
	subsPath := filepath.Join("data", "subscriptions", "member_subscriptions.json")
	data, err := os.ReadFile(subsPath)
	if err != nil {
		return
	}

	var subs []map[string]interface{}
	if err := json.Unmarshal(data, &subs); err != nil {
		return
	}

	filtered := make([]map[string]interface{}, 0, len(subs))
	for _, s := range subs {
		id, _ := s["member_id"].(string)
		if id != memberID {
			filtered = append(filtered, s)
		}
	}

	updatedData, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(subsPath, updatedData, 0644)
}
