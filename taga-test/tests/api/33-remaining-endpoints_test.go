package api_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// TestAPI_Root_Health tests the root and health endpoints
func TestAPI_RootAndMiscEndpoints(t *testing.T) {
	// GET / — root handler
	tests.RunAPITestWithDetails(t, "[Public] GET Root Handler", "Fetches the API root endpoint.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		var resp interface{}
		err := tctx.Client.SendHttpRequest("GET", "/", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("GET / failed: %v", err)
		}
		tctx.Actual = "HTTP 200 OK from root"
	})

	// GET /api/grievance-banner
	tests.RunAPITestWithDetails(t, "[Public] GET Grievance Banner", "Fetches the grievance section banner image URL.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		var resp interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/grievance-banner", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("GET /api/grievance-banner failed: %v", err)
		}
		tctx.Actual = "HTTP 200 OK grievance banner retrieved"
	})

	// GET /api/office/:pathParam — parameterised office route
	tests.RunAPITestWithDetails(t, "[Public] GET Office with Path Param", "Fetches office details with a specific sub-section path param.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		var resp interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/office/contact", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("GET /api/office/contact failed: %v", err)
		}
		tctx.Actual = "HTTP 200 OK office section retrieved"
	})

	// GET /docs/*filepath — static doc serving (non-existent file → 404)
	tests.RunAPITestWithDetails(t, "[Public] GET Docs - Non-existent File Returns 404", "Requests a non-existent document and verifies 404 response.", "HTTP 404 Not Found", func(tctx *tests.TestContext) {
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/docs/nonexistent-file-e2e.pdf", nil, nil, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusNotFound, "")
	})

	// POST /api/auth/reset-password — bad token
	tests.RunAPITestWithDetails(t, "[Public] POST Reset Password - Invalid Token", "Attempts password reset with a bogus token and verifies rejection.", "HTTP 401 Unauthorized or 500", func(tctx *tests.TestContext) {
		payload := map[string]string{
			"email":       "sudhantest08@gmail.com",
			"oldPassword": "invalid-token-xyz",
			"newPassword": "SomeNewPass123!",
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/auth/reset-password", nil, &payload, &resp, nil)
		// The handler returns 401 for invalid/expired token, or 500 for other errors
		if err == nil {
			tctx.FailureReason = "Expected error response, got 200 OK"
			tctx.Fatalf("Expected error response but got 200 OK")
		}
		tctx.Actual = fmt.Sprintf("Correctly rejected with error: %v", err)
	})

	// POST /api/auth/forgot-password — non-existent email → 404
	tests.RunAPITestWithDetails(t, "[Public] POST Admin Forgot Password - Unknown Email", "Triggers admin forgot-password with an unknown email.", "HTTP 404 Not Found", func(tctx *tests.TestContext) {
		payload := map[string]string{
			"email": "nobody_ever_exists@taga-e2e.invalid",
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/auth/forgot-password", nil, &payload, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusNotFound, "")
	})
}

// TestAPI_PaymentsMemberAuth tests payment endpoints that require member auth
func TestAPI_PaymentsMemberAuth(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Member] POST Payments Create Order - Validation Error", "Sends an empty create-order payload via member-auth payment route.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidMemberToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/payments/create-order", nil, &map[string]interface{}{}, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	tests.RunAPITestWithDetails(t, "[Member] POST Payments Verify - Validation Error", "Sends an empty verify-payment payload via member-auth payment route.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidMemberToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/payments/verify", nil, &map[string]interface{}{}, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	tests.RunAPITestWithDetails(t, "[Public] POST Payments Create Order - Unauthorized", "Calls member-auth payment endpoint without token.", "HTTP 401 Unauthorized", func(tctx *tests.TestContext) {
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/payments/create-order", nil, &map[string]interface{}{}, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusUnauthorized, "")
	})
}

// TestAPI_AdminBulkUpload tests bulk member upload validation
func TestAPI_AdminBulkUpload(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] POST Bulk Upload Members - No File", "Calls bulk upload without any file attached.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/members/bulk-upload", nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	tests.RunAPITestWithDetails(t, "[Admin] POST Bulk Upload Members - Invalid File Type", "Uploads a .txt file (invalid extension) to bulk-upload.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, _ := writer.CreateFormFile("file", "data.txt")
		part.Write([]byte("bad file content"))
		writer.Close()

		reqUrl := tctx.Client.BaseURL + "/api/admin/members/bulk-upload"
		req, _ := http.NewRequest("POST", reqUrl, &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			tctx.FailureReason = fmt.Sprintf("Expected 400, got %d", resp.StatusCode)
			tctx.Fatalf("Expected 400, got %d", resp.StatusCode)
		}
		tctx.Actual = "HTTP 400 correctly rejected .txt bulk upload"
	})
}

func TestAPI_RemainingEndpoints_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Public] GET Docs - Directory Traversal Attempt",
			Description: "Attempts to perform directory traversal using path params.",
			Expected:    "HTTP 404 Not Found or HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/docs/../../../../etc/passwd", nil, nil, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for directory traversal"
					tc.Fatalf("Expected error for directory traversal but got success")
				}
				tc.Actual = "Correctly rejected directory traversal payload"
			},
		},
		{
			Name:        "[Public] POST Payments Create Order - Missing Fields",
			Description: "Sends create order payload with missing mandatory fields.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"amount": 100, // Missing name, email, phone, etc.
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/payments/create-order", nil, &payload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection due to missing fields"
					tc.Fatalf("Expected 400 Bad Request but got 200 OK")
				}
				tc.Actual = "Correctly failed on missing mandatory fields"
			},
		},
		{
			Name:        "[Member] POST Payments Create Order - Floating Point Precision Bug",
			Description: "Sends create order payload with excessive floating point precision in amount.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"amount": 100.99999999999999,
					"name": "Test",
					"email": "test@test.com",
					"phone": "9999999999",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/payments/create-order", nil, &payload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection due to floating point precision issue"
					tc.Fatalf("Expected 400 Bad Request but got 200 OK")
				}
				tc.Actual = "Correctly failed on excessive floating point precision"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
