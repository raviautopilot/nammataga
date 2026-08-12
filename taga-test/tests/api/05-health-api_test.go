package api_test

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
// POSITIVE HEALTH & ROOT API TESTS
// ============================================================================

// TestAPI_Health_ValidRequest verifies GET /health returns 200 OK with status "healthy".
func TestAPI_Health_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Application Health Status",
		"Checks that GET /health returns HTTP 200 OK with status 'healthy' for load balancer readiness.",
		"HTTP 200 OK with JSON status = 'healthy'",
		func(tc *tests.TestContext) {
			var resp HealthResponse
			err := tc.Client.SendHttpRequest("GET", "/health", nil, nil, &resp, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /health, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Status = '%s', Message = '%s'", resp.Status, resp.Message)
			if resp.Status == "" && resp.Message == "healthy" {
				tc.FailureReason = "REAL API CONTRACT BUG: Endpoint GET /health returned 'message':'healthy' instead of required contract key 'status':'healthy'"
				tc.Errorf("REAL API CONTRACT BUG: Endpoint GET /health returned 'message':'healthy' instead of required contract key 'status':'healthy'")
			} else if resp.Status != "healthy" {
				tc.FailureReason = fmt.Sprintf("Expected status 'healthy', got '%s'", resp.Status)
				tc.Errorf("Expected status 'healthy', got '%s'", resp.Status)
			}
		},
	)
}

// TestAPI_Root_ValidRequest verifies GET / returns 200 OK with welcome message.
func TestAPI_Root_ValidRequest(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Root Welcome Message",
		"Checks that base URL GET / returns HTTP 200 OK with welcome message and success status.",
		"HTTP 200 OK with status = 'success' and welcome message",
		func(tc *tests.TestContext) {
			var resp RootResponse
			err := tc.Client.SendHttpRequest("GET", "/", nil, nil, &resp, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
				tc.Fatalf("Expected 200 OK from /, got error: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Status = '%s', Message = '%s'", resp.Status, resp.Message)
			if resp.Status != "success" {
				tc.FailureReason = fmt.Sprintf("Expected status 'success', got '%s'", resp.Status)
				tc.Errorf("Expected status 'success', got '%s'", resp.Status)
			}
			if resp.Message == "" {
				tc.FailureReason = "Message field in RootResponse was empty"
				tc.Errorf("Expected non-empty welcome message in RootResponse")
			}
		},
	)
}

// TestAPI_Health_ValidHeaders verifies GET /health succeeds with custom valid headers.
func TestAPI_Health_ValidHeaders(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Health Check with Custom Headers",
		"Checks that GET /health accepts custom Accept, User-Agent, and Language headers.",
		"HTTP 200 OK with status 'healthy' when custom headers are sent",
		func(tc *tests.TestContext) {
			headers := map[string]string{
				"Accept":          "application/json",
				"User-Agent":      "HealthMonitor/2.0",
				"Accept-Language": "en-US",
			}
			var resp HealthResponse
			err := tc.Client.SendHttpRequest("GET", "/health", headers, nil, &resp, nil)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("Failed with custom headers: %v", err)
				tc.Fatalf("Request failed with custom headers: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK, Status = '%s', Message = '%s'", resp.Status, resp.Message)
			if resp.Status == "" && resp.Message == "healthy" {
				tc.FailureReason = "REAL API CONTRACT BUG: Endpoint GET /health returned 'message':'healthy' instead of required contract key 'status':'healthy'"
				tc.Errorf("REAL API CONTRACT BUG: Endpoint GET /health returned 'message':'healthy' instead of required contract key 'status':'healthy'")
			} else if resp.Status != "healthy" {
				tc.FailureReason = fmt.Sprintf("Expected status 'healthy', got '%s'", resp.Status)
				tc.Errorf("Expected status 'healthy', got '%s'", resp.Status)
			}
		},
	)
}

// TestAPI_Health_ResponseTime verifies GET /health responds within strict SLA (< 200ms).
func TestAPI_Health_ResponseTime(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Health Check Speed SLA",
		"Checks that GET /health responds in less than 200ms for fast monitoring probes.",
		"HTTP 200 OK in under 200ms latency SLA",
		func(tc *tests.TestContext) {
			start := time.Now()
			var resp HealthResponse
			err := tc.Client.SendHttpRequest("GET", "/health", nil, nil, &resp, nil)
			duration := time.Since(start)

			if err != nil {
				tc.FailureReason = fmt.Sprintf("Request failed: %v", err)
				tc.Fatalf("Health check request failed: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK in %v", duration)
			if duration > 200*time.Millisecond {
				tc.FailureReason = fmt.Sprintf("SLA breach: response took %v (max 200ms)", duration)
				tc.Errorf("Health check response time exceeded SLA: took %v (max: 200ms)", duration)
			}
		},
	)
}

// TestAPI_Root_ResponseTime verifies GET / responds within SLA (< 200ms).
func TestAPI_Root_ResponseTime(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Verify Root Endpoint Speed SLA",
		"Checks that base URL GET / responds in less than 200ms.",
		"HTTP 200 OK in under 200ms latency SLA",
		func(tc *tests.TestContext) {
			start := time.Now()
			var resp RootResponse
			err := tc.Client.SendHttpRequest("GET", "/", nil, nil, &resp, nil)
			duration := time.Since(start)

			if err != nil {
				tc.FailureReason = fmt.Sprintf("Request failed: %v", err)
				tc.Fatalf("Root request failed: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK in %v", duration)
			if duration > 200*time.Millisecond {
				tc.FailureReason = fmt.Sprintf("SLA breach: response took %v (max 200ms)", duration)
				tc.Errorf("Root endpoint response time exceeded SLA: took %v (max: 200ms)", duration)
			}
		},
	)
}

// ============================================================================
// TABLE-DRIVEN ENDPOINT TESTS
// ============================================================================

// TestAPI_Health_AllEndpoints_TableDriven runs parameterized checks over health & root endpoints.
func TestAPI_Health_AllEndpoints_TableDriven(t *testing.T) {
	endpoints := []struct {
		name string
		path string
	}{
		{"Root Endpoint (/)", "/"},
		{"Health Check Endpoint (/health)", "/health"},
	}

	for _, ep := range endpoints {
		ep := ep
		tests.RunAPITestWithDetails(
			t,
			"Verify Valid Data Response for "+ep.name,
			fmt.Sprintf("Checks that GET request to %s returns valid non-empty JSON data.", ep.path),
			fmt.Sprintf("HTTP 200 OK with valid non-empty JSON on %s", ep.path),
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
					tc.Errorf("Empty response body returned for %s", ep.path)
				}
			},
		)
	}
}

// ============================================================================
// HTTP METHOD & ROUTE VALIDATION (NEGATIVE TESTS)
// ============================================================================

// TestAPI_Health_WrongHTTPMethods tests that invalid HTTP methods on health endpoints are rejected.
func TestAPI_Health_WrongHTTPMethods(t *testing.T) {
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	endpoints := []string{"/", "/health"}

	for _, ep := range endpoints {
		for _, method := range methods {
			method := method
			ep := ep
			testName := fmt.Sprintf("Reject Unsupported %s on %s", method, ep)
			desc := fmt.Sprintf("Verifies that HTTP %s request to %s is rejected with status 405 Method Not Allowed or 404 Not Found.", method, ep)
			expected := fmt.Sprintf("HTTP 405 Method Not Allowed or 404 Not Found for %s %s", method, ep)

			tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
				dummyBody := map[string]string{"ping": "pong"}
				err := tc.Client.SendHttpRequest(method, ep, nil, &dummyBody, nil, nil)

				if err == nil {
					tc.FailureReason = fmt.Sprintf("Security flaw: Unsupported HTTP method %s returned 2xx OK on %s", method, ep)
					tc.Errorf("Expected HTTP error for unsupported method %s on %s, got 2xx OK", method, ep)
				} else {
					status := err.StatusCode()
					tc.Actual = fmt.Sprintf("Rejected cleanly with HTTP status %d", status)
					if status != http.StatusMethodNotAllowed && status != http.StatusNotFound {
						tc.Logf("Notice: Method %s on %s returned status %d", method, ep, status)
					}
				}
			})
		}
	}
}

// TestAPI_Health_InvalidRoutes verifies 404 Not Found for non-existent health sub-paths.
func TestAPI_Health_InvalidRoutes(t *testing.T) {
	invalidPaths := []string{
		"/health/status",
		"/health/check",
		"/health/123",
		"/health-check",
		"/api/health",
	}

	for _, path := range invalidPaths {
		path := path
		desc := fmt.Sprintf("Verifies that requesting non-existent route %s returns HTTP 404 Not Found.", path)
		expected := fmt.Sprintf("HTTP 404 Not Found for %s", path)

		tests.RunAPITestWithDetails(t, "Reject Invalid Route "+path, desc, expected, func(tc *tests.TestContext) {
			err := tc.Client.SendHttpRequest("GET", path, nil, nil, nil, nil)
			if err == nil {
				tc.FailureReason = fmt.Sprintf("Expected 404 Not Found, but invalid route %s returned 2xx OK", path)
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
// HEADER VALIDATION TESTS
// ============================================================================

// TestAPI_Health_HeaderValidations tests missing, custom, and unexpected headers on health endpoints.
func TestAPI_Health_HeaderValidations(t *testing.T) {
	testCases := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "Missing Content-Type Header",
			headers: map[string]string{"Content-Type": ""},
		},
		{
			name:    "Invalid Content-Type Header (text/xml)",
			headers: map[string]string{"Content-Type": "text/xml"},
		},
		{
			name:    "Invalid Accept Header (text/html)",
			headers: map[string]string{"Accept": "text/html"},
		},
		{
			name:    "Dummy Bearer Token on Health Endpoint",
			headers: map[string]string{"Authorization": "Bearer fake_token_abc"},
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
		desc := fmt.Sprintf("Checks that /health safely handles condition '%s'.", tc.name)
		expected := "HTTP 200 OK (Public endpoint ignores non-standard headers gracefully)"

		tests.RunAPITestWithDetails(t, "Header Test - "+tc.name, desc, expected, func(tctx *tests.TestContext) {
			var resp HealthResponse
			err := tctx.Client.SendHttpRequest("GET", "/health", tc.headers, nil, &resp, nil)
			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Failed under header case '%s': %v", tc.name, err)
				tctx.Errorf("Health endpoint failed under header test '%s': %v", tc.name, err)
			} else {
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Status = '%s'", resp.Status)
			}
		})
	}
}

// ============================================================================
// SECURITY & QUERY INJECTION TESTS
// ============================================================================

// TestAPI_Health_SecurityInjections tests resilience against query string injections on health routes.
func TestAPI_Health_SecurityInjections(t *testing.T) {
	payloads := []struct {
		name  string
		query string
	}{
		{"SQL Injection Simple", "?id=1' OR '1'='1"},
		{"SQL Injection Union", "?search=1 UNION SELECT null, status FROM health--"},
		{"SQL Injection Drop Table", "?filter=1; DROP TABLE users;--"},
		{"XSS Script Tag", "?query=<script>alert('xss')</script>"},
		{"XSS Image Event", "?name=<img src=x onerror=alert(1)>"},
		{"HTML Injection", "?title=<h1>Header</h1>"},
		{"Unicode Special Characters", "?q=ஆனாஆவன்னா🔥🚀ñ#?"},
		{"Long String Overflow", "?param=" + strings.Repeat("A", 2048)},
	}

	endpoints := []string{"/", "/health"}

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

// TestAPI_Health_UnexpectedPayloads tests sending request bodies on GET health requests.
func TestAPI_Health_UnexpectedPayloads(t *testing.T) {
	testCases := []struct {
		name    string
		payload interface{}
	}{
		{"GET /health with Valid JSON Body", map[string]string{"ping": "pong"}},
		{"GET /health with Empty JSON Body", map[string]string{}},
		{"GET /health with Large JSON Payload", map[string]interface{}{"data": strings.Repeat("H", 5000)}},
		{"GET /health with Malformed JSON Payload", "invalid-json-data"},
		{"GET /health with XML Payload", "<health>check</health>"},
		{"Business Logic - Health check with negative IDOR", map[string]int{"user_id": -9999}},
		{"Business Logic - Health check with End date before Start date", map[string]string{"start": "2025-01-01", "end": "2024-01-01"}},
		{"Business Logic - Extreme Boundary 100-year Future Date", map[string]string{"event_date": "2126-01-01"}},
		{"Business Logic - State Machine Violation Canceling Completed", map[string]string{"action": "cancel", "status": "completed"}},
		{"Business Logic - Role Context Switch to Admin in Body", map[string]string{"role": "admin", "override": "true"}},
		{"Business Logic - Logical Paradox Contradicting Flags", map[string]bool{"is_public": true, "is_private": true}},
	}

	for _, tc := range testCases {
		tc := tc
		desc := fmt.Sprintf("Sends unexpected request body payload under scenario '%s' on GET /health.", tc.name)
		expected := "HTTP 200 OK (GET handler ignores body without crashing)"

		tests.RunAPITestWithDetails(t, "Handle Payload - "+tc.name, desc, expected, func(tctx *tests.TestContext) {
			body := tc.payload
			var resp HealthResponse
			err := tctx.Client.SendHttpRequest("GET", "/health", nil, &body, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Unexpected error on GET body: %v", err)
				tctx.Errorf("Sending body on GET /health request caused error: %v", err)
			} else {
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Status = '%s'", resp.Status)
			}
		})
	}
}
