package api_tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// ============================================================================
// POSITIVE API TESTS
// ============================================================================

// TestAPI_About_ValidRequest verifies GET /api/public/about returns 200 OK with valid organization info.
func TestAPI_About_ValidRequest(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about - Valid Request", func(t *testing.T, c *client.Client) {
		var resp AboutResponse
		err := c.SendHttpRequest("GET", "/api/public/about", nil, nil, &resp, nil)
		if err != nil {
			t.Fatalf("Expected 200 OK from /api/public/about, got error: %v", err)
		}

		if resp.ID == 0 {
			t.Errorf("Expected non-zero ID in AboutResponse, got 0")
		}
		if resp.Name == "" {
			t.Errorf("Expected non-empty Name in AboutResponse")
		}
		if resp.Acronym == "" {
			t.Errorf("Expected non-empty Acronym in AboutResponse")
		}
	})
}

// TestAPI_About_ValidHeaders verifies GET /api/public/about succeeds with custom valid request headers.
func TestAPI_About_ValidHeaders(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about - Valid Headers", func(t *testing.T, c *client.Client) {
		headers := map[string]string{
			"Accept":          "application/json",
			"User-Agent":      "TAGA-Test-AutomationClient/1.0",
			"Accept-Language": "en-US,en;q=0.9",
		}
		var resp AboutResponse
		err := c.SendHttpRequest("GET", "/api/public/about", headers, nil, &resp, nil)
		if err != nil {
			t.Fatalf("Failed request with custom valid headers: %v", err)
		}
	})
}

// TestAPI_About_ResponseTime verifies GET /api/public/about responds within acceptable SLA (< 500ms).
func TestAPI_About_ResponseTime(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about - SLA Response Time", func(t *testing.T, c *client.Client) {
		start := time.Now()
		var resp AboutResponse
		err := c.SendHttpRequest("GET", "/api/public/about", nil, nil, &resp, nil)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if duration > 500*time.Millisecond {
			t.Errorf("Response time exceeded SLA: took %v (max allowed: 500ms)", duration)
		}
	})
}

// TestAPI_AboutStats_ValidRequest verifies GET /api/public/about/stats returns 200 OK with statistics list.
func TestAPI_AboutStats_ValidRequest(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about/stats - Valid Request", func(t *testing.T, c *client.Client) {
		var stats []StatsResponse
		err := c.SendHttpRequest("GET", "/api/public/about/stats", nil, nil, &stats, nil)
		if err != nil {
			t.Fatalf("Expected 200 OK from /api/public/about/stats, got error: %v", err)
		}
		if len(stats) == 0 {
			t.Logf("Warning: Stats response list is empty")
		}
	})
}

// TestAPI_AboutObjectives_ValidRequest verifies GET /api/public/about/objectives returns 200 OK.
func TestAPI_AboutObjectives_ValidRequest(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about/objectives - Valid Request", func(t *testing.T, c *client.Client) {
		var objectives []Objective
		err := c.SendHttpRequest("GET", "/api/public/about/objectives", nil, nil, &objectives, nil)
		if err != nil {
			t.Fatalf("Expected 200 OK from /api/public/about/objectives, got error: %v", err)
		}
	})
}

// TestAPI_AboutServices_ValidRequest verifies GET /api/public/about/services returns 200 OK.
func TestAPI_AboutServices_ValidRequest(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about/services - Valid Request", func(t *testing.T, c *client.Client) {
		var services []ServiceResponse
		err := c.SendHttpRequest("GET", "/api/public/about/services", nil, nil, &services, nil)
		if err != nil {
			t.Fatalf("Expected 200 OK from /api/public/about/services, got error: %v", err)
		}
	})
}

// TestAPI_AboutContact_ValidRequest verifies GET /api/public/about/contact returns 200 OK with contact info.
func TestAPI_AboutContact_ValidRequest(t *testing.T) {
	tests.RunAPITest(t, "About API - GET /api/public/about/contact - Valid Request", func(t *testing.T, c *client.Client) {
		var contact ContactResponse
		err := c.SendHttpRequest("GET", "/api/public/about/contact", nil, nil, &contact, nil)
		if err != nil {
			t.Fatalf("Expected 200 OK from /api/public/about/contact, got error: %v", err)
		}
	})
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
		{"Main About Info", "/api/public/about"},
		{"Organization Stats", "/api/public/about/stats"},
		{"Organization Objectives", "/api/public/about/objectives"},
		{"Organization Services", "/api/public/about/services"},
		{"Contact Details", "/api/public/about/contact"},
	}

	for _, ep := range endpoints {
		ep := ep
		tests.RunAPITest(t, "TableDriven - GET "+ep.name, func(t *testing.T, c *client.Client) {
			var raw json.RawMessage
			err := c.SendHttpRequest("GET", ep.path, nil, nil, &raw, nil)
			if err != nil {
				t.Fatalf("Failed GET %s: %v", ep.path, err)
			}
			if len(raw) == 0 {
				t.Errorf("Empty JSON body returned for %s", ep.path)
			}
		})
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
			testName := "MethodValidation - " + method + " " + ep
			tests.RunAPITest(t, testName, func(t *testing.T, c *client.Client) {
				dummyBody := map[string]string{"data": "test"}
				err := c.SendHttpRequest(method, ep, nil, &dummyBody, nil, nil)

				if err == nil {
					t.Errorf("Expected HTTP error for unsupported method %s on %s, got 2xx OK", method, ep)
				} else {
					status := err.StatusCode()
					if status != http.StatusMethodNotAllowed && status != http.StatusNotFound {
						t.Logf("Notice: Method %s on %s returned HTTP status %d", method, ep, status)
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
		tests.RunAPITest(t, "RouteValidation - GET "+path, func(t *testing.T, c *client.Client) {
			err := c.SendHttpRequest("GET", path, nil, nil, nil, nil)
			if err == nil {
				t.Errorf("Expected 404 for invalid route %s, got 2xx", path)
			} else if err.StatusCode() != http.StatusNotFound {
				t.Errorf("Expected HTTP 404 for path %s, got %d", path, err.StatusCode())
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
			name:    "Invalid Content-Type Header",
			headers: map[string]string{"Content-Type": "text/invalid-type"},
		},
		{
			name:    "Invalid Accept Header",
			headers: map[string]string{"Accept": "application/xml"},
		},
		{
			name:    "Dummy Authorization Token on Public Endpoint",
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
	}

	for _, tc := range testCases {
		tc := tc
		tests.RunAPITest(t, "HeaderValidation - "+tc.name, func(t *testing.T, c *client.Client) {
			var resp AboutResponse
			err := c.SendHttpRequest("GET", "/api/public/about", tc.headers, nil, &resp, nil)
			if err != nil {
				t.Errorf("Public endpoint failed under header case '%s': %v", tc.name, err)
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
			testName := "SecurityInjection - " + ep + " " + p.name
			tests.RunAPITest(t, testName, func(t *testing.T, c *client.Client) {
				fullPath := ep + p.query
				var raw json.RawMessage
				err := c.SendHttpRequest("GET", fullPath, nil, nil, &raw, nil)

				if err != nil && err.StatusCode() == http.StatusInternalServerError {
					t.Errorf("Security injection test '%s' caused 500 Internal Server Error: %v", p.name, err)
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
	}

	for _, tc := range testCases {
		tc := tc
		tests.RunAPITest(t, "UnexpectedPayload - "+tc.name, func(t *testing.T, c *client.Client) {
			body := tc.payload
			var resp AboutResponse
			err := c.SendHttpRequest("GET", "/api/public/about", nil, &body, &resp, nil)

			if err != nil {
				t.Errorf("Sending body on GET request caused error: %v", err)
			}
		})
	}
}
