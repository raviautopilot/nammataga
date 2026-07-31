package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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

// isGracePeriod checks if current time is before Dec 31, 2026
func isGracePeriod() bool {
	graceEndDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.Local)
	return time.Now().Before(graceEndDate)
}

// GetMembersList godoc
// @Summary Get list of members
// @Description Get paginated list of members with search and filters
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
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

		paymentStatus := getPaymentStatusFromSubscription(email, subscriptionMap)

		if paymentFilter != "" && paymentFilter != "all" {
			if !strings.EqualFold(paymentStatus, paymentFilter) {
				continue
			}
		}

		if search != "" {
			name := strings.ToLower(getString(m, "name"))
			emailLower := strings.ToLower(email)
			mobile := getString(m, "mobile_number")
			if !strings.Contains(name, search) &&
				!strings.Contains(emailLower, search) &&
				!strings.Contains(mobile, search) {
				continue
			}
		}

		member := MemberListItem{
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
			MembershipStatus: getMembershipStatus(m),
		}
		memberList = append(memberList, member)
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

	c.JSON(http.StatusOK, MemberListResponse{
		Members:    paginatedMembers,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetMemberDistricts godoc
// @Summary Get list of districts with member counts
// @Description Get all districts that have members
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/admin/members/districts [get]
func GetMemberDistricts(c *gin.Context) {
	members, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
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

	c.JSON(http.StatusOK, districts)
}

// GetMemberStats godoc
// @Summary Get member statistics
// @Description Get total members, active, unpaid counts
// @Tags Admin Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/members/stats [get]
func GetMemberStats(c *gin.Context) {
	members, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	subscriptionMap := loadSubscriptionPaymentMap()

	total := len(members)
	active := 0
	unpaid := 0

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	newThisMonth := 0

	for _, m := range members {
		email := getString(m, "emailId")
		paymentStatus := getPaymentStatusFromSubscription(email, subscriptionMap)

		if getMembershipStatus(m) == "Active" {
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

	// 🔥 During grace period, unpaid is always 0
	if isGracePeriod() {
		unpaid = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"totalMembers":  total,
		"activeMembers": active,
		"unpaid":        unpaid,
		"newThisMonth":  newThisMonth,
	})
}

// loadSubscriptionPaymentMap loads subscription data and creates a map of email -> isPaid
func loadSubscriptionPaymentMap() map[string]bool {
	subscriptionMap := make(map[string]bool)

	// 🔥 GRACE PERIOD: Skip file loading, all lookups return false (but overridden by getPaymentStatus)
	if isGracePeriod() {
		return subscriptionMap
	}

	filePath := filepath.Join("data", "subscriptions", "member_subscriptions.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return subscriptionMap
	}

	var subscriptions []map[string]interface{}
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		return subscriptionMap
	}

	for _, sub := range subscriptions {
		subID, _ := sub["subscription_id"].(string)
		if subID != "annual-subscription" {
			continue
		}
		if email, ok := sub["member_email"].(string); ok {
			if status, ok := sub["status"].(string); ok && status == "active" {
				subscriptionMap[email] = true
			}
		}
	}

	return subscriptionMap
}

// getPaymentStatusFromSubscription returns payment status based on subscription data
func getPaymentStatusFromSubscription(email string, subscriptionMap map[string]bool) string {
	// 🔥 GRACE PERIOD: Everyone is "Paid" until December 31, 2026
	if isGracePeriod() {
		return "Paid"
	}

	if isPaid, exists := subscriptionMap[email]; exists && isPaid {
		return "Paid"
	}
	return "Unpaid"
}



// getMembershipStatus returns membership status
func getMembershipStatus(member map[string]interface{}) string {
	// 🔥 GRACE PERIOD: Everyone is "Active" until December 31, 2026
	if isGracePeriod() {
		return "Active"
	}

	email := getString(member, "emailId")
	subscriptionMap := loadSubscriptionPaymentMap()
	if isPaid, exists := subscriptionMap[email]; exists && isPaid {
		return "Active"
	}
	return "Inactive"
}
