package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/audit"
	"taga-api/service/member"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"github.com/xuri/excelize/v2"
)

// Data Transfer Objects
type RegistrationRequest struct {
	SNo                      int    `json:"sno"`
	TagaID                   string `json:"tagaId"`
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
	EmailId                  string `json:"email_id"`
	TBFNumber                string `json:"tbf_number"`
	CPSGPFNumber             string `json:"cps_gpf_number"`
	PaymentStatus            string `json:"payment_status"`
}

type RegistrationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

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
	PaymentStatus            string `json:"paymentStatus"`
	PaymentStatusSnake       string `json:"payment_status"`
}

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
	PaymentStatus            string `json:"payment_status"`
	PaymentStatusCamel       string `json:"paymentStatus"`
}

type MemberListItem struct {
	ID               string `json:"id"`
	TagaID           string `json:"tagaId"`
	Name             string `json:"name"`
	Initial          string `json:"initial"`
	Gender           string `json:"gender"`
	District         string `json:"district"`
	Designation      string `json:"designation"`
	RecruitmentBatch string `json:"recruitment_batch"`
	MobileNumber     string `json:"mobile_number"`
	Email            string `json:"email"`
	PaymentStatus    string `json:"payment_status"`
	MembershipStatus string `json:"membership_status"`
}

type MemberListResponse struct {
	Members    []MemberListItem `json:"members"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type BulkUploadResult struct {
	SuccessCount int                 `json:"success_count"`
	FailedCount  int                 `json:"failed_count"`
	TotalCount   int                 `json:"total_count"`
	Failed       []BulkUploadFailure `json:"failed"`
	ProcessedAt  string              `json:"processed_at"`
}

type BulkUploadFailure struct {
	RowNumber int      `json:"row_number"`
	Email     string   `json:"email"`
	Errors    []string `json:"errors"`
}

// HandleMemberRegistration Swagger godoc
// @Summary Handle new member registration
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Registration file (JSON/CSV format)"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/member-registration [post]
func HandleMemberRegistration(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "File upload failed")
		return
	}

	if !isValidFileFormat(file.Filename) {
		respondError(c, http.StatusBadRequest, "Invalid file format. Only JSON or CSV allowed")
		return
	}

	registrations, err := parseRegistrationFile(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse file", "details": err.Error()})
		return
	}

	existingMembers, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read existing members", "details": err.Error()})
		return
	}

	var allErrors []map[string]interface{}
	var validRegistrations []RegistrationRequest

	for idx, reg := range registrations {
		errors := validateRegistration(&reg, existingMembers)
		if len(errors) > 0 {
			allErrors = append(allErrors, map[string]interface{}{
				"index":  idx,
				"email":  reg.EmailId,
				"errors": errors,
			})
		} else {
			validRegistrations = append(validRegistrations, reg)
		}
	}

	if len(allErrors) > 0 {
		if err := sendErrorEmailToAdmin(allErrors); err != nil {
			config.Logger.Error("Failed to send error email to admin", zap.Error(err))
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "Registration validation failed",
			"details": allErrors,
		})
		return
	}

	var results []map[string]interface{}
	for _, reg := range validRegistrations {
		tempPassword := generateTempPassword()

		if err := storeMemberDetails(&reg, tempPassword, existingMembers); err != nil {
			config.Logger.Error("Failed to store member details", zap.String("email", reg.EmailId), zap.Error(err))
			results = append(results, map[string]interface{}{
				"email":  reg.EmailId,
				"status": "failed",
				"error":  err.Error(),
			})
			continue
		}

		if err := sendSuccessEmail(reg.EmailId, tempPassword); err != nil {
			config.Logger.Error("Failed to send registration success email", zap.String("email", reg.EmailId), zap.Error(err))
		}

		results = append(results, map[string]interface{}{
			"email":  reg.EmailId,
			"status": "success",
		})
	}

	respondOK(c, gin.H{
		"message": "Member registration completed",
		"total":   len(validRegistrations),
		"results": results,
	})
}

// AddMember godoc
// @Summary Add a new member
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member body AddMemberRequest true "Member details"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/members/add [post]
func AddMember(c *gin.Context) {
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	existingMembers, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	for _, m := range existingMembers {
		if email, ok := m["emailId"].(string); ok && email == req.Email {
			respondError(c, http.StatusBadRequest, "Email already registered")
			return
		}
		if mobile, ok := m["mobile_number"].(string); ok && mobile == req.MobileNumber {
			respondError(c, http.StatusBadRequest, "Mobile number already registered")
			return
		}
	}

	tempPassword := generateTempPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

	paymentStatus := req.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = req.PaymentStatusSnake
	}
	if paymentStatus == "" {
		paymentStatus = "Unpaid"
	}
	dob := req.DateOfBirth
	if dob == "" {
		dob = "1995-10-01"
	}
	isPaid := strings.EqualFold(paymentStatus, "Paid")

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
		DateOfBirth:              dob,
		MobileNumber:             req.MobileNumber,
		EmailId:                  req.Email,
		TbfNumber:                req.TbfNumber,
		CpsGpfNumber:             req.CpsGpfNumber,
		Password:                 string(hashedPassword),
		FirstLogin:               true,
		CreatedAt:                time.Now().Format(time.RFC3339),
		PaymentStatus:            paymentStatus,
		SubscriptionActive:       isPaid,
	}

	if err := member.SaveMember(newMember); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save member")
		return
	}

	if isPaid {
		go createManualAnnualSubscription(newMember.ID, newMember.EmailId, newMember.Name)
	}

	go sendSuccessEmail(req.Email, tempPassword)

	// Audit member creation — do not include the temp password
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionCreate, audit.ModuleMember,
		"member", newMember.TagaID,
		fmt.Sprintf("Admin created new member %s (tagaId: %s)", newMember.Name, newMember.TagaID),
		nil, map[string]interface{}{
			"tagaId":   newMember.TagaID,
			"name":     newMember.Name,
			"emailId":  newMember.EmailId,
			"district": newMember.WorkingDistrict,
		})

	respondOK(c, gin.H{
		"message":       "Member added successfully",
		"member_id":     newMember.ID,
		"temp_password": tempPassword,
	})
}

// UpdateMember godoc
// @Summary Update member details
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Param member body UpdateMemberRequest true "Updated member fields"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/members/{id} [put]
func UpdateMember(c *gin.Context) {
	memberID := c.Param("id")
	if memberID == "" {
		respondError(c, http.StatusBadRequest, "Member ID is required")
		return
	}

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to parse members")
		return
	}

	found := false
	var oldState map[string]interface{}
	for i, m := range members {
		id, _ := m["id"].(string)
		if id != memberID {
			continue
		}
		// Capture old state before modification
		oldCopy := make(map[string]interface{}, len(m))
		for k, v := range m {
			oldCopy[k] = v
		}
		oldState = oldCopy

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
			members[i]["username"] = req.EmailId
		}
		if req.TbfNumber != "" {
			members[i]["tbf_number"] = req.TbfNumber
		}
		if req.CpsGpfNumber != "" {
			members[i]["cps_gpf_number"] = req.CpsGpfNumber
		}
		pStatus := req.PaymentStatus
		if pStatus == "" {
			pStatus = req.PaymentStatusCamel
		}
		if pStatus != "" {
			members[i]["payment_status"] = pStatus
			members[i]["subscription_active"] = strings.EqualFold(pStatus, "Paid")
		}

		members[i]["updated_at"] = time.Now().Format(time.RFC3339)
		found = true
		break
	}

	if !found {
		respondError(c, http.StatusNotFound, "Member not found")
		return
	}

	updatedData, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save members")
		return
	}

	if err := os.WriteFile(cfg.MembersFile, updatedData, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save members")
		return
	}

	// Audit member update with old/new diff
	resourceTagaID := ""
	if oldState != nil {
		resourceTagaID, _ = oldState["tagaId"].(string)
	}
	var newState map[string]interface{}
	for _, m := range members {
		if id, _ := m["id"].(string); id == memberID {
			newState = m
			break
		}
	}
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionUpdate, audit.ModuleMember,
		"member", resourceTagaID,
		fmt.Sprintf("Admin updated member %s (tagaId: %s)", memberID, resourceTagaID),
		audit.Sanitize(oldState), audit.Sanitize(newState))

	respondMessage(c, "Member updated successfully")
}

// DeleteMember godoc
// @Summary Delete a member permanently
// @Tags Admin Members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/members/{id} [delete]
func DeleteMember(c *gin.Context) {
	memberID := c.Param("id")
	if memberID == "" {
		respondError(c, http.StatusBadRequest, "Member ID is required")
		return
	}

	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to parse members")
		return
	}

	found := false
	var deletedMember map[string]interface{}
	var deletedName string
	filtered := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		id, _ := m["id"].(string)
		if id == memberID {
			found = true
			deletedName, _ = m["name"].(string)
			// Capture the member before deletion for audit
			deletedMember = m
			continue
		}
		filtered = append(filtered, m)
	}

	if !found {
		respondError(c, http.StatusNotFound, "Member not found")
		return
	}

	updatedData, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save members")
		return
	}

	if err := os.WriteFile(cfg.MembersFile, updatedData, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save members")
		return
	}

	// Audit member deletion
	resourceTagaID := ""
	if deletedMember != nil {
		resourceTagaID, _ = deletedMember["tagaId"].(string)
	}
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionDelete, audit.ModuleMember,
		"member", resourceTagaID,
		fmt.Sprintf("Admin deleted member '%s' (tagaId: %s)", deletedName, resourceTagaID),
		audit.Sanitize(deletedMember), nil)

	// Cleanups and backups
	memberEmail := getString(deletedMember, "emailId")
	if memberEmail == "" {
		memberEmail = getString(deletedMember, "email") // fallback
	}
	
	memberTagaID := getString(deletedMember, "tagaId")
	
	go cleanupMemberSubscriptions(memberID, memberTagaID)
	go cleanupPaymentTransactions(memberID, memberTagaID)
	go cleanupMemberPayments(memberEmail)
	go cleanupMemberNotifications(memberID, memberTagaID)
	go cleanupMembershipApplications(memberEmail)
	go saveDeletedMember(deletedMember)
	
	respondMessage(c, fmt.Sprintf("Member '%s' deleted permanently", deletedName))
}

func cleanupMemberNotifications(memberID, memberTagaID string) {
	if memberID == "" && memberTagaID == "" {
		return
	}
	notificationsFile := filepath.Join("data", "notifications", "member_notifications.json")
	data, err := os.ReadFile(notificationsFile)
	if err != nil {
		return
	}

	var notifs []map[string]interface{}
	if err := json.Unmarshal(data, &notifs); err != nil {
		return
	}

	filtered := make([]map[string]interface{}, 0, len(notifs))
	for _, n := range notifs {
		id, _ := n["member_id"].(string)
		if id != memberID && id != memberTagaID {
			filtered = append(filtered, n)
		}
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	os.WriteFile(notificationsFile, updatedData, 0644)
}

func cleanupMembershipApplications(memberEmail string) {
	if memberEmail == "" {
		return
	}
	membershipFile := filepath.Join("data", "memberlogin", "membership.json")
	data, err := os.ReadFile(membershipFile)
	if err != nil {
		return
	}

	var apps []map[string]interface{}
	if err := json.Unmarshal(data, &apps); err != nil {
		return
	}

	filtered := make([]map[string]interface{}, 0, len(apps))
	for _, a := range apps {
		email, _ := a["email"].(string)
		if email != memberEmail {
			filtered = append(filtered, a)
		}
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	os.WriteFile(membershipFile, updatedData, 0644)
}

func saveDeletedMember(member map[string]interface{}) {
	if member == nil {
		return
	}
	deletedFile := config.Config.DeletedMembersFile
	var deletedMembers []map[string]interface{}

	data, err := os.ReadFile(deletedFile)
	if err == nil {
		json.Unmarshal(data, &deletedMembers)
	}

	member["deleted_at"] = time.Now().Format(time.RFC3339)
	deletedMembers = append(deletedMembers, member)

	updatedData, _ := json.MarshalIndent(deletedMembers, "", "  ")
	os.WriteFile(deletedFile, updatedData, 0644)
}

func cleanupMemberPayments(memberEmail string) {
	if memberEmail == "" {
		return
	}
	paymentsFile := config.Config.ProcessedPaymentsFile
	data, err := os.ReadFile(paymentsFile)
	if err != nil {
		return
	}

	var payments []map[string]interface{}
	if err := json.Unmarshal(data, &payments); err != nil {
		return
	}

	filtered := make([]map[string]interface{}, 0, len(payments))
	for _, p := range payments {
		email, _ := p["member_email"].(string)
		if email != memberEmail {
			filtered = append(filtered, p)
		}
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	os.WriteFile(paymentsFile, updatedData, 0644)
}

func cleanupPaymentTransactions(memberID, memberTagaID string) {
	if memberID == "" && memberTagaID == "" {
		return
	}
	txnsPath := filepath.Join("data", "subscriptions", "payment_transactions.json")
	data, err := os.ReadFile(txnsPath)
	if err != nil {
		return
	}

	var txns []map[string]interface{}
	if err := json.Unmarshal(data, &txns); err != nil {
		return
	}

	filtered := make([]map[string]interface{}, 0, len(txns))
	for _, t := range txns {
		id, _ := t["member_id"].(string)
		if id != memberID && id != memberTagaID {
			filtered = append(filtered, t)
		}
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	os.WriteFile(txnsPath, updatedData, 0644)
}

func cleanupMemberSubscriptions(memberID, memberTagaID string) {
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
		if id != memberID && id != memberTagaID {
			filtered = append(filtered, s)
		}
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	_ = os.WriteFile(subsPath, updatedData, 0644)
}

// BulkUploadMembers godoc
// @Summary Bulk upload members
// @Tags Admin Members
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV or Excel file (.csv, .xlsx, .xls)"
// @Success 200 {object} BulkUploadResult
// @Router /api/admin/members/bulk-upload [post]
func BulkUploadMembers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "File is required")
		return
	}

	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		respondError(c, http.StatusBadRequest, "File too large. Maximum 10MB allowed")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	supportedExts := map[string]bool{".csv": true, ".xlsx": true, ".xls": true}
	if !supportedExts[ext] {
		respondError(c, http.StatusBadRequest, "Only CSV and Excel files (.csv, .xlsx, .xls) are allowed")
		return
	}

	tempDir, err := os.MkdirTemp("", "bulk_upload_*")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to process upload")
		return
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save file")
		return
	}

	var registrations []RegistrationRequest
	if ext == ".csv" {
		registrations, err = parseCSVFileForBulk(tempPath)
	} else {
		registrations, err = parseExcelFileForBulk(tempPath)
	}

	if err != nil {
		respondError(c, http.StatusBadRequest, "Failed to parse file: "+err.Error())
		return
	}

	if len(registrations) == 0 {
		respondError(c, http.StatusBadRequest, "No valid records found in file")
		return
	}

	if len(registrations) > 500 {
		respondError(c, http.StatusBadRequest, "Maximum 500 members per upload allowed")
		return
	}

	existingMembers, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read member database")
		return
	}

	existingEmailMap := make(map[string]bool)
	existingMobileMap := make(map[string]bool)
	existingTagaIDMap := make(map[string]bool)

	for _, m := range existingMembers {
		if email, ok := m["emailId"].(string); ok {
			existingEmailMap[strings.ToLower(email)] = true
		}
		if mobile, ok := m["mobile_number"].(string); ok && mobile != "" {
			existingMobileMap[mobile] = true
		}
		if tagaID, ok := m["tagaId"].(string); ok && tagaID != "" {
			existingTagaIDMap[tagaID] = true
		}
	}

	result := BulkUploadResult{
		ProcessedAt: time.Now().Format(time.RFC3339),
		TotalCount:  len(registrations),
		Failed:      []BulkUploadFailure{},
	}

	var mu sync.Mutex
	for idx, reg := range registrations {
		var errors []string
		if existingTagaIDMap[reg.TagaID] {
			errors = append(errors, "TAGA ID already exists")
		}

		validationErrors := validateRegistrationForBulk(&reg, existingEmailMap, existingMobileMap)
		errors = append(errors, validationErrors...)

		if len(errors) > 0 {
			mu.Lock()
			result.FailedCount++
			result.Failed = append(result.Failed, BulkUploadFailure{
				RowNumber: idx + 2,
				Email:     reg.EmailId,
				Errors:    errors,
			})
			mu.Unlock()
			continue
		}

		existingTagaIDMap[reg.TagaID] = true
		tempPassword := generateTempPassword()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

		// Derive payment status from the uploaded file value (mirrors AddMember logic).
		paymentStatus := strings.TrimSpace(reg.PaymentStatus)
		if paymentStatus == "" {
			paymentStatus = "Unpaid"
		}
		isPaid := strings.EqualFold(paymentStatus, "Paid")

		dob := reg.DateOfBirth
		if dob == "" {
			dob = "1995-10-01"
		}

		newMember := map[string]interface{}{
			"id":                        uuid.New().String(),
			"tagaId":                    reg.TagaID,
			"username":                  reg.EmailId,
			"sno":                       idx + 1,
			"name":                      reg.Name,
			"initial":                   reg.Initial,
			"gender":                    reg.Gender,
			"father_name":               reg.FatherName,
			"mother_name":               reg.MotherName,
			"educational_qualification": reg.EducationalQualification,
			"designation":               reg.Designation,
			"working_district":          reg.WorkingDistrict,
			"native_district":           reg.NativeDistrict,
			"recruitment_batch":         reg.RecruitmentBatch,
			"seniority_number":          reg.SeniorityNumber,
			"residential_address":       reg.ResidentialAddress,
			"permanent_address":         reg.PermanentAddress,
			"date_of_birth":             dob,
			"mobile_number":             reg.MobileNumber,
			"emailId":                   reg.EmailId,
			"tbf_number":                reg.TBFNumber,
			"cps_gpf_number":            reg.CPSGPFNumber,
			"password":                  string(hashedPassword),
			"first_login":               true,
			"created_at":                time.Now().Format(time.RFC3339),
			"payment_status":            paymentStatus,
			"subscription_active":       isPaid,
			"subscription_end_date":     nil,
			"subscription_updated_at":   nil,
		}
		existingMembers = append(existingMembers, newMember)
		existingEmailMap[strings.ToLower(reg.EmailId)] = true
		existingMobileMap[reg.MobileNumber] = true

		if isPaid {
			go createManualAnnualSubscription(newMember["id"].(string), reg.EmailId, reg.Name)
		}

		mu.Lock()
		result.SuccessCount++
		mu.Unlock()

		go sendSuccessEmail(reg.EmailId, tempPassword)
	}

	cfg := config.GetConfig()
	data, err := json.MarshalIndent(existingMembers, "", "  ")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save members")
		return
	}

	if err := os.WriteFile(cfg.MembersFile, data, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save members")
		return
	}

	respondOK(c, result)
}

// GetMembersList godoc
// @Summary Get list of members
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by name, email, mobile"
// @Param district query string false "Filter by district"
// @Param payment_status query string false "Filter by payment status (paid/unpaid)"
// @Success 200 {object} MemberListResponse
// @Router /api/admin/members [get]
func GetMembersList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := strings.ToLower(c.Query("search"))
	districtFilter := c.Query("district")
	paymentFilter := c.Query("payment_status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	subscriptionMap := loadSubscriptionPaymentMap()
	var memberList []MemberListItem

	for _, m := range members {
		email := getString(m, "emailId")

		if districtFilter != "" && districtFilter != "all" {
			mDistrict := getString(m, "working_district")
			if !strings.EqualFold(mDistrict, districtFilter) {
				continue
			}
		}

		paymentStatus := getPaymentStatusFromMember(m, subscriptionMap)
		if paymentFilter != "" && paymentFilter != "all" {
			if !strings.EqualFold(paymentStatus, paymentFilter) {
				continue
			}
		}

		if search != "" {
			cleanSearch := strings.TrimSpace(strings.ToLower(search))
			name := strings.ToLower(getString(m, "name"))
			emailLower := strings.TrimSpace(strings.ToLower(email))
			mobile := strings.TrimSpace(getString(m, "mobile_number"))
			tagaID := strings.TrimSpace(strings.ToLower(getString(m, "tagaId")))
			designation := strings.TrimSpace(strings.ToLower(getString(m, "designation")))

			if !strings.Contains(name, cleanSearch) &&
				!strings.Contains(emailLower, cleanSearch) &&
				!strings.Contains(mobile, cleanSearch) &&
				!strings.Contains(tagaID, cleanSearch) &&
				!strings.Contains(designation, cleanSearch) {
				continue
			}
		}

		memberListItem := MemberListItem{
			ID:               getString(m, "id"),
			TagaID:           getString(m, "tagaId"),
			Name:             getString(m, "name"),
			Initial:          getString(m, "initial"),
			Gender:           getString(m, "gender"),
			District:         getString(m, "working_district"),
			Designation:      getString(m, "designation"),
			RecruitmentBatch: getString(m, "recruitment_batch"),
			MobileNumber:     getString(m, "mobile_number"),
			Email:            email,
			PaymentStatus:    paymentStatus,
			MembershipStatus: getMembershipStatus(m, subscriptionMap),
		}
		memberList = append(memberList, memberListItem)
	}

	total := len(memberList)
	totalPages := (total + limit - 1) / limit
	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}

	var paginatedMembers []MemberListItem
	if start < total {
		paginatedMembers = memberList[start:end]
	} else {
		paginatedMembers = []MemberListItem{}
	}

	respondOK(c, MemberListResponse{
		Members:    paginatedMembers,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetMemberDistricts godoc
// @Summary Get list of districts with member counts
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/admin/members/districts [get]
func GetMemberDistricts(c *gin.Context) {
	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	districtMap := make(map[string]int)
	for _, m := range members {
		district := getString(m, "working_district")
		if district != "" {
			districtMap[district]++
		}
	}

	var districts []map[string]interface{}
	for name, count := range districtMap {
		districts = append(districts, map[string]interface{}{
			"name":  name,
			"count": count,
		})
	}

	respondOK(c, districts)
}

// GetMemberStats godoc
// @Summary Get member statistics
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/members/stats [get]
func GetMemberStats(c *gin.Context) {
	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	subscriptionMap := loadSubscriptionPaymentMap()
	total := 0
	active := 0
	unpaid := 0

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	newThisMonth := 0

	for _, m := range members {
		if deleted, ok := m["deleted"].(bool); ok && deleted {
			continue
		}
		total++

		paymentStatus := getPaymentStatusFromMember(m, subscriptionMap)

		if getMembershipStatus(m, subscriptionMap) == "Active" {
			active++
		}
		if paymentStatus == "Unpaid" {
			unpaid++
		}

		createdAtStr := getString(m, "created_at")
		if createdAtStr != "" {
			if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				y, m, _ := t.Date()
				if y == currentYear && m == currentMonth {
					newThisMonth++
				}
			} else if t, err := time.Parse("2006-01-02", createdAtStr); err == nil {
				y, m, _ := t.Date()
				if y == currentYear && m == currentMonth {
					newThisMonth++
				}
			}
		}
	}

	respondOK(c, gin.H{
		"totalMembers":  total,
		"activeMembers": active,
		"unpaid":        unpaid,
		"newThisMonth":  newThisMonth,
	})
}

// ExportMembersToExcel godoc
// @Summary Export members to Excel
// @Tags Admin Reports
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param district query string false "Filter by district"
// @Param payment_status query string false "Filter by payment status (paid/unpaid)"
// @Success 200 {file} file
// @Router /api/admin/members/export [get]
func ExportMembersToExcel(c *gin.Context) {
	district := c.Query("district")
	paymentStatus := c.Query("payment_status")

	allMembers, err := getAllMembersFromStorage()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch members")
		return
	}

	filteredMembers := make([]map[string]interface{}, 0)
	for _, memberMap := range allMembers {
		if district != "" && district != "all" {
			workingDistrict := getMapString(memberMap, "working_district")
			if !strings.EqualFold(workingDistrict, district) {
				continue
			}
		}

		if paymentStatus != "" && paymentStatus != "all" {
			memberPaymentStatus := getMapString(memberMap, "payment_status")
			if !strings.EqualFold(memberPaymentStatus, paymentStatus) {
				continue
			}
		}

		filteredMembers = append(filteredMembers, memberMap)
	}

	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheetName := "Member Details"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create sheet")
		return
	}
	f.SetActiveSheet(index)

	headers := []string{
		"TAGA ID", "Name", "Initial", "Gender", "District",
		"Designation", "Batch", "Mobile", "Email", "Payment Status", "Membership Status",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#2E7D32"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		if headerStyle != 0 {
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}

	paidCount := 0
	unpaidCount := 0
	activeCount := 0

	for rowIdx, memberMap := range filteredMembers {
		rowNum := rowIdx + 2

		tagaID := getMapString(memberMap, "tagaId")
		name := getMapString(memberMap, "name")
		initial := getMapString(memberMap, "initial")
		gender := getMapString(memberMap, "gender")
		workingDistrict := getMapString(memberMap, "working_district")
		designation := getMapString(memberMap, "designation")
		recruitmentBatch := getMapString(memberMap, "recruitment_batch")
		mobileNumber := getMapString(memberMap, "mobile_number")
		email := getMapString(memberMap, "emailId")
		paymentStatusVal := getMapString(memberMap, "payment_status")
		membershipStatus := getMapString(memberMap, "membership_status")

		if strings.EqualFold(paymentStatusVal, "paid") {
			paidCount++
		} else if strings.EqualFold(paymentStatusVal, "unpaid") {
			unpaidCount++
		}

		if strings.EqualFold(membershipStatus, "active") {
			activeCount++
		}

		rowData := []interface{}{tagaID, name, initial, gender, workingDistrict,
			designation, recruitmentBatch, mobileNumber, email, paymentStatusVal, membershipStatus}

		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%s%d", string(rune('A'+colIdx)), rowNum)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	summarySheet := "Summary"
	_, _ = f.NewSheet(summarySheet)

	f.MergeCell(summarySheet, "A1", "B1")
	f.SetCellValue(summarySheet, "A1", "MEMBER EXPORT SUMMARY")

	summaryData := [][]interface{}{
		{"Export Date", time.Now().Format("2006-01-02 15:04:05")},
		{"", ""},
		{"TOTAL MEMBERS", len(filteredMembers)},
		{"", ""},
		{"PAID MEMBERS", paidCount},
		{"UNPAID MEMBERS", unpaidCount},
		{"", ""},
		{"ACTIVE MEMBERS", activeCount},
		{"", ""},
		{"FILTERS APPLIED", ""},
		{"District", getFilterDisplay(district)},
		{"Payment Status", getFilterDisplay(paymentStatus)},
	}

	for rowIdx, row := range summaryData {
		rowNum := rowIdx + 3
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", rowNum), row[0])
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", rowNum), row[1])
	}

	f.DeleteSheet("Sheet1")
	filename := fmt.Sprintf("TAGA_Members_Export_%s.xlsx", time.Now().Format("2006-01-02_15-04-05"))

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := f.Write(c.Writer); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to generate Excel file")
		return
	}
}

// formatReportDate formats date strings cleanly as YYYY-MM-DD or returns "N/A"
func formatReportDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return "N/A"
	}
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t.Format("2006-01-02")
	}
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t.Format("2006-01-02")
	}
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t.Format("2006-01-02")
	}
	if t, err := time.Parse("02/01/2006", dateStr); err == nil {
		return t.Format("2006-01-02")
	}
	if len(dateStr) >= 10 && dateStr[4] == '-' && dateStr[7] == '-' {
		return dateStr[:10]
	}
	return dateStr
}

func getReportValue(val string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return "N/A"
	}
	return val
}

func getMemberReportTagaID(member map[string]interface{}) string {
	if tagaID := strings.TrimSpace(getString(member, "tagaId")); tagaID != "" {
		return tagaID
	}
	if tagaID := strings.TrimSpace(getString(member, "taga_id")); tagaID != "" {
		return tagaID
	}
	return "N/A"
}

func matchesPeriod(createdAtStr, period string) bool {
	if period == "" || period == "all_time" {
		return true
	}
	createdAtStr = strings.TrimSpace(createdAtStr)
	if createdAtStr == "" {
		return false
	}
	var t time.Time
	var err error
	if t, err = time.Parse(time.RFC3339, createdAtStr); err != nil {
		if t, err = time.Parse(time.RFC3339Nano, createdAtStr); err != nil {
			if t, err = time.Parse("2006-01-02", createdAtStr); err != nil {
				if t, err = time.Parse("02/01/2006", createdAtStr); err != nil {
					return true
				}
			}
		}
	}
	now := time.Now()
	switch period {
	case "current_month":
		return t.Year() == now.Year() && t.Month() == now.Month()
	case "last_month":
		lastMonth := now.AddDate(0, -1, 0)
		return t.Year() == lastMonth.Year() && t.Month() == lastMonth.Month()
	case "current_quarter":
		currentQ := (int(now.Month()) - 1) / 3
		tQ := (int(t.Month()) - 1) / 3
		return t.Year() == now.Year() && currentQ == tQ
	case "current_year":
		return t.Year() == now.Year()
	default:
		return true
	}
}

// GenerateMemberReport godoc
// @Summary Generate member report
// @Tags Admin Reports
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param report_type query string false "Report type (membership, financial, etc.)" default(membership)
// @Param period query string false "Period filter (current_month, last_month, current_quarter, current_year, all_time)" default(all_time)
// @Success 200 {file} file
// @Router /api/admin/reports/members [get]
func GenerateMemberReport(c *gin.Context) {
	reportType := c.DefaultQuery("report_type", "membership")
	period := c.DefaultQuery("period", "all_time")

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	var filteredMembers []map[string]interface{}
	for _, member := range members {
		if matchesPeriod(getString(member, "created_at"), period) {
			filteredMembers = append(filteredMembers, member)
		}
	}

	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheetName := "Membership Report"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create sheet")
		return
	}
	f.SetActiveSheet(index)

	headers := []string{
		"TAGA ID", "Name", "Initial", "Gender", "Father Name", "Mother Name",
		"Educational Qualification", "Designation", "Working District", "Native District",
		"Recruitment Batch", "Seniority Number", "Mobile Number", "Email ID",
		"TBF Number", "CPS/GPF Number", "Registration Date",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#2E7D32"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
		if headerStyle != 0 {
			_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}

	for rowIdx, member := range filteredMembers {
		rowNum := rowIdx + 2

		tagaID := getMemberReportTagaID(member)
		regDate := formatReportDate(getString(member, "created_at"))

		rowValues := []interface{}{
			tagaID,
			getReportValue(getString(member, "name")),
			getReportValue(getString(member, "initial")),
			getReportValue(getString(member, "gender")),
			getReportValue(getString(member, "father_name")),
			getReportValue(getString(member, "mother_name")),
			getReportValue(getString(member, "educational_qualification")),
			getReportValue(getString(member, "designation")),
			getReportValue(getString(member, "working_district")),
			getReportValue(getString(member, "native_district")),
			getReportValue(getString(member, "recruitment_batch")),
			getReportValue(getString(member, "seniority_number")),
			getReportValue(getString(member, "mobile_number")),
			getReportValue(getString(member, "emailId")),
			getReportValue(getString(member, "tbf_number")),
			getReportValue(getString(member, "cps_gpf_number")),
			regDate,
		}

		for colIdx, val := range rowValues {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			_ = f.SetCellValue(sheetName, cell, val)
		}
	}

	// Set column widths
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheetName, colName, colName, 20)
	}

	_ = f.DeleteSheet("Sheet1")

	filename := fmt.Sprintf("TAGA_%s_report_%s_%s.xlsx", reportType, period, time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := f.Write(c.Writer); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to generate Excel file")
		return
	}
}

// Helpers for registration and bulk parsing
func isValidFileFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".json" || ext == ".csv"
}

func parseRegistrationFile(file *multipart.FileHeader) ([]RegistrationRequest, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".json":
		return parseJSONFile(src)
	case ".csv":
		return parseCSVFile(src)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func parseJSONFile(reader io.Reader) ([]RegistrationRequest, error) {
	var registrations []RegistrationRequest
	if err := json.NewDecoder(reader).Decode(&registrations); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return registrations, nil
}

func parseCSVFile(reader io.Reader) ([]RegistrationRequest, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must contain header and at least one record")
	}

	getField := func(row []string, idx int) string {
		if idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	var registrations []RegistrationRequest
	for i := 1; i < len(records); i++ {
		if len(records[i]) < 1 {
			continue
		}
		registrations = append(registrations, RegistrationRequest{
			SNo:                      i,
			TagaID:                   getField(records[i], 0),
			Name:                     getField(records[i], 1),
			Initial:                  getField(records[i], 2),
			Gender:                   getField(records[i], 3),
			FatherName:               getField(records[i], 4),
			MotherName:               getField(records[i], 5),
			EducationalQualification: getField(records[i], 6),
			Designation:              getField(records[i], 7),
			WorkingDistrict:          getField(records[i], 8),
			NativeDistrict:           getField(records[i], 9),
			RecruitmentBatch:         getField(records[i], 10),
			SeniorityNumber:          getField(records[i], 11),
			ResidentialAddress:       getField(records[i], 12),
			PermanentAddress:         getField(records[i], 13),
			DateOfBirth:              getField(records[i], 14),
			MobileNumber:             getField(records[i], 15),
			EmailId:                  getField(records[i], 16),
			TBFNumber:                getField(records[i], 17),
			CPSGPFNumber:             getField(records[i], 18),
			PaymentStatus:            getField(records[i], 19),
		})
	}

	return registrations, nil
}

func parseCSVFileForBulk(filePath string) ([]RegistrationRequest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parseCSVFile(file)
}

func parseExcelFileForBulk(filePath string) ([]RegistrationRequest, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel file has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must contain header and at least one data row")
	}

	getField := func(row []string, idx int) string {
		if idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	var registrations []RegistrationRequest
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || getField(row, 0) == "" {
			continue
		}

		registrations = append(registrations, RegistrationRequest{
			SNo:                      i,
			TagaID:                   getField(row, 0),
			Name:                     getField(row, 1),
			Initial:                  getField(row, 2),
			Gender:                   getField(row, 3),
			FatherName:               getField(row, 4),
			MotherName:               getField(row, 5),
			EducationalQualification: getField(row, 6),
			Designation:              getField(row, 7),
			WorkingDistrict:          getField(row, 8),
			NativeDistrict:           getField(row, 9),
			RecruitmentBatch:         getField(row, 10),
			SeniorityNumber:          getField(row, 11),
			ResidentialAddress:       getField(row, 12),
			PermanentAddress:         getField(row, 13),
			DateOfBirth:              getField(row, 14),
			MobileNumber:             getField(row, 15),
			EmailId:                  getField(row, 16),
			TBFNumber:                getField(row, 17),
			CPSGPFNumber:             getField(row, 18),
			PaymentStatus:            getField(row, 19),
		})
	}

	return registrations, nil
}

func validateRegistration(req *RegistrationRequest, existingMembers []map[string]interface{}) []RegistrationError {
	var errors []RegistrationError
	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, RegistrationError{Field: "name", Message: "Name is required"})
	}

	if strings.TrimSpace(req.EmailId) == "" {
		errors = append(errors, RegistrationError{Field: "email", Message: "Email is required"})
	} else if !isValidEmail(req.EmailId) {
		errors = append(errors, RegistrationError{Field: "email", Message: "Invalid email format"})
	} else {
		for _, member := range existingMembers {
			if email, ok := member["emailId"].(string); ok && strings.EqualFold(email, req.EmailId) {
				errors = append(errors, RegistrationError{Field: "email", Message: "Email is already registered"})
				break
			}
		}
	}

	if strings.TrimSpace(req.MobileNumber) == "" {
		errors = append(errors, RegistrationError{Field: "phone", Message: "Phone is required"})
	} else if !isValidPhone(req.MobileNumber) {
		errors = append(errors, RegistrationError{Field: "phone", Message: "Invalid phone format"})
	} else {
		for _, member := range existingMembers {
			if mobile, ok := member["mobile_number"].(string); ok && mobile == req.MobileNumber {
				errors = append(errors, RegistrationError{Field: "phone", Message: "Mobile number is already registered"})
				break
			}
		}
	}

	return errors
}

func validateRegistrationForBulk(req *RegistrationRequest, existingEmailMap map[string]bool, existingMobileMap map[string]bool) []string {
	var errors []string
	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "Name is required")
	}

	email := strings.TrimSpace(req.EmailId)
	if email == "" {
		errors = append(errors, "Email is required")
	} else if !isValidEmail(email) {
		errors = append(errors, "Invalid email format")
	} else if existingEmailMap[strings.ToLower(email)] {
		errors = append(errors, "Email is already registered")
	}

	mobile := strings.TrimSpace(req.MobileNumber)
	if mobile == "" {
		errors = append(errors, "Mobile number is required")
	} else if !isValidPhone(mobile) {
		errors = append(errors, "Invalid mobile number format")
	} else if existingMobileMap[mobile] {
		errors = append(errors, "Mobile number is already registered")
	}

	return errors
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^[0-9+\-\s()]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

func storeMemberDetails(req *RegistrationRequest, tempPassword string, existingMembers []map[string]interface{}) error {
	cfg := config.GetConfig()
	membersFilePath := cfg.MembersFile

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	dob := req.DateOfBirth
	if dob == "" {
		dob = "1995-10-01"
	}

	newMember := map[string]interface{}{
		"id":                        uuid.New().String(),
		"tagaId":                    req.TagaID,
		"username":                  req.EmailId,
		"sno":                       req.SNo,
		"name":                      req.Name,
		"initial":                   req.Initial,
		"gender":                    req.Gender,
		"father_name":               req.FatherName,
		"mother_name":               req.MotherName,
		"educational_qualification": req.EducationalQualification,
		"designation":               req.Designation,
		"working_district":          req.WorkingDistrict,
		"native_district":           req.NativeDistrict,
		"recruitment_batch":         req.RecruitmentBatch,
		"seniority_number":          req.SeniorityNumber,
		"residential_address":       req.ResidentialAddress,
		"permanent_address":         req.PermanentAddress,
		"date_of_birth":             dob,
		"mobile_number":             req.MobileNumber,
		"emailId":                   req.EmailId,
		"tbf_number":                req.TBFNumber,
		"cps_gpf_number":            req.CPSGPFNumber,
		"password":                  string(hashedPassword),
		"first_login":               true,
		"created_at":                time.Now().Format(time.RFC3339),
		"password_changed_at":       nil,
		"payment_status":            "Unpaid",
		"subscription_active":       false,
		"subscription_end_date":     nil,
		"subscription_updated_at":   nil,
		"deleted":                   false,
		"deleted_at":                nil,
	}

	updatedMembers := append(existingMembers, newMember)
	data, err := json.MarshalIndent(updatedMembers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}

	return os.WriteFile(membersFilePath, data, 0644)
}

func sendErrorEmailToAdmin(errors []map[string]interface{}) error {
	cfg := config.GetConfig()
	adminEmail := cfg.FromEmail
	if adminEmail == "" {
		adminEmail = cfg.AdminEmail
	}
	if adminEmail == "" {
		return fmt.Errorf("admin email not configured")
	}

	subject := "⚠️ TAGA Admin Alert - Member Registration Validation Errors"
	
	var errorRows strings.Builder
	for _, err := range errors {
		var fieldErrors strings.Builder
		if errList, ok := err["errors"].([]RegistrationError); ok {
			for _, e := range errList {
				fieldErrors.WriteString(fmt.Sprintf("<div style='font-size: 13px; color: #991b1b; margin-bottom: 4px;'>&bull; <strong>%s:</strong> %s</div>", html.EscapeString(e.Field), html.EscapeString(e.Message)))
			}
		}

		errorRows.WriteString(fmt.Sprintf(`
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 12px; font-size: 14px; font-weight: 600; color: #1f2937; vertical-align: top; width: 35%%;">%v</td>
                <td style="padding: 12px; vertical-align: top;">%s</td>
            </tr>`, err["email"], fieldErrors.String()))
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Registration Validation Errors</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #dc2626 0%%, #991b1b 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Admin Alert
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 800;">Registration Validation Failed</h1>
            <p style="margin: 0; font-size: 14px; opacity: 0.9;">One or more member entries require attention</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                The following registration requests encountered validation issues during processing:
            </p>

            <table style="width: 100%%; border-collapse: collapse; border: 1px solid #e5e7eb; border-radius: 8px; overflow: hidden; margin-bottom: 24px;">
                <thead>
                    <tr style="background: #f8fafc; border-bottom: 2px solid #e2e8f0;">
                        <th style="padding: 10px 12px; text-align: left; font-size: 13px; color: #64748b; font-weight: 600;">Email</th>
                        <th style="padding: 10px 12px; text-align: left; font-size: 13px; color: #64748b; font-weight: 600;">Validation Errors</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>

            <p style="margin: 0; font-size: 13px; color: #6b7280;">
                Please correct these records in the administration portal or notify the respective applicants.
            </p>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">Automated Admin System &bull; &copy; 2026 TAGA</p>
        </div>
    </div>
</body>
</html>`, errorRows.String())

	return sendEmail(adminEmail, subject, body)
}

func loadSubscriptionPaymentMap() map[string]bool {
	subscriptionMap := make(map[string]bool)

	filePath := filepath.Join("data", "subscriptions", "member_subscriptions.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return subscriptionMap
	}

	var subscriptions []map[string]interface{}
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		return subscriptionMap
	}

	now := time.Now()
	for _, sub := range subscriptions {
		subID, _ := sub["subscription_id"].(string)
		if subID != "annual-subscription" {
			continue
		}
		if email, ok := sub["member_email"].(string); ok {
			if status, ok := sub["status"].(string); ok && status == "active" {
				if endDateStr, ok := sub["end_date"].(string); ok {
					if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil && now.Before(endDate) {
						subscriptionMap[email] = true
					}
				}
			}
		}
	}
	return subscriptionMap
}

func getPaymentStatusFromMember(m map[string]interface{}, subscriptionMap map[string]bool) string {
	if pStatus, ok := m["payment_status"].(string); ok && pStatus != "" {
		if strings.EqualFold(pStatus, "Paid") {
			return "Paid"
		}
		if strings.EqualFold(pStatus, "Unpaid") {
			return "Unpaid"
		}
	}
	if subActive, ok := m["subscription_active"].(bool); ok && subActive {
		return "Paid"
	}
	email := getString(m, "emailId")
	if isPaid, exists := subscriptionMap[email]; exists && isPaid {
		return "Paid"
	}
	return "Unpaid"
}

func getMembershipStatus(member map[string]interface{}, subscriptionMap map[string]bool) string {
	if getPaymentStatusFromMember(member, subscriptionMap) == "Paid" {
		return "Active"
	}
	return "Inactive"
}

func getFilterDisplay(filter string) string {
	if filter == "" || filter == "all" {
		return "None (All members)"
	}
	return filter
}

func getAllMembersFromStorage() ([]map[string]interface{}, error) {
	members, err := member.GetAllMembers()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(members))
	for i, m := range members {
		membershipStatus := "Inactive"
		paymentStatusLower := strings.ToLower(m.PaymentStatus)
		if paymentStatusLower == "paid" || m.SubscriptionActive {
			membershipStatus = "Active"
		}

		result[i] = map[string]interface{}{
			"id":                        m.ID,
			"tagaId":                    m.TagaID,
			"name":                      m.Name,
			"initial":                   m.Initial,
			"gender":                    m.Gender,
			"father_name":               m.FatherName,
			"mother_name":               m.MotherName,
			"educational_qualification": m.EducationalQualification,
			"designation":               m.Designation,
			"working_district":          m.WorkingDistrict,
			"native_district":           m.NativeDistrict,
			"recruitment_batch":         m.RecruitmentBatch,
			"seniority_number":          m.SeniorityNumber,
			"mobile_number":             m.MobileNumber,
			"emailId":                   m.EmailId,
			"tbf_number":                m.TbfNumber,
			"cps_gpf_number":            m.CpsGpfNumber,
			"date_of_birth":             m.DateOfBirth,
			"residential_address":       m.ResidentialAddress,
			"permanent_address":         m.PermanentAddress,
			"payment_status":            m.PaymentStatus,
			"membership_status":         membershipStatus,
		}
	}
	return result, nil
}

func getMapString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// createManualAnnualSubscription creates an active annual subscription record for a paid member
func createManualAnnualSubscription(memberID, memberEmail, memberName string) {
	startDate := time.Now()
	endDate := getMembershipYearEnd(startDate)
	nextDueDate := endDate.AddDate(0, 0, 1)

	// Expire any existing active annual subscription
	expireExistingAnnualSubscriptions(memberEmail)

	memberSub := model.MemberSubscription{
		ID:               uuid.New().String(),
		MemberID:         memberID,
		MemberEmail:      memberEmail,
		MemberName:       memberName,
		SubscriptionID:   "annual-subscription",
		SubscriptionName: "Annual Subscription",
		Amount:           0, // Admin override
		OrderID:          "admin_override_" + uuid.New().String()[:8],
		PaymentID:        "admin_override_" + uuid.New().String()[:8],
		Status:           "active",
		StartDate:        startDate,
		EndDate:          endDate,
		LastPaidDate:     startDate,
		NextDueDate:      nextDueDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	saveMemberSubscription(memberSub)
}

// GetMockEmails godoc
// @Summary Get mock emails for E2E testing
// @Tags Admin Members
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/admin/mock-emails [get]
func GetMockEmails(c *gin.Context) {
	// WARNING: In a real production environment, this endpoint should be disabled or strictly secured!
	mockEmailFile := "data/emails/mock_emails.json"
	data, err := os.ReadFile(mockEmailFile)
	if err != nil {
		c.JSON(http.StatusOK, make(map[string]string))
		return
	}
	
	var mockEmails map[string]string
	if err := json.Unmarshal(data, &mockEmails); err != nil {
		c.JSON(http.StatusOK, make(map[string]string))
		return
	}
	
	c.JSON(http.StatusOK, mockEmails)
}

// ClearMockEmails godoc
// @Summary Clear mock emails for E2E testing
// @Tags Admin Members
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/admin/mock-emails [delete]
func ClearMockEmails(c *gin.Context) {
	mockEmailFile := "data/emails/mock_emails.json"
	os.Remove(mockEmailFile)
	c.JSON(http.StatusOK, gin.H{"status": "cleared"})
}

// GetPendingEditRequests godoc
// @Summary Get all pending member edit requests
// @Tags Admin Edit Requests
// @Produce json
// @Security BearerAuth
// @Router /api/admin/edit-requests [get]
func GetPendingEditRequests(c *gin.Context) {
	file, err := os.ReadFile("data/requests/pending_requests.json")
	if err != nil {
		c.JSON(http.StatusOK, []model.FieldEditRequest{})
		return
	}
	var reqs []model.FieldEditRequest
	json.Unmarshal(file, &reqs)
	if reqs == nil {
		reqs = []model.FieldEditRequest{}
	}
	c.JSON(http.StatusOK, reqs)
}

type BulkProcessItem struct {
	ID           string `json:"id"`
	Status       string `json:"status"` // "approved" or "rejected"
	AdminRemarks string `json:"adminRemarks"`
}

type BulkProcessPayload struct {
	Items []BulkProcessItem `json:"items"`
}

// BulkProcessEditRequests godoc
// @Summary Bulk approve or reject field edit requests
// @Tags Admin Edit Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkProcessPayload true "Bulk Approval Details"
// @Router /api/admin/edit-requests/bulk-process [post]
func BulkProcessEditRequests(c *gin.Context) {
	var payload BulkProcessPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Read pending
	var pending []model.FieldEditRequest
	file, err := os.ReadFile("data/requests/pending_requests.json")
	if err == nil && len(file) > 0 {
		json.Unmarshal(file, &pending)
	}

	members, _ := readExistingMembers()
	membersUpdated := false

	var newPending []model.FieldEditRequest
	var processedThisBatch []model.FieldEditRequest

	// Process payload
	now := time.Now().Format(time.RFC3339)
	for i := range pending {
		var matchedItem *BulkProcessItem
		for j := range payload.Items {
			if payload.Items[j].ID == pending[i].ID {
				matchedItem = &payload.Items[j]
				break
			}
		}

		if matchedItem != nil {
			// Found it!
			pending[i].Status = matchedItem.Status
			pending[i].AdminRemarks = matchedItem.AdminRemarks
			pending[i].ProcessedAt = now
			
			processedThisBatch = append(processedThisBatch, pending[i])

			if matchedItem.Status == "approved" {
				// Update member
				for k, m := range members {
					if mId, ok := m["id"].(string); ok && mId == pending[i].MemberID {
						members[k][pending[i].Field] = pending[i].NewValue
						membersUpdated = true
						break
					}
				}
			}
		} else {
			newPending = append(newPending, pending[i])
		}
	}

	// Save members
	if membersUpdated {
		data, _ := json.MarshalIndent(members, "", "  ")
		os.WriteFile(config.Config.MembersFile, data, 0644)
	}

	// Update pending list
	pData, _ := json.MarshalIndent(newPending, "", "  ")
	os.WriteFile("data/requests/pending_requests.json", pData, 0644)

	// Add to processed
	var processed []model.FieldEditRequest
	pFile, err := os.ReadFile("data/requests/processed_requests.json")
	if err == nil && len(pFile) > 0 {
		json.Unmarshal(pFile, &processed)
	}
	processed = append(processed, processedThisBatch...)
	procData, _ := json.MarshalIndent(processed, "", "  ")
	os.WriteFile("data/requests/processed_requests.json", procData, 0644)

	// Send emails grouped by Member
	memberGroups := make(map[string][]model.FieldEditRequest)
	for _, p := range processedThisBatch {
		memberGroups[p.Email] = append(memberGroups[p.Email], p)
	}

	for email, groupFields := range memberGroups {
		if len(groupFields) > 0 {
			service.SendMemberRequestProcessedEmail(email, groupFields[0].MemberName, groupFields)
		}
	}

	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionUpdate, audit.ModuleMember,
		"edit_requests", "bulk_process",
		fmt.Sprintf("Admin processed %d field edit requests in bulk", len(processedThisBatch)),
		nil, processedThisBatch)

	respondMessage(c, "Bulk request processed successfully")
}
