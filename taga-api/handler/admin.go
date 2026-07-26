package handler

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service/auth"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
}

type RegistrationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// HandleMemberRegistration Swagger godoc
// @Summary      Handle new member registration
// @Description  Process new member registration file upload, validate data, store member details, and send email with temporary password
// @Tags         Admin
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Registration file (JSON/CSV format)"
// @Success      200  {object}  map[string]interface{}  "Registration successful"
// @Failure      400  {object}  map[string]interface{}  "Invalid file or validation errors"
// @Failure      422  {object}  map[string]interface{}  "Registration validation failed"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /api/admin/member-registration [post]
func HandleMemberRegistration(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	if !isValidFileFormat(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Only JSON or CSV allowed"})
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
			log.Printf("Failed to send error email to admin: %v", err)
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
			log.Printf("Failed to store member %s: %v", reg.EmailId, err)
			results = append(results, map[string]interface{}{
				"email":  reg.EmailId,
				"status": "failed",
				"error":  err.Error(),
			})
			continue
		}

		if err := sendSuccessEmail(reg.EmailId, tempPassword); err != nil {
			log.Printf("Failed to send email to %s: %v", reg.EmailId, err)
		}

		results = append(results, map[string]interface{}{
			"email":  reg.EmailId,
			"status": "success",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member registration completed",
		"total":   len(validRegistrations),
		"results": results,
	})
}

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

		sno := 0
		if records[i][0] != "" {
			sno, _ = strconv.Atoi(records[i][0])
		}

		registrations = append(registrations, RegistrationRequest{
			SNo:                      sno,
			TagaID:                   getField(records[i], 1),
			Name:                     getField(records[i], 2),
			Initial:                  getField(records[i], 3),
			Gender:                   getField(records[i], 4),
			FatherName:               getField(records[i], 5),
			MotherName:               getField(records[i], 6),
			EducationalQualification: getField(records[i], 7),
			Designation:              getField(records[i], 8),
			WorkingDistrict:          getField(records[i], 9),
			NativeDistrict:           getField(records[i], 10),
			RecruitmentBatch:         getField(records[i], 11),
			SeniorityNumber:          getField(records[i], 12),
			ResidentialAddress:       getField(records[i], 13),
			PermanentAddress:         getField(records[i], 14),
			DateOfBirth:              getField(records[i], 15),
			MobileNumber:             getField(records[i], 16),
			EmailId:                  getField(records[i], 17),
			TBFNumber:                getField(records[i], 18),
			CPSGPFNumber:             getField(records[i], 19),
		})
	}

	return registrations, nil
}

func validateRegistration(req *RegistrationRequest, existingMembers []map[string]interface{}) []RegistrationError {
	var errors []RegistrationError

	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, RegistrationError{
			Field:   "name",
			Message: "Name is required",
		})
	}

	if strings.TrimSpace(req.EmailId) == "" {
		errors = append(errors, RegistrationError{
			Field:   "email",
			Message: "Email is required",
		})
	} else if !isValidEmail(req.EmailId) {
		errors = append(errors, RegistrationError{
			Field:   "email",
			Message: "Invalid email format",
		})
	} else {
		for _, member := range existingMembers {
			if email, ok := member["emailId"].(string); ok && strings.EqualFold(email, req.EmailId) {
				errors = append(errors, RegistrationError{
					Field:   "email",
					Message: "Email is already registered",
				})
				break
			}
		}
	}

	if strings.TrimSpace(req.MobileNumber) == "" {
		errors = append(errors, RegistrationError{
			Field:   "phone",
			Message: "Phone is required",
		})
	} else if !isValidPhone(req.MobileNumber) {
		errors = append(errors, RegistrationError{
			Field:   "phone",
			Message: "Invalid phone format",
		})
	} else {
		for _, member := range existingMembers {
			if mobile, ok := member["mobile_number"].(string); ok && mobile == req.MobileNumber {
				errors = append(errors, RegistrationError{
					Field:   "phone",
					Message: "Mobile number is already registered",
				})
				break
			}
		}
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
	membersFilePath := filepath.Join(cfg.MembersFile)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
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
		"date_of_birth":             req.DateOfBirth,
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
		// Soft delete fields (default: not deleted)
		"deleted":    false,
		"deleted_at": nil,
	}

	updatedMembers := append(existingMembers, newMember)

	data, err := json.MarshalIndent(updatedMembers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}

	if err := os.WriteFile(membersFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write members file: %w", err)
	}

	return nil
}

// sendErrorEmailToAdmin — fixed: no zero-width characters in HTML tags
func sendErrorEmailToAdmin(errors []map[string]interface{}) error {
	cfg := config.GetConfig()
	adminEmail := cfg.FromEmail

	if adminEmail == "" {
		return fmt.Errorf("admin email not configured")
	}

	subject := "Member Registration Validation Errors"

	var body strings.Builder
	body.WriteString("<h3>Registration Validation Failed</h3>")
	body.WriteString("<p>The following registration requests have errors:</p>")
	body.WriteString("<ul>")

	for _, err := range errors {
		body.WriteString(fmt.Sprintf("<li><strong>Email:</strong> %v<br>", err["email"]))
		body.WriteString("<strong>Errors:</strong><ul>")

		if errList, ok := err["errors"].([]RegistrationError); ok {
			for _, e := range errList {
				body.WriteString(fmt.Sprintf("<li>%s: %s</li>", e.Field, e.Message))
			}
		}

		body.WriteString("</ul></li>")
	}

	body.WriteString("</ul>")
	body.WriteString("<p>Please correct these errors and resubmit.</p>")

	return sendEmail(adminEmail, subject, body.String())
}

// InitPassword Swagger godoc
// @Summary      Initialize member password(s)
// @Description  Initialize passwords for all members, a specific member, or none (default).
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        memberId  query  string  false  "Member ID or keyword: all | none (default)"
// @Success      200  {string}  string  "Password initialization result"
// @Failure      400  {string}  string  "Invalid or missing parameter"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /api/admin/init-password [post]
func InitPassword(c *gin.Context) {
	memberId := c.DefaultQuery("memberId", "none")

	if memberId == "none" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please specify memberId parameter. Use 'all' to reset all passwords or a specific member ID",
		})
		return
	}

	if memberId == "all" {
		go resetPasswordForAll()
		c.JSON(http.StatusOK, gin.H{
			"message": "Password reset process started for all members. This may take a moment.",
		})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Reset password for specific member is not yet implemented",
	})
}

func resetPasswordForAll() {
	const filePath = "data/member/members.json"

	b, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var members []model.Member
	if err := json.Unmarshal(b, &members); err != nil {
		panic(err)
	}

	for i := range members {
		members[i].Username = members[i].MobileNumber
		plain := genRandomPassword(8)
		hash, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
		members[i].Password = string(hash)
		members[i].FirstLogin = true
	}

	out, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filePath, out, 0644); err != nil {
		panic(err)
	}
}

func genRandomPassword(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func UploadMemberRegistration(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	tempPath := "tmp/" + file.Filename
	_ = os.MkdirAll("tmp", 0755)

	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(500, gin.H{"error": "failed to save file"})
		return
	}

	result := auth.ProcessRegistrationFile(tempPath)

	c.JSON(200, result)
}
