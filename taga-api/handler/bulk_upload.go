package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"taga-api/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

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

// BulkUploadMembers godoc
// @Summary Bulk upload members
// @Description Upload multiple members via CSV or Excel file (supports .csv, .xlsx, .xls)
// @Tags Admin Members
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV or Excel file (.csv, .xlsx, .xls)"
// @Success 200 {object} BulkUploadResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/members/bulk-upload [post]
func BulkUploadMembers(c *gin.Context) {
	startTime := time.Now()

	// 1. Get and validate file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum 10MB allowed"})
		return
	}

	// Validate extension - Support CSV and Excel
	ext := strings.ToLower(filepath.Ext(file.Filename))
	supportedExts := map[string]bool{
		".csv":  true,
		".xlsx": true,
		".xls":  true,
	}
	if !supportedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV and Excel files (.csv, .xlsx, .xls) are allowed"})
		return
	}

	// 2. Save temp file
	tempDir, err := os.MkdirTemp("", "bulk_upload_*")
	if err != nil {
		config.Logger.Error("Failed to create temp directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process upload"})
		return
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		config.Logger.Error("Failed to save uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 3. Parse file based on extension
	var registrations []RegistrationRequest
	if ext == ".csv" {
		registrations, err = parseCSVFileForBulk(tempPath)
	} else {
		registrations, err = parseExcelFileForBulk(tempPath)
	}

	if err != nil {
		config.Logger.Error("Failed to parse file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse file: " + err.Error()})
		return
	}

	if len(registrations) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid records found in file"})
		return
	}

	// Check limit - Maximum 500 members
	if len(registrations) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 500 members per upload allowed"})
		return
	}

	// 4. Read existing members
	existingMembers, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read existing members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read member database"})
		return
	}

	// 5. Create a map for faster duplicate checking
	existingEmailMap := make(map[string]bool)
	existingMobileMap := make(map[string]bool)
	for _, m := range existingMembers {
		if email, ok := m["emailId"].(string); ok {
			existingEmailMap[strings.ToLower(email)] = true
		}
		if mobile, ok := m["mobile_number"].(string); ok && mobile != "" {
			existingMobileMap[mobile] = true
		}
	}

	// Also check existing TAGA IDs for uniqueness
	existingTagaIDMap := make(map[string]bool)
	for _, m := range existingMembers {
		if tagaID, ok := m["tagaId"].(string); ok && tagaID != "" {
			existingTagaIDMap[tagaID] = true
		}
	}

	// 6. Process registrations
	result := BulkUploadResult{
		ProcessedAt: time.Now().Format(time.RFC3339),
		TotalCount:  len(registrations),
		Failed:      []BulkUploadFailure{},
	}

	var mu sync.Mutex

	for idx, reg := range registrations {
		// Check for duplicate TAGA ID
		var errors []string
		if existingTagaIDMap[reg.TagaID] {
			errors = append(errors, "TAGA ID already exists")
		}

		// Validate other fields
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

		// Add to TAGA ID map for duplicate checking within same file
		existingTagaIDMap[reg.TagaID] = true

		// Process valid registration
		tempPassword := generateTempPassword()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

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
			"date_of_birth":             reg.DateOfBirth,
			"mobile_number":             reg.MobileNumber,
			"emailId":                   reg.EmailId,
			"tbf_number":                reg.TBFNumber,
			"cps_gpf_number":            reg.CPSGPFNumber,
			"password":                  string(hashedPassword),
			"first_login":               true,
			"created_at":                time.Now().Format(time.RFC3339),
			// Payment fields
			"payment_status":          "Unpaid",
			"subscription_active":     false,
			"subscription_end_date":   nil,
			"subscription_updated_at": nil,
		}
		existingMembers = append(existingMembers, newMember)

		// Update maps for duplicate checking within same file
		existingEmailMap[strings.ToLower(reg.EmailId)] = true
		existingMobileMap[reg.MobileNumber] = true

		mu.Lock()
		result.SuccessCount++
		mu.Unlock()

		// Send email asynchronously
		go sendSuccessEmail(reg.EmailId, tempPassword)
	}

	// 7. Save all members to file
	cfg := config.GetConfig()
	membersFilePath := cfg.MembersFile

	data, err := json.MarshalIndent(existingMembers, "", "  ")
	if err != nil {
		config.Logger.Error("Failed to marshal members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	if err := os.WriteFile(membersFilePath, data, 0644); err != nil {
		config.Logger.Error("Failed to write members file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	config.Logger.Info("Bulk upload completed",
		zap.Int("success", result.SuccessCount),
		zap.Int("failed", result.FailedCount),
		zap.Int("total", result.TotalCount),
		zap.Duration("duration", time.Since(startTime)),
	)

	// 🚨 FIRST validate
	if result.SuccessCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input file",
			"details": result,
		})
		return
	}

	// ✅ ONLY NOW save data
	// cfg := config.GetConfig()
	// data, err := json.MarshalIndent(existingMembers, "", "  ")
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save"})
	// 	return
	// }

	if err := os.WriteFile(membersFilePath, data, 0644); err != nil {
		config.Logger.Error("Failed to write members file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save members"})
		return
	}

	// ✅ Final success response
	c.JSON(http.StatusOK, result)
}

// parseCSVFileForBulk parses CSV file for bulk upload (TAGA format - NO SNo column)
func parseCSVFileForBulk(filePath string) ([]RegistrationRequest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have header row and at least one data row")
	}

	var registrations []RegistrationRequest

	// Skip header row (i=0 is header, i=1 is first data row)
	for i := 1; i < len(records); i++ {
		row := records[i]
		if isEmptyRow(row) {
			continue
		}

		// Check if row has enough columns (expected 19 columns)
		if len(row) < 19 {
			return nil, fmt.Errorf("row %d has insufficient columns (expected 19, got %d)", i+1, len(row))
		}

		reg := RegistrationRequest{
			SNo:                      i,
			TagaID:                   getFieldSafe(row, 0),  // Column 0: Taga ID
			Name:                     getFieldSafe(row, 1),  // Column 1: Name
			Initial:                  getFieldSafe(row, 2),  // Column 2: Initial
			Gender:                   getFieldSafe(row, 3),  // Column 3: Gender
			FatherName:               getFieldSafe(row, 4),  // Column 4: Father Name
			MotherName:               getFieldSafe(row, 5),  // Column 5: Mother Name
			EducationalQualification: getFieldSafe(row, 6),  // Column 6: Educational Qualification
			Designation:              getFieldSafe(row, 7),  // Column 7: Designation
			WorkingDistrict:          getFieldSafe(row, 8),  // Column 8: Working District
			NativeDistrict:           getFieldSafe(row, 9),  // Column 9: Native District
			RecruitmentBatch:         getFieldSafe(row, 10), // Column 10: Recruitment Batch
			SeniorityNumber:          getFieldSafe(row, 11), // Column 11: Seniority Number
			ResidentialAddress:       getFieldSafe(row, 12), // Column 12: Residential Address
			PermanentAddress:         getFieldSafe(row, 13), // Column 13: Permanent Address
			DateOfBirth:              getFieldSafe(row, 14), // Column 14: Date of Birth
			MobileNumber:             getFieldSafe(row, 15), // Column 15: Mobile Number
			EmailId:                  getFieldSafe(row, 16), // Column 16: Email ID
			TBFNumber:                getFieldSafe(row, 17), // Column 17: TBF Number
			CPSGPFNumber:             getFieldSafe(row, 18), // Column 18: CPS/GPF Number
		}
		registrations = append(registrations, reg)
	}

	return registrations, nil
}

// parseExcelFileForBulk parses Excel file for bulk upload (TAGA format)
func parseExcelFileForBulk(filePath string) ([]RegistrationRequest, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no sheet found in Excel file")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must have header row and at least one data row")
	}

	var registrations []RegistrationRequest

	// Skip header row
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isEmptyRow(row) {
			continue
		}

		// Check if row has enough columns
		if len(row) < 19 {
			return nil, fmt.Errorf("row %d has insufficient columns (expected 19, got %d)", i+1, len(row))
		}

		reg := RegistrationRequest{
			SNo:                      i,
			TagaID:                   getFieldSafe(row, 0),  // Column 0: Taga ID
			Name:                     getFieldSafe(row, 1),  // Column 1: Name
			Initial:                  getFieldSafe(row, 2),  // Column 2: Initial
			Gender:                   getFieldSafe(row, 3),  // Column 3: Gender
			FatherName:               getFieldSafe(row, 4),  // Column 4: Father Name
			MotherName:               getFieldSafe(row, 5),  // Column 5: Mother Name
			EducationalQualification: getFieldSafe(row, 6),  // Column 6: Educational Qualification
			Designation:              getFieldSafe(row, 7),  // Column 7: Designation
			WorkingDistrict:          getFieldSafe(row, 8),  // Column 8: Working District
			NativeDistrict:           getFieldSafe(row, 9),  // Column 9: Native District
			RecruitmentBatch:         getFieldSafe(row, 10), // Column 10: Recruitment Batch
			SeniorityNumber:          getFieldSafe(row, 11), // Column 11: Seniority Number
			ResidentialAddress:       getFieldSafe(row, 12), // Column 12: Residential Address
			PermanentAddress:         getFieldSafe(row, 13), // Column 13: Permanent Address
			DateOfBirth:              getFieldSafe(row, 14), // Column 14: Date of Birth
			MobileNumber:             getFieldSafe(row, 15), // Column 15: Mobile Number
			EmailId:                  getFieldSafe(row, 16), // Column 16: Email ID
			TBFNumber:                getFieldSafe(row, 17), // Column 17: TBF Number
			CPSGPFNumber:             getFieldSafe(row, 18), // Column 18: CPS/GPF Number
		}
		registrations = append(registrations, reg)
	}

	return registrations, nil
}



// validateRegistrationForBulk validates registration with map lookups for speed
func validateRegistrationForBulk(reg *RegistrationRequest, existingEmails, existingMobiles map[string]bool) []string {
	var errors []string

	// Validate TAGA ID
	if strings.TrimSpace(reg.TagaID) == "" {
		errors = append(errors, "TAGA ID is required")
	}

	// Validate Name
	if strings.TrimSpace(reg.Name) == "" {
		errors = append(errors, "Name is required")
	}

	// Validate Email
	if strings.TrimSpace(reg.EmailId) == "" {
		errors = append(errors, "Email is required")
	} else if !isValidEmail(reg.EmailId) {
		errors = append(errors, "Invalid email format")
	} else if existingEmails[strings.ToLower(reg.EmailId)] {
		errors = append(errors, "Email already exists")
	}

	// Validate Mobile and check duplicate
	if strings.TrimSpace(reg.MobileNumber) == "" {
		errors = append(errors, "Mobile number is required")
	} else if !isValidPhone(reg.MobileNumber) {
		errors = append(errors, "Invalid mobile number format")
	} else if existingMobiles[reg.MobileNumber] {
		errors = append(errors, "Mobile number already exists")
	}

	// Validate Date of Birth
	if strings.TrimSpace(reg.DateOfBirth) == "" {
		errors = append(errors, "Date of birth is required")
	}

	// Validate Designation
	if strings.TrimSpace(reg.Designation) == "" {
		errors = append(errors, "Designation is required")
	}

	// Validate Educational Qualification
	if strings.TrimSpace(reg.EducationalQualification) == "" {
		errors = append(errors, "Educational qualification is required")
	}

	return errors
}

// Helper functions
func getFieldSafe(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

func isEmptyRow(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
