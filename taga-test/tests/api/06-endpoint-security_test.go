package api_tests

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"e2e-template/tests"
)

type EndpointDefinition struct {
	Method            string `json:"method"`
	Path              string `json:"path"`
	Category          string `json:"category"`
	ExpectedProtected bool   `json:"expected_protected"`
}

// SwaggerEndpoints lists all 70 endpoints from the API Swagger specification.
var SwaggerEndpoints = []EndpointDefinition{
	// Health & Root (Public)
	{Method: "GET", Path: "/", Category: "Health & Root", ExpectedProtected: false},
	{Method: "GET", Path: "/health", Category: "Health & Root", ExpectedProtected: false},

	// Public Info
	{Method: "GET", Path: "/api/public/about", Category: "Public Info", ExpectedProtected: false},
	{Method: "GET", Path: "/api/public/about/stats", Category: "Public Info", ExpectedProtected: false},
	{Method: "GET", Path: "/api/public/about/objectives", Category: "Public Info", ExpectedProtected: false},
	{Method: "GET", Path: "/api/public/about/services", Category: "Public Info", ExpectedProtected: false},
	{Method: "GET", Path: "/api/public/about/contact", Category: "Public Info", ExpectedProtected: false},
	{Method: "GET", Path: "/api/logo", Category: "Public Info", ExpectedProtected: false},
	{Method: "POST", Path: "/api/webhook/razorpay", Category: "Webhook", ExpectedProtected: false},

	// Events & Gallery (Public)
	{Method: "GET", Path: "/api/events/upcoming", Category: "Events", ExpectedProtected: false},
	{Method: "GET", Path: "/api/gallery/years", Category: "Gallery", ExpectedProtected: false},
	{Method: "GET", Path: "/api/gallery", Category: "Gallery", ExpectedProtected: false},

	// Office Bearers & Office Info (Public)
	{Method: "GET", Path: "/api/office-bearers/state-executive", Category: "Office Bearers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/office-bearers/districts", Category: "Office Bearers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/office-bearers/district-office-bearers", Category: "Office Bearers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/office", Category: "Office", ExpectedProtected: false},

	// Resources
	{Method: "GET", Path: "/api/resources", Category: "Resources", ExpectedProtected: true},
	{Method: "GET", Path: "/api/resources/external-links", Category: "Resources", ExpectedProtected: true},
	{Method: "GET", Path: "/api/resources/1", Category: "Resources", ExpectedProtected: true},

	// Grievances (Public Submission & Reference Data)
	{Method: "GET", Path: "/api/grievances", Category: "Grievances", ExpectedProtected: false},
	{Method: "GET", Path: "/api/categories", Category: "Grievances", ExpectedProtected: false},
	{Method: "GET", Path: "/api/priorities", Category: "Grievances", ExpectedProtected: false},

	// TAGA Towers (Public Catalog)
	{Method: "GET", Path: "/api/towers/rooms", Category: "TAGA Towers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/towers/availability", Category: "TAGA Towers", ExpectedProtected: false},

	// Member Auth Entry Points (Unauthenticated Entry Points)
	{Method: "POST", Path: "/api/admin/login", Category: "Admin Login", ExpectedProtected: false},
	{Method: "POST", Path: "/api/member/login", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/auth/forgot-password", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/auth/member-forgot-password", Category: "Member Auth", ExpectedProtected: false},

	{Method: "POST", Path: "/api/auth/reset-password", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/member/logout", Category: "Member Auth", ExpectedProtected: false},

	// Member Protected Routes (Auth Required)
	{Method: "GET", Path: "/api/member/profile", Category: "Member Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/member/profile", Category: "Member Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/member/notifications", Category: "Member Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/member/notifications/1/read", Category: "Member Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/member/notifications/unread/count", Category: "Member Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/member/change-password", Category: "Member Protected", ExpectedProtected: false},

	// Subscription Protected Routes (Auth Required)
	{Method: "POST", Path: "/api/subscriptions/create-order", Category: "Subscription Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/subscriptions/verify-payment", Category: "Subscription Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/subscriptions/status", Category: "Subscription Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/subscriptions/member-paid", Category: "Subscription Protected", ExpectedProtected: true},

	// Admin Protected Routes (AdminAuthMiddleware Required)
	{Method: "POST", Path: "/api/admin/announcements/send", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/events/create", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/admin/events/1", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "DELETE", Path: "/api/admin/events/1", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/gallery/upload", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "DELETE", Path: "/api/admin/gallery/1", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/init-password", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/member-registration", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/members", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/members/add", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/members/bulk-upload", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/members/districts", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/members/export", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/members/stats", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/admin/members/1", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "DELETE", Path: "/api/admin/members/1", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/office-bearers/backup/restore", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/office-bearers/backups", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/office-bearers/district/test", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/admin/office-bearers/district/test", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/office-bearers/districts", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/reports/members", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/resources/upload", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "DELETE", Path: "/api/admin/resources/cat1/doc1", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/send-renewal-reminders", Category: "Admin Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/towers/admin/bookings", Category: "TAGA Towers (Admin)", ExpectedProtected: true},
	{Method: "POST", Path: "/admin/upload-registration", Category: "Legacy Admin Protected", ExpectedProtected: true},

	// Non-Existent Routes
	{Method: "GET", Path: "/api/nonexistent/admin", Category: "Invalid Route", ExpectedProtected: false},
	{Method: "POST", Path: "/api/admin/hidden", Category: "Invalid Route", ExpectedProtected: false},

	// Business Logic Negative
	{Method: "GET", Path: "/api/member/profile?user_id=-1", Category: "Business Logic IDOR", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/members?page=-5", Category: "Business Logic Negative Pagination", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/members?page=9223372036854775807", Category: "Business Logic Extreme Bounds", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/admin/events/1?action=cancel&status=completed", Category: "Business Logic State Machine Violation", ExpectedProtected: true},
	{Method: "POST", Path: "/api/admin/members/add?role=superadmin&override=true", Category: "Business Logic Role Context Switch", ExpectedProtected: true},
	{Method: "GET", Path: "/api/admin/events/1?start_date=2120-01-01&end_date=1990-01-01", Category: "Business Logic Paradox", ExpectedProtected: true},
}

// TestAPI_06_EndpointSecurity audits all swagger endpoints individually, populating report entries for every endpoint.
func TestAPI_06_EndpointSecurity(t *testing.T) {
	baseURL := tests.GlobalConfig.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, ep := range SwaggerEndpoints {
		testName := fmt.Sprintf("[%s] %s", ep.Method, ep.Path)
		description := fmt.Sprintf("Category: %s | Security audit for %s %s", ep.Category, ep.Method, ep.Path)

		expectedStr := "UNSECURED (Public Access Allowed: HTTP 200/400/404)"
		if ep.ExpectedProtected {
			expectedStr = "SECURED (Authentication Required: HTTP 401/403)"
		}

		endpointCopy := ep

		tests.RunAPITestWithDetails(
			t,
			testName,
			description,
			expectedStr,
			func(tc *tests.TestContext) {
				fullURL := strings.TrimRight(baseURL, "/") + endpointCopy.Path
				req, err := http.NewRequest(endpointCopy.Method, fullURL, nil)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("Failed to create HTTP request: %v", err)
					tc.Fatalf("Failed to create HTTP request: %v", err)
				}

				req.Header.Set("User-Agent", "APISecurityAudit/1.0")
				if endpointCopy.Method == "POST" || endpointCopy.Method == "PUT" {
					req.Header.Set("Content-Type", "application/json")
				}

				resp, err := httpClient.Do(req)
				statusCode := 0
				if err == nil {
					statusCode = resp.StatusCode
					_, _ = io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				} else {
					tc.FailureReason = fmt.Sprintf("Network execution error: %v", err)
					tc.Fatalf("Network execution error: %v", err)
				}

				isSecured := (statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden)

				if endpointCopy.ExpectedProtected {
					if isSecured {
						tc.Actual = fmt.Sprintf("HTTP %d (SECURED)", statusCode)
					} else {
						tc.Actual = fmt.Sprintf("HTTP %d (UNAUTHENTICATED ACCESS ALLOWED - SECURITY RISK)", statusCode)
						tc.FailureReason = fmt.Sprintf("Endpoint %s %s should be SECURED (HTTP 401/403) but returned HTTP %d", endpointCopy.Method, endpointCopy.Path, statusCode)
						tc.Errorf("Security Risk: Protected endpoint %s %s allowed unauthenticated access (HTTP %d)", endpointCopy.Method, endpointCopy.Path, statusCode)
					}
				} else {
					if isSecured {
						tc.Actual = fmt.Sprintf("HTTP %d (UNEXPECTEDLY PROTECTED)", statusCode)
					} else {
						tc.Actual = fmt.Sprintf("HTTP %d (UNSECURED / PUBLIC)", statusCode)
					}
				}
			},
		)
	}
}
