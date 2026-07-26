package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateMemberReport godoc
// @Summary Generate member report
// @Description Export member data as CSV report
// @Tags Admin Reports
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param report_type query string false "Report type (membership/financial/activities/district)"
// @Param period query string false "Period (current_month/last_month/current_quarter/current_year/all_time)"
// @Success 200 {file} file
// @Router /api/admin/reports/members [get]
func GenerateMemberReport(c *gin.Context) {
	reportType := c.DefaultQuery("report_type", "membership")
	period := c.DefaultQuery("period", "all_time")

	// Use the variables to avoid unused warning
	_ = reportType // Mark as used for now (can be used for filtering later)
	_ = period     // Mark as used for now (can be used for date filtering later)

	// Read members
	members, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	// Set filename
	filename := fmt.Sprintf("TAGA_%s_report_%s.csv", reportType, time.Now().Format("2006-01-02"))

	// Set response headers
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write headers
	headers := []string{
		"Member ID", "Name", "Initial", "Gender", "Father Name", "Mother Name",
		"Educational Qualification", "Designation", "Working District", "Native District",
		"Recruitment Batch", "Seniority Number", "Mobile Number", "Email ID",
		"TBF Number", "CPS/GPF Number", "Registration Date",
	}
	writer.Write(headers)

	// Write member data
	for _, member := range members {
		row := []string{
			getString(member, "id"),
			getString(member, "name"),
			getString(member, "initial"),
			getString(member, "gender"),
			getString(member, "father_name"),
			getString(member, "mother_name"),
			getString(member, "educational_qualification"),
			getString(member, "designation"),
			getString(member, "working_district"),
			getString(member, "native_district"),
			getString(member, "recruitment_batch"),
			getString(member, "seniority_number"),
			getString(member, "mobile_number"),
			getString(member, "emailId"),
			getString(member, "tbf_number"),
			getString(member, "cps_gpf_number"),
			getString(member, "created_at"),
		}
		writer.Write(row)
	}

	// Write summary
	writer.Write([]string{})
	writer.Write([]string{"SUMMARY"})
	writer.Write([]string{fmt.Sprintf("Total Members,%d", len(members))})

	// Count by district
	districtCount := make(map[string]int)
	for _, member := range members {
		district := getString(member, "working_district")
		if district != "" {
			districtCount[district]++
		}
	}

	writer.Write([]string{})
	writer.Write([]string{"DISTRICT WISE BREAKDOWN"})
	for district, count := range districtCount {
		writer.Write([]string{district, strconv.Itoa(count)})
	}
}
