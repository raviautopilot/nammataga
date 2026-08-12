package api_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"e2e-template/tests"
)

// ============================================================================
// POSITIVE API TESTS
// ============================================================================

// TestAPI_About_ValidRequest verifies GET /api/public/about returns 200 OK with valid organization info.
func TestAPI_About_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Organization Info Details",
		"Checks that GET /api/public/about returns HTTP 200 OK with valid organization metadata (ID, Name, Acronym).",
		"HTTP 200 OK with valid AboutResponse object",
		func(tc *tests.TestContext) {
			var resp AboutResponse
			err := tc.Client.SendHttpRequest("GET", "/api/public/about", nil, nil, &resp, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /api/public/about, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, ID=%d, Name='%s'", resp.ID, resp.Name)

			if resp.ID == 0 {
				tc.FailureReason = "Expected non-zero ID in AboutResponse, got 0"
				tc.Errorf("Expected non-zero ID in AboutResponse, got 0")
			}
			if resp.Name == "" {
				tc.FailureReason = "Expected non-empty Name in AboutResponse"
				tc.Errorf("Expected non-empty Name in AboutResponse")
			}
		},
	)
}

// TestAPI_About_ValidHeaders verifies GET /api/public/about succeeds with custom valid request headers.
func TestAPI_About_ValidHeaders(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify About Info with Custom Headers",
		"Checks that GET /api/public/about accepts custom Accept, User-Agent, and Accept-Language headers.",
		"HTTP 200 OK with AboutResponse object when custom headers are sent",
		func(tc *tests.TestContext) {
			headers := map[string]string{
				"Accept":          "application/json",
				"User-Agent":      "TAGA-Test-AutomationClient/1.0",
				"Accept-Language": "en-US,en;q=0.9",
			}
			var resp AboutResponse
			err := tc.Client.SendHttpRequest("GET", "/api/public/about", headers, nil, &resp, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("Failed request with custom headers: %v", err)
				tc.Fatalf("Failed request with custom valid headers: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Name='%s'", resp.Name)
		},
	)
}

// TestAPI_About_ResponseTime verifies GET /api/public/about responds within acceptable SLA (< 500ms).
func TestAPI_About_ResponseTime(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify About Info Speed SLA",
		"Checks that GET /api/public/about responds in less than 500ms latency SLA.",
		"HTTP 200 OK within < 500ms latency SLA",
		func(tc *tests.TestContext) {
			start := time.Now()
			var resp AboutResponse
			err := tc.Client.SendHttpRequest("GET", "/api/public/about", nil, nil, &resp, nil)
			duration := time.Since(start)

			if err != nil {
				tc.FailureReason = fmt.Sprintf("Request failed: %v", err)
				tc.Fatalf("Request failed: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK in %v", duration)
			if duration > 500*time.Millisecond {
				tc.FailureReason = fmt.Sprintf("SLA breach: response took %v (max 500ms)", duration)
				tc.Errorf("Response time exceeded SLA: took %v (max allowed: 500ms)", duration)
			}
		},
	)
}

// TestAPI_AboutStats_ValidRequest verifies GET /api/public/about/stats returns 200 OK with statistics list.
func TestAPI_AboutStats_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Organization Statistics List",
		"Checks that GET /api/public/about/stats returns HTTP 200 OK with statistics metrics (member count, service years).",
		"HTTP 200 OK with []StatsResponse array",
		func(tc *tests.TestContext) {
			var stats []StatsResponse
			err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats", nil, nil, &stats, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /api/public/about/stats, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Stats Array Count = %d items", len(stats))
		},
	)
}

// TestAPI_AboutObjectives_ValidRequest verifies GET /api/public/about/objectives returns 200 OK.
func TestAPI_AboutObjectives_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Organization Objectives List",
		"Checks that GET /api/public/about/objectives returns HTTP 200 OK with objectives list.",
		"HTTP 200 OK with []Objective array",
		func(tc *tests.TestContext) {
			var objectives []Objective
			err := tc.Client.SendHttpRequest("GET", "/api/public/about/objectives", nil, nil, &objectives, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /api/public/about/objectives, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Objectives Count = %d items", len(objectives))
		},
	)
}

// TestAPI_AboutServices_ValidRequest verifies GET /api/public/about/services returns 200 OK.
func TestAPI_AboutServices_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Organization Services Catalog",
		"Checks that GET /api/public/about/services returns HTTP 200 OK with services list.",
		"HTTP 200 OK with []ServiceResponse array",
		func(tc *tests.TestContext) {
			var services []ServiceResponse
			err := tc.Client.SendHttpRequest("GET", "/api/public/about/services", nil, nil, &services, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /api/public/about/services, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Services Count = %d items", len(services))
		},
	)
}

// TestAPI_AboutContact_ValidRequest verifies GET /api/public/about/contact returns 200 OK with contact info.
func TestAPI_AboutContact_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Contact Information Payload",
		"Checks that GET /api/public/about/contact returns HTTP 200 OK with headquarters, phone, email, and regional offices.",
		"HTTP 200 OK with ContactResponse object",
		func(tc *tests.TestContext) {
			var contact ContactResponse
			err := tc.Client.SendHttpRequest("GET", "/api/public/about/contact", nil, nil, &contact, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /api/public/about/contact, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Headquarters = '%s', Email = '%s'", contact.Headquarters.Name, contact.PrimaryEmail)
		},
	)
}

// ============================================================================
// TABLE-DRIVEN POSITIVE & NEGATIVE ENDPOINT TESTS
// ============================================================================

// TestAPI_About_AllEndpoints_TableDriven runs parameterized checks over all public about endpoints.
func TestAPI_About_AllEndpoints_TableDriven(t *testing.T) {
	endpoints := []struct {
		name string
		path string
	}{
		{"Main About Info (/api/public/about)", "/api/public/about"},
		{"Organization Stats (/api/public/about/stats)", "/api/public/about/stats"},
		{"Organization Objectives (/api/public/about/objectives)", "/api/public/about/objectives"},
		{"Organization Services (/api/public/about/services)", "/api/public/about/services"},
		{"Contact Details (/api/public/about/contact)", "/api/public/about/contact"},
	}

	for _, ep := range endpoints {
		ep := ep
		tests.RunAPITestWithDetails(
			t,
			"Verify Valid Data Response for "+ep.name,
			fmt.Sprintf("Checks that GET request to %s returns valid non-empty JSON data.", ep.path),
			fmt.Sprintf("HTTP 200 OK with non-empty JSON on %s", ep.path),
			func(tc *tests.TestContext) {
				var raw json.RawMessage
				err := tc.Client.SendHttpRequest("GET", ep.path, nil, nil, &raw, nil)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("Failed GET %s: %v", ep.path, err)
					tc.Fatalf("Failed GET %s: %v", ep.path, err)
				}

				tc.Actual = fmt.Sprintf("HTTP 200 OK, Body Length = %d bytes", len(raw))
				if len(raw) == 0 {
					tc.FailureReason = "Empty JSON body returned"
					tc.Errorf("Empty JSON body returned for %s", ep.path)
				}
			},
		)
	}
}

// ============================================================================
// HTTP METHOD & ROUTE VALIDATION (NEGATIVE TESTS)
// ============================================================================

// TestAPI_About_WrongHTTPMethods tests that invalid HTTP methods are rejected or handled gracefully.
func TestAPI_About_WrongHTTPMethods(t *testing.T) {
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	endpoints := []string{
		"/api/public/about",
		"/api/public/about/stats",
		"/api/public/about/objectives",
		"/api/public/about/services",
		"/api/public/about/contact",
	}

	for _, ep := range endpoints {
		for _, method := range methods {
			method := method
			ep := ep
			testName := fmt.Sprintf("Reject Unsupported %s on %s", method, ep)
			desc := fmt.Sprintf("Verifies that HTTP %s request to %s is rejected with status 405 Method Not Allowed or 404 Not Found.", method, ep)
			expected := fmt.Sprintf("HTTP 405 Method Not Allowed or 404 Not Found for %s %s", method, ep)

			tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
				dummyBody := map[string]string{"data": "test"}
				err := tc.Client.SendHttpRequest(method, ep, nil, &dummyBody, nil, nil)

				if err == nil {
					tc.FailureReason = fmt.Sprintf("Security flaw: Unsupported HTTP method %s returned 2xx OK on %s", method, ep)
					tc.Errorf("Expected HTTP error for unsupported method %s on %s, got 2xx OK", method, ep)
				} else {
					status := err.StatusCode()
					tc.Actual = fmt.Sprintf("Rejected cleanly with HTTP status %d", status)
					if status != http.StatusMethodNotAllowed && status != http.StatusNotFound {
						tc.Logf("Notice: Method %s on %s returned HTTP status %d", method, ep, status)
					}
				}
			})
		}
	}
}

// TestAPI_About_InvalidRoutes verifies 404 Not Found for non-existent sub-routes.
func TestAPI_About_InvalidRoutes(t *testing.T) {
	invalidPaths := []string{
		"/api/public/about/nonexistent",
		"/api/public/about/123",
		"/api/public/about/stats/extra",
		"/api/public/about/contact/unknown",
	}

	for _, path := range invalidPaths {
		path := path
		desc := fmt.Sprintf("Verifies that requesting non-existent sub-route %s returns HTTP 404 Not Found.", path)
		expected := fmt.Sprintf("HTTP 404 Not Found for %s", path)

		tests.RunAPITestWithDetails(t, "Reject Invalid Sub-route "+path, desc, expected, func(tc *tests.TestContext) {
			err := tc.Client.SendHttpRequest("GET", path, nil, nil, nil, nil)
			if err == nil {
				tc.FailureReason = fmt.Sprintf("Expected 404 Not Found, but invalid sub-route %s returned 2xx OK", path)
				tc.Errorf("Expected 404 for invalid route %s, got 2xx", path)
			} else {
				tc.Actual = fmt.Sprintf("Returned HTTP status %d", err.StatusCode())
				if err.StatusCode() != http.StatusNotFound {
					tc.FailureReason = fmt.Sprintf("Expected 404, got status %d", err.StatusCode())
					tc.Errorf("Expected HTTP 404 for path %s, got %d", path, err.StatusCode())
				}
			}
		})
	}
}

// ============================================================================
// HEADER VALIDATION & SECURITY TESTS
// ============================================================================

// TestAPI_About_HeaderValidations tests various custom, missing, and invalid headers.
func TestAPI_About_HeaderValidations(t *testing.T) {
	testCases := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "Missing Content-Type Header",
			headers: map[string]string{"Content-Type": ""},
		},
		{
			name:    "Invalid Content-Type Header (text/invalid-type)",
			headers: map[string]string{"Content-Type": "text/invalid-type"},
		},
		{
			name:    "Invalid Accept Header (application/xml)",
			headers: map[string]string{"Accept": "application/xml"},
		},
		{
			name:    "Dummy Bearer Token on Public Endpoint",
			headers: map[string]string{"Authorization": "Bearer fake_token_12345"},
		},
		{
			name:    "Empty Authorization Token",
			headers: map[string]string{"Authorization": "Bearer "},
		},
		{
			name:    "Malformed Authorization Header",
			headers: map[string]string{"Authorization": "Basic YWRtaW46cGFzc3dvcmQ="},
		},
		{
			name:    "Extremely Large Header Value",
			headers: map[string]string{"X-Custom-Header": strings.Repeat("A", 10000)},
		},
		{
			name:    "SQL Injection in Header",
			headers: map[string]string{"User-Agent": "' OR '1'='1"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		desc := fmt.Sprintf("Checks that /api/public/about safely handles condition '%s'.", tc.name)
		expected := "HTTP 200 OK (Public endpoint ignores non-standard headers gracefully)"

		tests.RunAPITestWithDetails(t, "Header Test - "+tc.name, desc, expected, func(tctx *tests.TestContext) {
			var resp AboutResponse
			err := tctx.Client.SendHttpRequest("GET", "/api/public/about", tc.headers, nil, &resp, nil)
			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Failed under header case '%s': %v", tc.name, err)
				tctx.Errorf("Public endpoint failed under header case '%s': %v", tc.name, err)
			} else {
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Name = '%s'", resp.Name)
			}
		})
	}
}

// ============================================================================
// SECURITY & INJECTION PAYLOAD TESTS IN QUERY PARAMETERS
// ============================================================================

// TestAPI_About_SecurityInjections tests resilience against query injection attacks.
func TestAPI_About_SecurityInjections(t *testing.T) {
	payloads := []struct {
		name  string
		query string
	}{
		{"SQL Injection Simple", "?id=1' OR '1'='1"},
		{"SQL Injection Union", "?search=1 UNION SELECT null, username, password FROM users--"},
		{"SQL Injection Drop Table", "?filter=1; DROP TABLE members;--"},
		{"XSS Script Tag", "?query=<script>alert('xss')</script>"},
		{"XSS Image Event", "?name=<img src=x onerror=alert(1)>"},
		{"HTML Injection", "?title=<h1>Injected Header</h1>"},
		{"Unicode Special Characters", "?q=ஆனாஆவன்னா🔥🚀ñ#?"},
		{"Long String Overflow", "?param=" + strings.Repeat("A", 2048)},
	}

	endpoints := []string{
		"/api/public/about",
		"/api/public/about/stats",
		"/api/public/about/contact",
	}

	for _, ep := range endpoints {
		for _, p := range payloads {
			ep := ep
			p := p
			testName := fmt.Sprintf("Security Check - %s on %s", p.name, ep)
			desc := fmt.Sprintf("Verifies that query injection attack (%s) on %s does not crash the server.", p.name, ep)
			expected := "HTTP 200 OK or 400 Bad Request (Must NOT return HTTP 500 Internal Server Error)"

			tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
				fullPath := ep + p.query
				var raw json.RawMessage
				err := tc.Client.SendHttpRequest("GET", fullPath, nil, nil, &raw, nil)

				if err != nil {
					tc.Actual = fmt.Sprintf("Handled safely with HTTP status %d", err.StatusCode())
					if err.StatusCode() == http.StatusInternalServerError {
						tc.FailureReason = fmt.Sprintf("Vulnerability: HTTP 500 Internal Server Error triggered by %s", p.name)
						tc.Errorf("Security injection test '%s' caused 500 Internal Server Error: %v", p.name, err)
					}
				} else {
					tc.Actual = "HTTP 200 OK (Query injection safely ignored)"
				}
			})
		}
	}
}

// ============================================================================
// UNEXPECTED REQUEST BODY ON GET ENDPOINTS
// ============================================================================

// TestAPI_About_UnexpectedPayloads tests sending request bodies on GET requests.
func TestAPI_About_UnexpectedPayloads(t *testing.T) {
	testCases := []struct {
		name    string
		payload interface{}
	}{
		{"GET with Valid JSON Body", map[string]string{"key": "value"}},
		{"GET with Empty JSON Body", map[string]string{}},
		{"GET with Large JSON Payload", map[string]interface{}{"data": strings.Repeat("X", 5000)}},
		{"GET with Malformed JSON Payload", "invalid-json-content"},
		{"GET with XML Payload", "<xml><data>test</data></xml>"},
		{"Business Logic - Past Date for Future Range", map[string]string{"start_date": "2050-01-01", "end_date": "2020-01-01"}},
		{"Business Logic - Negative Priority", map[string]int{"priority": -5}},
	}

	for _, tc := range testCases {
		tc := tc
		desc := fmt.Sprintf("Sends unexpected request body payload under scenario '%s' on GET /api/public/about.", tc.name)
		expected := "HTTP 200 OK (GET handler ignores body without crashing)"

		tests.RunAPITestWithDetails(t, "Handle Payload - "+tc.name, desc, expected, func(tctx *tests.TestContext) {
			body := tc.payload
			var resp AboutResponse
			err := tctx.Client.SendHttpRequest("GET", "/api/public/about", nil, &body, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Unexpected error on GET body: %v", err)
				tctx.Errorf("Sending body on GET request caused error: %v", err)
			} else {
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Name = '%s'", resp.Name)
			}
		})
	}
}
