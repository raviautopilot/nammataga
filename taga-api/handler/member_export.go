package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"taga-api/service/member"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ExportMembersToExcel godoc
// @Summary Export members to Excel
// @Description Export all members with filters to Excel file with summary sheet
// @Tags Admin Reports
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param district query string false "Filter by district"
// @Param payment_status query string false "Filter by payment status (paid/unpaid)"
// @Success 200 {file} file
// @Router /api/admin/members/export [get]
func ExportMembersToExcel(c *gin.Context) {
	// Get filter parameters
	district := c.Query("district")
	paymentStatus := c.Query("payment_status")

	// Debug: Log received filters
	fmt.Printf("=== Export Members Called ===\n")
	fmt.Printf("District filter: '%s'\n", district)
	fmt.Printf("Payment Status filter: '%s'\n", paymentStatus)

	// Get all members
	allMembers, err := getAllMembersFromStorage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	fmt.Printf("Total members in storage: %d\n", len(allMembers))

	// Apply filters
	filteredMembers := make([]map[string]interface{}, 0)
	for idx, memberMap := range allMembers {
		// Debug first few members
		if idx < 5 {
			fmt.Printf("Member %d - Name: '%s', Working District: '%s', Payment Status: '%s'\n",
				idx,
				getMapString(memberMap, "name"),
				getMapString(memberMap, "working_district"),
				getMapString(memberMap, "payment_status"))
		}

		// Apply district filter
		if district != "" && district != "all" {
			workingDistrict := getMapString(memberMap, "working_district")
			if !strings.EqualFold(workingDistrict, district) {
				if idx < 3 {
					fmt.Printf("  Skipped due to district: '%s' != '%s'\n", workingDistrict, district)
				}
				continue
			}
		}

		// Apply payment status filter (case-insensitive)
		if paymentStatus != "" && paymentStatus != "all" {
			memberPaymentStatus := getMapString(memberMap, "payment_status")
			// Debug comparison
			if idx < 5 {
				fmt.Printf("  Comparing payment: member='%s' vs filter='%s', EqualFold=%v\n",
					memberPaymentStatus, paymentStatus, strings.EqualFold(memberPaymentStatus, paymentStatus))
			}
			if !strings.EqualFold(memberPaymentStatus, paymentStatus) {
				continue
			}
		}

		filteredMembers = append(filteredMembers, memberMap)
	}

	fmt.Printf("Filtered members count: %d\n", len(filteredMembers))

	// If no members found, still create Excel with headers and summary
	if len(filteredMembers) == 0 {
		fmt.Printf("WARNING: No members found with current filters!\n")
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing Excel file:", err)
		}
	}()

	// ==================== SHEET 1: Member Details ====================
	sheetName := "Member Details"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sheet"})
		return
	}
	f.SetActiveSheet(index)

	// Define headers (11 columns as requested)
	headers := []string{
		"TAGA ID", "Name", "Initial", "Gender", "District",
		"Designation", "Batch", "Mobile", "Email", "Payment Status", "Membership Status",
	}

	// Write headers with styling
	headerStyle, err := f.NewStyle(&excelize.Style{
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
	if err != nil {
		fmt.Println("Error creating header style:", err)
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		if headerStyle != 0 {
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}

	// Write data rows
	paidCount := 0
	unpaidCount := 0
	activeCount := 0

	for rowIdx, memberMap := range filteredMembers {
		rowNum := rowIdx + 2

		// Get values
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

		// Count statistics (case-insensitive)
		paymentStatusLower := strings.ToLower(paymentStatusVal)
		if paymentStatusLower == "paid" {
			paidCount++
		} else if paymentStatusLower == "unpaid" {
			unpaidCount++
		}

		membershipStatusLower := strings.ToLower(membershipStatus)
		if membershipStatusLower == "active" {
			activeCount++
		}

		// Write row
		rowData := []interface{}{tagaID, name, initial, gender, workingDistrict,
			designation, recruitmentBatch, mobileNumber, email, paymentStatusVal, membershipStatus}

		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%s%d", string(rune('A'+colIdx)), rowNum)
			f.SetCellValue(sheetName, cell, value)

			// Add border style to data cells
			dataStyle, _ := f.NewStyle(&excelize.Style{
				Border: []excelize.Border{
					{Type: "left", Color: "#CCCCCC", Style: 1},
					{Type: "right", Color: "#CCCCCC", Style: 1},
					{Type: "top", Color: "#CCCCCC", Style: 1},
					{Type: "bottom", Color: "#CCCCCC", Style: 1},
				},
			})
			if dataStyle != 0 {
				f.SetCellStyle(sheetName, cell, cell, dataStyle)
			}
		}
	}

	// If no members found, add a message row
	if len(filteredMembers) == 0 {
		messageCell := "A2"
		f.SetCellValue(sheetName, messageCell, "No members found with the selected filters")
		f.MergeCell(sheetName, "A2", fmt.Sprintf("%s2", string(rune('A'+len(headers)-1))))
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		colLetter := string(rune('A' + i))
		f.SetColWidth(sheetName, colLetter, colLetter, 18)
	}

	// Add filter to header row
	if len(filteredMembers) > 0 {
		endCol := string(rune('A' + len(headers) - 1))
		endRow := len(filteredMembers) + 1
		f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", endCol, endRow), []excelize.AutoFilterOptions{})
	}

	// ==================== SHEET 2: Summary ====================
	summarySheet := "Summary"
	_, err = f.NewSheet(summarySheet)
	if err != nil {
		fmt.Println("Error creating summary sheet:", err)
	}

	// Summary title
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E8F5E9"}, Pattern: 1},
	})
	f.MergeCell(summarySheet, "A1", "B1")
	f.SetCellValue(summarySheet, "A1", "MEMBER EXPORT SUMMARY")
	f.SetCellStyle(summarySheet, "A1", "A1", titleStyle)

	// Summary data
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

	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11},
		Border: []excelize.Border{
			{Type: "left", Color: "#CCCCCC", Style: 1},
			{Type: "right", Color: "#CCCCCC", Style: 1},
			{Type: "top", Color: "#CCCCCC", Style: 1},
			{Type: "bottom", Color: "#CCCCCC", Style: 1},
		},
	})

	for rowIdx, row := range summaryData {
		rowNum := rowIdx + 3
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", rowNum), row[0])
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", rowNum), row[1])
		f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), summaryStyle)
	}

	// Bold for labels
	labelStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
	})
	for rowIdx := 3; rowIdx < len(summaryData)+3; rowIdx++ {
		if rowIdx == 5 || rowIdx == 7 || rowIdx == 10 {
			continue // Skip empty rows
		}
		f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), labelStyle)
	}

	// Set column widths for summary sheet
	f.SetColWidth(summarySheet, "A", "A", 25)
	f.SetColWidth(summarySheet, "B", "B", 30)

	// Remove default Sheet1
	f.DeleteSheet("Sheet1")

	// Set filename
	filename := fmt.Sprintf("TAGA_Members_Export_%s.xlsx", time.Now().Format("2006-01-02_15-04-05"))

	// Set response headers
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Write to response
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}
}

// Helper function to get filter display value
func getFilterDisplay(filter string) string {
	if filter == "" || filter == "all" {
		return "None (All members)"
	}
	return filter
}

// getAllMembersFromStorage reads all members from the JSON file
func getAllMembersFromStorage() ([]map[string]interface{}, error) {
	members, err := member.GetAllMembers()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Raw members loaded from file: %d\n", len(members))

	// Convert to slice of maps for easier handling
	result := make([]map[string]interface{}, len(members))
	for i, m := range members {
		// Debug raw payment status for first few members
		if i < 5 {
			fmt.Printf("Raw member %d - Name: '%s', PaymentStatus: '%s', SubscriptionActive: %v\n",
				i, m.Name, m.PaymentStatus, m.SubscriptionActive)
		}

		// Determine membership status based on payment
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

// getMapString - Helper function to safely get string from map (renamed to avoid conflict)
func getMapString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
