package api_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type SubRecord struct {
	ID               string    `json:"id"`
	MemberID         string    `json:"member_id"`
	MemberEmail      string    `json:"member_email"`
	MemberName       string    `json:"member_name"`
	SubscriptionID   string    `json:"subscription_id"`
	SubscriptionName string    `json:"subscription_name"`
	Amount           int       `json:"amount"`
	OrderID          string    `json:"order_id"`
	PaymentID        string    `json:"payment_id"`
	Status           string    `json:"status"` // pending, active, expired, cancelled
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	LastPaidDate     time.Time `json:"last_paid_date"`
	NextDueDate      time.Time `json:"next_due_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func TestAPI_SubscriptionStatus_E2E(t *testing.T) {
	const email = "sudhantest08@gmail.com"
	dbPath := filepath.Join("..", "..", "..", "taga-api", "data", "subscriptions", "member_subscriptions.json")

	// Create directory if not exists
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Read existing subscriptions
	var existingSubs []SubRecord
	data, err := os.ReadFile(dbPath)
	if err == nil {
		_ = json.Unmarshal(data, &existingSubs)
	}

	// Create test subscription record
	testRecord := SubRecord{
		ID:             "sub-test-e2e-id-123",
		MemberEmail:    email,
		SubscriptionID: "annual-subscription",
		Status:         "active",
		StartDate:      time.Now(),
		EndDate:        time.Now().Add(365 * 24 * time.Hour),
		NextDueDate:    time.Now().Add(365 * 24 * time.Hour),
		UpdatedAt:      time.Now(),
	}

	// Append and write to file
	updatedSubs := append(existingSubs, testRecord)
	out, err := json.MarshalIndent(updatedSubs, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal subscriptions: %v", err)
	}
	if err := os.WriteFile(dbPath, out, 0644); err != nil {
		t.Fatalf("Failed to write test subscription: %v", err)
	}

	// Make sure we clean up at the end of the test
	t.Cleanup(func() {
		// Restore original file
		var currentSubs []SubRecord
		cdata, cerr := os.ReadFile(dbPath)
		if cerr == nil {
			_ = json.Unmarshal(cdata, &currentSubs)
			var filtered []SubRecord
			for _, sub := range currentSubs {
				if sub.ID != "sub-test-e2e-id-123" {
					filtered = append(filtered, sub)
				}
			}
			fout, _ := json.MarshalIndent(filtered, "", "  ")
			_ = os.WriteFile(dbPath, fout, 0644)
		}
	})

	// Run status query test
	tests.RunAPITestWithDetails(t, "[Member] GET Subscription Status - E2E Active Check", "Verifies status check endpoint returns active state for paid member.", "HTTP 200 OK containing status active", func(tctx *tests.TestContext) {
		token := getValidMemberToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/subscriptions/status?email="+email, nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}

		status, _ := resp["status"].(string)
		if status != "active" {
			tctx.FailureReason = fmt.Sprintf("Expected status 'active', got '%s'", status)
			tctx.Fatalf("Expected status 'active', got '%s'", status)
		}
		tctx.Actual = "HTTP 200 OK, verified active status successfully"
	})
}

func TestAPI_NegativeScenarios_SubscriptionStatus(t *testing.T) {
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
			Name:           "Missing Auth - Error Expected",
			Persona:        "Anonymous",
			Description:    "Access endpoint without token",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=test@test.com",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Invalid Token - Error Expected",
			Persona:        "Attacker",
			Description:    "Access with fake token",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=test@test.com",
			AuthType:       "invalid",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Member",
			Description:    "Use wrong method for endpoint",
			Method:         "POST",
			Endpoint:       "/api/subscriptions/status?email=test@test.com",
			AuthType:       "member",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Member",
			Description:    "SQLi attempt in URL email",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=' OR '1'='1",
			AuthType:       "member",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "XSS Payload in Param - Error Expected",
			Persona:        "Member",
			Description:    "XSS script injection in endpoint",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=<script>alert(1)</script>",
			AuthType:       "member",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Missing Required Query Param - Error Expected",
			Persona:        "Member",
			Description:    "Missing email parameter",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status",
			AuthType:       "member",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "State Machine Violation - Cancelled But Active",
			Persona:        "Member",
			Description:    "Status check for a logically impossible state (not directly triggerable, but simulating via params)",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=cancelled@active.com&status=cancelled",
			AuthType:       "member",
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:           "Extreme Boundary - Max Length Email",
			Persona:        "Member",
			Description:    "Email param exceeding max string limit",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=extremely-long-email-address-which-exceeds-standard-varchar-limits-by-a-lot-and-should-be-rejected-by-the-system-immediately-upon-parsing-due-to-database-constraints@test.com",
			AuthType:       "member",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Role Context Switching - Admin Checking Member Status w/o Member Token",
			Persona:        "Admin",
			Description:    "Admin uses admin token on member route",
			Method:         "GET",
			Endpoint:       "/api/subscriptions/status?email=test@test.com",
			AuthType:       "admin",
			ExpectedStatus: http.StatusForbidden,
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
			var auth client.Authenticator
			switch tc.AuthType {
			case "admin":
				token := getValidAdminToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "member":
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "invalid":
				auth = &client.BearerTokenAuth{Token: "invalid-jwt-token-12345"}
			case "none":
				auth = nil
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest(tc.Method, tc.Endpoint, tc.Headers, tc.Payload, &resp, auth)
			if err == nil {
				tctx.Fatalf("Expected error for negative scenario, got none. Response: %v", resp)
			}
		})
	}
}
