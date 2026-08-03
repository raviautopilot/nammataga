package api_tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// SwaggerEndpoints lists all endpoints from the API Swagger specification.
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
	{Method: "GET", Path: "/api/member/banner", Category: "Public Info", ExpectedProtected: false},
	{Method: "POST", Path: "/api/webhook/razorpay", Category: "Webhook", ExpectedProtected: false},

	// Resources (Public)
	{Method: "GET", Path: "/api/resources", Category: "Resources", ExpectedProtected: false},
	{Method: "GET", Path: "/api/resources/external-links", Category: "Resources", ExpectedProtected: false},
	{Method: "GET", Path: "/api/resources/1", Category: "Resources", ExpectedProtected: false},

	// Events & Gallery (Public)
	{Method: "GET", Path: "/api/events/upcoming", Category: "Events", ExpectedProtected: false},
	{Method: "GET", Path: "/api/gallery/years", Category: "Gallery", ExpectedProtected: false},
	{Method: "GET", Path: "/api/gallery", Category: "Gallery", ExpectedProtected: false},

	// Office Bearers & Office Info (Public)
	{Method: "GET", Path: "/api/office-bearers/state-executive", Category: "Office Bearers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/office-bearers/districts", Category: "Office Bearers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/office-bearers/district-office-bearers", Category: "Office Bearers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/office", Category: "Office", ExpectedProtected: false},

	// Grievances (Public)
	{Method: "GET", Path: "/api/grievances", Category: "Grievances", ExpectedProtected: false},
	{Method: "GET", Path: "/api/categories", Category: "Grievances", ExpectedProtected: false},
	{Method: "GET", Path: "/api/priorities", Category: "Grievances", ExpectedProtected: false},

	// TAGA Towers Public
	{Method: "GET", Path: "/api/towers/rooms", Category: "TAGA Towers", ExpectedProtected: false},
	{Method: "GET", Path: "/api/towers/availability", Category: "TAGA Towers", ExpectedProtected: false},

	// Member Auth Entry Points (Public)
	{Method: "POST", Path: "/api/auth/forgot-password", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/auth/reset-password", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/auth/member-forgot-password", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/member/login", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/member/logout", Category: "Member Auth", ExpectedProtected: false},
	{Method: "POST", Path: "/api/admin/login", Category: "Admin Login", ExpectedProtected: false},

	// Member Protected Routes (Auth Required)
	{Method: "GET", Path: "/api/member/profile", Category: "Member Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/member/profile", Category: "Member Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/member/notifications", Category: "Member Protected", ExpectedProtected: true},
	{Method: "PUT", Path: "/api/member/notifications/1/read", Category: "Member Protected", ExpectedProtected: true},
	{Method: "GET", Path: "/api/member/notifications/unread/count", Category: "Member Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/member/change-password", Category: "Member Protected", ExpectedProtected: true},

	// Payment & Subscription Protected Routes (Auth Required)
	{Method: "POST", Path: "/api/payments/create-order", Category: "Payment Protected", ExpectedProtected: true},
	{Method: "POST", Path: "/api/payments/verify", Category: "Payment Protected", ExpectedProtected: true},
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
}

type SecurityAuditResult struct {
	Method            string `json:"method"`
	Path              string `json:"path"`
	Category          string `json:"category"`
	StatusCode        int    `json:"status_code"`
	ExpectedProtected bool   `json:"expected_protected"`
	IsSecured         bool   `json:"is_secured"`
	Flag              string `json:"flag"`
	SecurityStatus    string `json:"security_status"`
}

// TestAPI_06_EndpointSecurity audits all swagger endpoints for unauthenticated access & security controls.
func TestAPI_06_EndpointSecurity(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"API Security Audit - Secured vs Unsecured Endpoints",
		"Audits API endpoints against Swagger spec to ensure protected endpoints return 401/403 without auth headers.",
		"All protected endpoints return HTTP 401/403 Unauthorized, and public endpoints return non-401/403 responses.",
		func(tc *tests.TestContext) {
			baseURL := tests.GlobalConfig.BaseURL
			if baseURL == "" {
				baseURL = "https://api.nammataga.com"
			}

			httpClient := &http.Client{
				Timeout: 10 * time.Second,
			}

			var results []SecurityAuditResult
			var securityRisks []SecurityAuditResult
			var securedCount int
			var publicCount int

			fmt.Println("====================================================================================================")
			fmt.Println(" 🔒 TAGA-API Security Audit - Endpoint Protection Report")
			fmt.Printf(" Target Base URL: %s\n", baseURL)
			fmt.Println("====================================================================================================")
			fmt.Printf("%-6s %-7s %-15s %-45s %-25s\n", "METHOD", "STATUS", "FLAG", "ENDPOINT", "CATEGORY")
			fmt.Println(strings.Repeat("-", 105))

			for _, ep := range SwaggerEndpoints {
				fullURL := strings.TrimRight(baseURL, "/") + ep.Path
				req, err := http.NewRequest(ep.Method, fullURL, nil)
				if err != nil {
					tc.Fatalf("Failed to create request for %s %s: %v", ep.Method, ep.Path, err)
				}

				req.Header.Set("User-Agent", "APISecurityAudit/1.0")
				if ep.Method == "POST" || ep.Method == "PUT" {
					req.Header.Set("Content-Type", "application/json")
				}

				resp, err := httpClient.Do(req)
				statusCode := 0
				if err == nil {
					statusCode = resp.StatusCode
					_, _ = io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}

				// An endpoint is secured if unauthenticated access returns 401 (Unauthorized) or 403 (Forbidden)
				isSecured := (statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden)

				var flag string
				var securityStatus string

				if ep.ExpectedProtected {
					if isSecured {
						flag = "OK"
						securityStatus = "PROTECTED (SECURE)"
						securedCount++
					} else {
						flag = "SECURITY_RISK"
						securityStatus = fmt.Sprintf("UNAUTHENTICATED ACCESS (HTTP %d)", statusCode)
						securityRisks = append(securityRisks, SecurityAuditResult{
							Method:            ep.Method,
							Path:              ep.Path,
							Category:          ep.Category,
							StatusCode:        statusCode,
							ExpectedProtected: ep.ExpectedProtected,
							IsSecured:         isSecured,
							Flag:              flag,
							SecurityStatus:    securityStatus,
						})
					}
				} else {
					if isSecured {
						flag = "WARNING"
						securityStatus = "UNEXPECTEDLY PROTECTED"
					} else {
						flag = "PUBLIC"
						securityStatus = fmt.Sprintf("PUBLIC (HTTP %d)", statusCode)
						publicCount++
					}
				}

				res := SecurityAuditResult{
					Method:            ep.Method,
					Path:              ep.Path,
					Category:          ep.Category,
					StatusCode:        statusCode,
					ExpectedProtected: ep.ExpectedProtected,
					IsSecured:         isSecured,
					Flag:              flag,
					SecurityStatus:    securityStatus,
				}
				results = append(results, res)

				flagSymbol := "🌐 PUBLIC"
				if flag == "SECURITY_RISK" {
					flagSymbol = "⚠️ RISK"
				} else if flag == "OK" {
					flagSymbol = "✅ SECURE"
				}

				fmt.Printf("%-6s %-7d %-15s %-45s %-25s\n", ep.Method, statusCode, flagSymbol, ep.Path, ep.Category)
			}

			fmt.Println("\n====================================================================================================")
			fmt.Println(" 📊 SECURITY AUDIT SUMMARY")
			fmt.Println("====================================================================================================")
			fmt.Printf("Total Endpoints Audited  : %d\n", len(results))
			fmt.Printf("Protected Endpoints (OK) : %d\n", securedCount)
			fmt.Printf("Public Endpoints         : %d\n", publicCount)
			fmt.Printf("Security Risks Detected  : %d\n", len(securityRisks))
			fmt.Println("====================================================================================================")

			// Save Audit JSON Report to evidence directory
			evidenceDir := tests.EvidenceDir
			if evidenceDir != "" {
				reportPath := filepath.Join(evidenceDir, "reports", "endpoint-security-audit.json")
				_ = os.MkdirAll(filepath.Dir(reportPath), 0755)
				data, err := json.MarshalIndent(results, "", "  ")
				if err == nil {
					_ = os.WriteFile(reportPath, data, 0644)
					fmt.Printf("📄 Audit JSON report saved to: %s\n", reportPath)
				}
			}

			tc.Actual = fmt.Sprintf("Audited %d endpoints: %d protected (OK), %d public, %d security risks", len(results), securedCount, publicCount, len(securityRisks))

			if len(securityRisks) > 0 {
				tc.FailureReason = fmt.Sprintf("Security audit detected %d protected endpoints allowing unauthenticated access", len(securityRisks))
				tc.Errorf("Security Audit Failed: %d protected endpoints allow unauthenticated access!", len(securityRisks))
			}
		},
	)
}
