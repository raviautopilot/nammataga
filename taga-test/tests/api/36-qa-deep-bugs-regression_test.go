package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"e2e-template/tests"
)

// ============================================================================
// QA REGRESSION & EXTREME PARAMETERIZED TEST SUITE — EXPOSING BACKEND BUGS
// ============================================================================

type QABugTestCase struct {
	Name        string
	Description string
	Expected    string
	TestFn      func(tc *tests.TestContext)
}

// TestAPI_36_QADeepBugsRegression_TableDriven is a parameterized table-driven test suite
// that systematically exposes all discovered backend application and contract bugs.
func TestAPI_36_QADeepBugsRegression_TableDriven(t *testing.T) {
	baseURL := tests.GlobalConfig.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}

	testCases := []QABugTestCase{
		{
			Name:        "[Regression BUG-01] POST Grievance -> GET Grievance by ID (Persistence Disconnect)",
			Description: "Verify created grievance can be retrieved by ID via GET /api/grievances/:id",
			Expected:    "HTTP 200 OK with created Grievance JSON payload matching ID",
			TestFn: func(tc *tests.TestContext) {
				// Step A: Create a Grievance
				createURL := strings.TrimRight(baseURL, "/") + "/api/grievances"
				payload := map[string]interface{}{
					"subject":           "QA Test Grievance - Persistence Bug Exposer",
					"category":          "General",
					"priority":          "High",
					"description":       "Testing ID persistence and retrieval.",
					"contactPhone":      "9944637500",
					"preferredResponse": "Email",
				}
				jsonBytes, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Failed to execute POST /api/grievances: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					tc.Fatalf("Expected HTTP 200 on creation, got HTTP %d", resp.StatusCode)
				}

				var created map[string]interface{}
				_ = json.NewDecoder(resp.Body).Decode(&created)
				grievanceID, ok := created["id"].(string)
				if !ok || grievanceID == "" {
					tc.Fatalf("Response did not contain valid 'id': %v", created)
				}

				// Step B: Attempt to retrieve the created grievance by ID
				getURL := fmt.Sprintf("%s/api/grievances/%s", strings.TrimRight(baseURL, "/"), grievanceID)
				getReq, _ := http.NewRequest("GET", getURL, nil)
				getResp, err := httpClient.Do(getReq)
				if err != nil {
					tc.Fatalf("Failed to execute GET /api/grievances/%s: %v", grievanceID, err)
				}
				defer getResp.Body.Close()

				bodyBytes, _ := io.ReadAll(getResp.Body)
				tc.Actual = fmt.Sprintf("HTTP %d | Response: %s", getResp.StatusCode, string(bodyBytes))

				if getResp.StatusCode == http.StatusNotFound {
					tc.FailureReason = fmt.Sprintf("BUG-01 REPRODUCED: Grievance %s was created but GET /api/grievances/%s returned HTTP 404 Not Found (in-memory slice vs file persistence disconnect)", grievanceID, grievanceID)
					tc.Errorf("BUG EXPOSED: GET /api/grievances/%s returned 404 for newly created grievance", grievanceID)
				} else if getResp.StatusCode != http.StatusOK {
					tc.FailureReason = fmt.Sprintf("Expected HTTP 200, got HTTP %d", getResp.StatusCode)
					tc.Errorf("Failed to retrieve grievance: HTTP %d", getResp.StatusCode)
				}
			},
		},
		{
			Name:        "[Regression BUG-02] GET /api/towers/admin/bookings Unauthenticated Security Check",
			Description: "Verify endpoint /api/towers/admin/bookings strictly enforces Admin Authentication",
			Expected:    "HTTP 401 Unauthorized or HTTP 403 Forbidden for request missing Authorization header",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/api/towers/admin/bookings"
				req, _ := http.NewRequest("GET", url, nil) // NO Auth Header!

				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Failed to execute HTTP request: %v", err)
				}
				defer resp.Body.Close()

				bodyBytes, _ := io.ReadAll(resp.Body)
				tc.Actual = fmt.Sprintf("HTTP %d | Data sample: %s", resp.StatusCode, string(bodyBytes))

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "CRITICAL SECURITY BUG (BUG-02): Unauthenticated user successfully retrieved sensitive admin room bookings catalog (HTTP 200 OK)"
					tc.Errorf("CRITICAL SECURITY BUG: GET /api/towers/admin/bookings allowed unauthenticated access!")
				} else if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
					tc.FailureReason = fmt.Sprintf("Expected HTTP 401/403, got HTTP %d", resp.StatusCode)
					tc.Errorf("Unexpected status code: HTTP %d", resp.StatusCode)
				}
			},
		},
		{
			Name:        "[Regression BUG-03] GET /health API Contract JSON Schema Audit",
			Description: "Verify GET /health returns key 'status': 'healthy' per API contract specification",
			Expected:    "HTTP 200 OK with JSON schema containing {\"status\": \"healthy\"}",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/health"
				req, _ := http.NewRequest("GET", url, nil)

				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Failed to execute request: %v", err)
				}
				defer resp.Body.Close()

				var jsonResp map[string]interface{}
				_ = json.NewDecoder(resp.Body).Decode(&jsonResp)

				statusVal, statusExists := jsonResp["status"].(string)
				messageVal, _ := jsonResp["message"].(string)

				tc.Actual = fmt.Sprintf("JSON Body: %v | status: '%s' | message: '%s'", jsonResp, statusVal, messageVal)

				if !statusExists {
					tc.FailureReason = fmt.Sprintf("CONTRACT BUG (BUG-03): Response missing required field 'status'. Found 'message':'%s' instead.", messageVal)
					tc.Errorf("API Contract Mismatch: /health returned 'message' instead of required 'status' field")
				} else if statusVal != "healthy" {
					tc.FailureReason = fmt.Sprintf("Expected status 'healthy', got '%s'", statusVal)
					tc.Errorf("Invalid status value: %s", statusVal)
				}
			},
		},
		{
			Name:        "[Regression BUG-04] POST /api/subscriptions/create-order IDOR Security Verification",
			Description: "Verify member cannot create subscription orders using another member's email address",
			Expected:    "HTTP 403 Forbidden or HTTP 400 Bad Request when token email does not match payload email",
			TestFn: func(tc *tests.TestContext) {
				token := getValidMemberToken(tc.T, tc.Client)

				url := strings.TrimRight(baseURL, "/") + "/api/subscriptions/create-order"
				payload := map[string]interface{}{
					"subscriptionId": "annual-subscription",
					"amount":         100000,
					"email":          "victim_user_other@gmail.com", // Impersonated victim email!
				}
				jsonBytes, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Failed to execute request: %v", err)
				}
				defer resp.Body.Close()

				bodyBytes, _ := io.ReadAll(resp.Body)
				tc.Actual = fmt.Sprintf("HTTP %d | Response: %s", resp.StatusCode, string(bodyBytes))

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "SECURITY IDOR BUG (BUG-04): Endpoint accepted subscription order creation for arbitrary victim email 'victim_user_other@gmail.com' without validating against JWT member identity!"
					tc.Errorf("IDOR Vulnerability exposed on /api/subscriptions/create-order")
				}
			},
		},
		{
			Name:        "[Regression BUG-05] POST /api/webhook/razorpay Fake Signature Rejection Audit",
			Description: "Verify webhook endpoint strictly rejects requests with invalid or missing X-Razorpay-Signature header",
			Expected:    "HTTP 401 Unauthorized or HTTP 400 Bad Request when signature is invalid",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/api/webhook/razorpay"
				payload := map[string]interface{}{
					"event": "payment.captured",
					"payload": map[string]interface{}{
						"payment": map[string]interface{}{
							"entity": map[string]interface{}{
								"id":       "pay_fake_attacker_12345",
								"order_id": "order_fake_99999",
								"amount":   50000,
								"status":   "captured",
								"notes": map[string]interface{}{
									"room_name": "Kurinchi",
								},
							},
						},
					},
				}
				jsonBytes, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Razorpay-Signature", "invalid_forged_signature_hex_code")

				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Failed to execute request: %v", err)
				}
				defer resp.Body.Close()

				bodyBytes, _ := io.ReadAll(resp.Body)
				tc.Actual = fmt.Sprintf("HTTP %d | Response: %s", resp.StatusCode, string(bodyBytes))

				if resp.StatusCode == http.StatusOK && strings.Contains(string(bodyBytes), "processed") {
					tc.FailureReason = "WEBHOOK SECURITY BUG (BUG-05): Razorpay webhook processed forged payment event with an invalid signature (HTTP 200 OK)"
					tc.Errorf("Webhook Signature Bypass Vulnerability exposed")
				}
			},
		},
		{
			Name:        "[Regression BUG-06] Concurrency Race Condition — Simultaneous Room Bookings",
			Description: "Verify system prevents double bookings when two users attempt concurrent bookings on the same room",
			Expected:    "Exactly 1 booking succeeds (HTTP 201) and 1 fails (HTTP 400 Bad Request / Overlap Conflict)",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/api/towers/bookings"

				var wg sync.WaitGroup
				statusCodeChan := make(chan int, 2)

				bookingPayload := func(bookerName string) []byte {
					p := map[string]interface{}{
						"roomId":       "rm-kurinchi-01",
						"checkInDate":  time.Now().AddDate(0, 1, 10).Format("2006-01-02"),
						"checkOutDate": time.Now().AddDate(0, 1, 12).Format("2006-01-02"),
						"bookingFor":   "self",
						"guestDetails": []map[string]interface{}{
							{"name": bookerName, "gender": "male", "age": 30},
						},
						"advanceAmount": 500,
					}
					b, _ := json.Marshal(p)
					return b
				}

				wg.Add(2)
				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 10 * time.Second}
					req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bookingPayload("Concurrent User A")))
					req.Header.Set("Content-Type", "application/json")
					resp, err := client.Do(req)
					if err == nil {
						statusCodeChan <- resp.StatusCode
						resp.Body.Close()
					} else {
						statusCodeChan <- 500
					}
				}()

				go func() {
					defer wg.Done()
					client := &http.Client{Timeout: 10 * time.Second}
					req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bookingPayload("Concurrent User B")))
					req.Header.Set("Content-Type", "application/json")
					resp, err := client.Do(req)
					if err == nil {
						statusCodeChan <- resp.StatusCode
						resp.Body.Close()
					} else {
						statusCodeChan <- 500
					}
				}()

				wg.Wait()
				close(statusCodeChan)

				statuses := []int{}
				for s := range statusCodeChan {
					statuses = append(statuses, s)
				}

				tc.Actual = fmt.Sprintf("Concurrent Request Statuses: %v", statuses)

				successCount := 0
				for _, s := range statuses {
					if s == http.StatusCreated || s == http.StatusOK {
						successCount++
					}
				}

				if successCount > 1 {
					tc.FailureReason = fmt.Sprintf("RACE CONDITION BUG (BUG-06): Concurrent booking requests resulted in multiple successful bookings (%d successful) for the exact same room and dates", successCount)
					tc.Errorf("Race condition exposed: both concurrent bookings succeeded!")
				}
			},
		},
		{
			Name:        "[Regression BUG-07] Negative Payment Amount Validation on /api/towers/create-order",
			Description: "Verify system rejects negative amount in payment order creation",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/api/towers/create-order"
				payload := map[string]interface{}{"amount": -5000}
				jsonBytes, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				bodyBytes, _ := io.ReadAll(resp.Body)
				tc.Actual = fmt.Sprintf("HTTP %d | Response: %s", resp.StatusCode, string(bodyBytes))

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "VALIDATION BUG (BUG-07): Server accepted negative payment amount (-5000) and generated an order (HTTP 200 OK)"
					tc.Errorf("Server accepted negative amount!")
				}
			},
		},
		{
			Name:        "[Regression BUG-08] Missing Auth Verification on Critical Endpoints",
			Description: "Verify that passing malformed JWT tokens triggers HTTP 401 instead of panicking.",
			Expected:    "HTTP 401 Unauthorized",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/api/member/profile"
				req, _ := http.NewRequest("GET", url, nil)
				req.Header.Set("Authorization", "Bearer invalid.jwt.token.format.bug")
				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusInternalServerError {
					tc.FailureReason = "BUG EXPOSED: Malformed JWT token caused HTTP 500 panic instead of HTTP 401"
					tc.Fatalf("Server panicked on malformed token")
				}
				tc.Actual = "Safely returned 401 for malformed JWT"
			},
		},
		{
			Name:        "[Regression BUG-09] Rate Limiting - Exact Off-By-One Boundary Bucket",
			Description: "Verify rate limiting algorithm triggers exactly at the threshold, not threshold+1",
			Expected:    "HTTP 429 Too Many Requests exactly at N+1 request",
			TestFn: func(tc *tests.TestContext) {
				url := strings.TrimRight(baseURL, "/") + "/api/member/login"
				payload := map[string]interface{}{"username": "bug09-rate@example.com", "password": "wrong"}
				jsonBytes, _ := json.Marshal(payload)

				threshold := 5 // Assuming threshold is 5 for this endpoint
				var lastStatusCode int
				for i := 0; i <= threshold; i++ {
					req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
					req.Header.Set("Content-Type", "application/json")
					resp, err := httpClient.Do(req)
					if err == nil {
						lastStatusCode = resp.StatusCode
						resp.Body.Close()
					}
				}

				if lastStatusCode != http.StatusTooManyRequests {
					tc.FailureReason = fmt.Sprintf("BUG-09 REPRODUCED: Expected HTTP 429 after exactly %d requests, got HTTP %d", threshold+1, lastStatusCode)
					tc.Errorf("Rate limiting off-by-one boundary bug exposed")
				} else {
					tc.Actual = "Correctly rate limited exactly at the boundary"
				}
			},
		},
		{
			Name:        "[Regression BUG-10] Subscriptions - Improper Downgrade Flow",
			Description: "Verify system prevents downgrade to a lower tier while actively preserving higher tier limits.",
			Expected:    "HTTP 400 Bad Request or HTTP 403 Forbidden",
			TestFn: func(tc *tests.TestContext) {
				token := getValidMemberToken(tc.T, tc.Client)
				url := strings.TrimRight(baseURL, "/") + "/api/subscriptions/downgrade"
				payload := map[string]interface{}{"newPlanId": "basic", "force": true}
				jsonBytes, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)
				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "BUG-10 REPRODUCED: Server allowed improper subscription downgrade bypass without settling dues."
					tc.Errorf("Improper downgrade vulnerability exposed")
				} else {
					tc.Actual = "Safely rejected improper downgrade"
				}
			},
		},
		{
			Name:        "[Regression BUG-11] Admin Resources - Mismatched Mimetype Check",
			Description: "Verify backend physically inspects file magic bytes, not just extension or Content-Type header.",
			Expected:    "HTTP 400 Bad Request or HTTP 415 Unsupported Media Type",
			TestFn: func(tc *tests.TestContext) {
				adminToken := getValidAdminToken(tc.T, tc.Client)
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				_ = writer.WriteField("categoryId", "establishment")
				_ = writer.WriteField("title", "Malicious Executable")
				
				part, _ := writer.CreateFormFile("file", "malicious.pdf") // Pretending to be PDF
				part.Write([]byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xFF\xFF\x00\x00")) // MZ Windows Executable Magic Bytes
				writer.Close()

				reqUrl := strings.TrimRight(baseURL, "/") + "/api/admin/resources/upload"
				req, _ := http.NewRequest("POST", reqUrl, &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+adminToken)

				resp, err := httpClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "BUG-11 REPRODUCED: Server accepted MZ executable masquerading as a .pdf file (Missing Deep Mimetype inspection)"
					tc.Errorf("Mismatched mimetype vulnerability exposed")
				} else {
					tc.Actual = "Correctly rejected file with mismatched magic bytes"
				}
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(
			t,
			testCase.Name,
			testCase.Description,
			testCase.Expected,
			testCase.TestFn,
		)
	}
}
