package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/tests"
)

// WebhookEntity represents the payment details inside the webhook payload
type WebhookEntity struct {
	ID       string `json:"id"`
	OrderID  string `json:"order_id"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Email    string `json:"email"`
}

type WebhookInnerPayload struct {
	Payment struct {
		Entity WebhookEntity `json:"entity"`
	} `json:"payment"`
}

type WebhookRequestPayload struct {
	Event   string              `json:"event"`
	Payload WebhookInnerPayload `json:"payload"`
}

// ============================================================================
// 1. Static Images & Docs serving
// ============================================================================

func TestAPI_Static_Assets(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Static Image Logo", "Retrieves the main application Logo asset.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		err := tctx.Client.SendHttpRequest("GET", "/api/images/Logo.jpg", nil, nil, nil, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK static image served successfully"
	})

	tests.RunAPITestWithDetails(t, "[Public] GET Static Image Member Banner", "Retrieves member banner static image.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		err := tctx.Client.SendHttpRequest("GET", "/api/images/member-banner.jpg", nil, nil, nil, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK static member banner served successfully"
	})

	tests.RunAPITestWithDetails(t, "[Public] GET Non-existent Document File", "Attempts fetching a non-existent document file path.", "HTTP 404 Not Found", func(tctx *tests.TestContext) {
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/docs/nonexistent-doc.pdf", nil, nil, &resp, nil)

		assertErrorStatus(tctx, err, http.StatusNotFound, "File not found")
	})
}

// ============================================================================
// 2. Swagger Documentation Routes
// ============================================================================

func TestAPI_Static_Swagger(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Swagger API Docs Index", "Retrieves index page for API swagger documentation.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		err := tctx.Client.SendHttpRequest("GET", "/swagger/index.html", nil, nil, nil, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK swagger index page loaded successfully"
	})
}

// ============================================================================
// 3. Razorpay Webhooks
// ============================================================================

func TestAPI_Static_Webhook(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] POST Webhook - Ignored Event type", "Sends a non-payment.captured event to the webhook.", "HTTP 200 OK containing Ignored event message", func(tctx *tests.TestContext) {
		var payload WebhookRequestPayload
		payload.Event = "payment.failed"

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/webhook/razorpay", nil, &payload, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		msg, _ := resp["message"].(string)
		if msg != "Ignored event" {
			tctx.FailureReason = fmt.Sprintf("Expected 'Ignored event', got: %s", msg)
			tctx.Errorf("Unexpected message: %s", msg)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
	})

	tests.RunAPITestWithDetails(t, "[Public] POST Webhook - Bad Signature", "Sends a captured payment event but with an invalid verification signature header.", "HTTP 401 Unauthorized", func(tctx *tests.TestContext) {
		var payload WebhookRequestPayload
		payload.Event = "payment.captured"
		payload.Payload.Payment.Entity = WebhookEntity{
			ID:       "pay_999",
			OrderID:  "order_999",
			Amount:   3500,
			Currency: "INR",
			Status:   "captured",
			Email:    "sudhantest08@gmail.com",
		}

		headers := map[string]string{
			"X-Razorpay-Signature": "invalid_sig_abc_123",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/webhook/razorpay", headers, &payload, &resp, nil)

		assertErrorStatus(tctx, err, http.StatusUnauthorized, "Invalid signature")
	})
}

func TestAPI_NegativeScenarios_WebStatic(t *testing.T) {
	type TestCaseType struct {
		Name           string
		Persona        string
		Description    string
		Method         string
		Endpoint       string
		AuthType       string
		Payload        interface{}
		Headers        map[string]string
		ExpectedStatus int
		ExpectedSub    string
	}

	testCases := []TestCaseType{
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Anonymous",
			Description:    "POST to a GET only static file route",
			Method:         "POST",
			Endpoint:       "/api/images/Logo.jpg",
			AuthType:       "none",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "Missing Required Fields - Webhook",
			Persona:        "Anonymous",
			Description:    "Send webhook without body",
			Method:         "POST",
			Endpoint:       "/api/webhook/razorpay",
			AuthType:       "none",
			Payload:        nil,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Malformed JSON Payload",
			Persona:        "Anonymous",
			Description:    "Invalid JSON to webhook",
			Method:         "POST",
			Endpoint:       "/api/webhook/razorpay",
			AuthType:       "none",
			Payload:        "broken{json",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "XSS in static path",
			Persona:        "Anonymous",
			Description:    "Path traversal or XSS attempt",
			Method:         "GET",
			Endpoint:       "/api/images/<script>alert(1)</script>",
			AuthType:       "none",
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:           "SQLi in path",
			Persona:        "Anonymous",
			Description:    "SQLi attempt in static route",
			Method:         "GET",
			Endpoint:       "/api/images/' OR '1'='1",
			AuthType:       "none",
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:           "State Machine Violation - Refunded Payment Captured",
			Persona:        "System",
			Description:    "Webhook for capturing an already refunded payment",
			Method:         "POST",
			Endpoint:       "/api/webhook/razorpay",
			AuthType:       "none",
			Headers:        map[string]string{"X-Razorpay-Signature": "valid_sig_mock"},
			Payload:        map[string]interface{}{"event": "payment.captured", "payload": map[string]interface{}{"payment": map[string]interface{}{"entity": map[string]interface{}{"status": "refunded", "amount": 500}}}},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Extreme Boundary - Max Int Amount",
			Persona:        "System",
			Description:    "Webhook with max integer amount",
			Method:         "POST",
			Endpoint:       "/api/webhook/razorpay",
			AuthType:       "none",
			Headers:        map[string]string{"X-Razorpay-Signature": "valid_sig_mock"},
			Payload:        map[string]interface{}{"event": "payment.captured", "payload": map[string]interface{}{"payment": map[string]interface{}{"entity": map[string]interface{}{"status": "captured", "amount": 9223372036854775807}}}},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Logical Paradox - Missing Order ID",
			Persona:        "System",
			Description:    "Captured payment without an order ID",
			Method:         "POST",
			Endpoint:       "/api/webhook/razorpay",
			AuthType:       "none",
			Headers:        map[string]string{"X-Razorpay-Signature": "valid_sig_mock"},
			Payload:        map[string]interface{}{"event": "payment.captured", "payload": map[string]interface{}{"payment": map[string]interface{}{"entity": map[string]interface{}{"status": "captured", "order_id": "", "amount": 3500}}}},
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s %s - %s", tc.Persona, tc.Method, tc.Endpoint, tc.Name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.ExpectedStatus)
		if tc.ExpectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.ExpectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.Description, expectedStr, func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest(tc.Method, tc.Endpoint, tc.Headers, tc.Payload, &resp, nil)
			if err == nil {
				tctx.Fatalf("Expected error for negative scenario, got none. Response: %v", resp)
			}
		})
	}
}
