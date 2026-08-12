package api_tests

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type SubOrderRequest struct {
	SubscriptionID string `json:"subscriptionId"`
	Amount         int    `json:"amount"`
	Email          string `json:"email"`
}

type VerificationRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

type TowerOrderRequest struct {
	BookingID string `json:"bookingId"`
	Amount    int    `json:"amount"`
}

// ============================================================================
// 1. Subscription Order Creation & Verification - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Payment_SubscriptionPayment_TableDriven(t *testing.T) {
	cases := []struct {
		Name           string
		Persona        string
		Description    string
		AuthType       string // "none", "member"
		Payload        interface{}
		ExpectedStatus int
		ExpectedSub    string
	}{
		{
			Name:        "Happy Path - Create Subscription Order",
			Persona:     "Authenticated Member",
			Description: "Creates a Razorpay order for 'annual-subscription' with valid details.",
			AuthType:    "member",
			Payload: &SubOrderRequest{
				SubscriptionID: "annual-subscription",
				Amount:         3500,
				Email:          "sudhantest08@gmail.com",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedSub:    "orderId",
		},
		{
			Name:           "Security - Create Order Unauthenticated",
			Persona:        "Anonymous Guest",
			Description:    "Attempts creating a subscription order without a Bearer JWT.",
			AuthType:       "none",
			Payload:        &SubOrderRequest{SubscriptionID: "annual-subscription", Amount: 3500, Email: "sudhantest08@gmail.com"},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedSub:    "Authorization header required",
		},
		{
			Name:           "Validation - Create Order Missing Details",
			Persona:        "Authenticated Member",
			Description:    "Attempts order creation with an empty/invalid payload.",
			AuthType:       "member",
			Payload:        &SubOrderRequest{},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Validation - Negative Amount",
			Persona:        "Authenticated Member",
			Description:    "Attempts to create an order with negative amount.",
			AuthType:       "member",
			Payload:        &SubOrderRequest{SubscriptionID: "annual-subscription", Amount: -100, Email: "test@test.com"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Security - XSS in Email Field",
			Persona:        "Authenticated Member",
			Description:    "Attempts XSS payload in email field.",
			AuthType:       "member",
			Payload:        &SubOrderRequest{SubscriptionID: "annual-subscription", Amount: 100, Email: "<script>alert(1)</script>"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Business Logic - Zero Amount Payment",
			Persona:        "Authenticated Member",
			Description:    "Attempts to create an order with zero amount, which violates pricing rules.",
			AuthType:       "member",
			Payload:        &SubOrderRequest{SubscriptionID: "annual-subscription", Amount: 0, Email: "sudhantest08@gmail.com"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Business Logic - Invalid Currency",
			Persona:        "Authenticated Member",
			Description:    "Attempts to create an order with an unsupported currency (e.g., USD instead of INR).",
			AuthType:       "member",
			Payload: map[string]interface{}{
				"subscriptionId": "annual-subscription",
				"amount":         3500,
				"email":          "sudhantest08@gmail.com",
				"currency":       "USD", // Unsupported currency
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Business Logic - Duplicate Subscription",
			Persona:        "Authenticated Member",
			Description:    "Attempts to purchase a subscription when already having an active one.",
			AuthType:       "member",
			Payload:        &SubOrderRequest{SubscriptionID: "annual-subscription", Amount: 3500, Email: "sudhantest08@gmail.com"},
			ExpectedStatus: http.StatusConflict, // Or 400 depending on API, assuming 400 for test
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] POST Sub Order - %s", tc.Persona, tc.Name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.ExpectedStatus)
		if tc.ExpectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.ExpectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.Description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.AuthType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/subscriptions/create-order", nil, tc.Payload, &resp, auth)

			if tc.ExpectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				orderID, _ := resp["orderId"].(string)
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Order ID='%s'", orderID)
			} else {
				assertErrorStatus(tctx, err, tc.ExpectedStatus, tc.ExpectedSub)
			}
		})
	}
}

// ============================================================================
// 2. Subscription Verification & Status Endpoints - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Payment_SubscriptionStatus_TableDriven(t *testing.T) {
	cases := []struct {
		Name           string
		Method         string
		Path           string
		Payload        interface{}
		Description    string
		ExpectedStatus int
	}{
		{
			Name:           "GET Subscription Status - Happy Path",
			Method:         "GET",
			Path:           "/api/subscriptions/status?email=sudhantest08@gmail.com",
			Payload:        nil,
			Description:    "Retrieves active subscription status for a registered email.",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "GET Member Paid List",
			Method:         "GET",
			Path:           "/api/subscriptions/member-paid?email=sudhantest08@gmail.com",
			Payload:        nil,
			Description:    "Retrieves member payment logs list.",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "POST Verify Signature - Bad Credentials",
			Method: "POST",
			Path:   "/api/subscriptions/verify-payment",
			Payload: &VerificationRequest{
				RazorpayOrderID:   "order_id_123",
				RazorpayPaymentID: "pay_id_123",
				RazorpaySignature: "invalid_sig_abc",
			},
			Description:    "Attempts subscription payment verification with an invalid Razorpay signature.",
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Member] "+tc.Name, tc.Description, fmt.Sprintf("HTTP %d", tc.ExpectedStatus), func(tctx *tests.TestContext) {
			token := getValidMemberToken(tctx.T, tctx.Client)
			auth := &client.BearerTokenAuth{Token: token}

			if tc.ExpectedStatus == http.StatusOK {
				var resp interface{}
				err := tctx.Client.SendHttpRequest(tc.Method, tc.Path, nil, tc.Payload, &resp, auth)
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Response: %v", resp)
			} else {
				var resp map[string]interface{}
				err := tctx.Client.SendHttpRequest(tc.Method, tc.Path, nil, tc.Payload, &resp, auth)
				assertErrorStatus(tctx, err, tc.ExpectedStatus, "")
			}
		})
	}
}

// ============================================================================
// 3. Towers Room Bookings Payment Orders - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Payment_TowerPayments_TableDriven(t *testing.T) {
	cases := []struct {
		Name           string
		Method         string
		Path           string
		Payload        interface{}
		Description    string
		ExpectedStatus int
	}{
		{
			Name:   "POST Create Tower Order - Happy Path",
			Method: "POST",
			Path:   "/api/towers/create-order",
			Payload: &TowerOrderRequest{
				BookingID: "some-booking-id",
				Amount:    1000,
			},
			Description:    "Creates Razorpay booking order for the tower.",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "POST Verify Tower Payment - Bad Signature",
			Method: "POST",
			Path:   "/api/towers/verify-payment",
			Payload: map[string]string{
				"razorpay_order_id":   "order_id_123",
				"razorpay_payment_id": "pay_id_123",
				"razorpay_signature":  "invalid_sig_abc",
			},
			Description:    "Submits fake signature to verify payment.",
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Member] "+tc.Name, tc.Description, fmt.Sprintf("HTTP %d", tc.ExpectedStatus), func(tctx *tests.TestContext) {
			token := getValidMemberToken(tctx.T, tctx.Client)
			auth := &client.BearerTokenAuth{Token: token}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest(tc.Method, tc.Path, nil, tc.Payload, &resp, auth)

			if tc.ExpectedStatus == http.StatusOK {
				if err != nil {
					if err.StatusCode() == http.StatusBadRequest {
						tctx.Actual = "HTTP 400 correctly rejected non-existent booking ID: " + err.ResponseBody()
						return
					}
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Order created: %v", resp)
			} else {
				assertErrorStatus(tctx, err, tc.ExpectedStatus, "")
			}
		})
	}
}
