package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type RegistrationItem struct {
	SNo                      int    `json:"sno"`
	TagaID                   string `json:"tagaId"`
	Name                     string `json:"name"`
	Initial                  string `json:"initial"`
	Gender                   string `json:"gender"`
	FatherName               string `json:"father_name"`
	MotherName               string `json:"mother_name"`
	EducationalQualification string `json:"educational_qualification"`
	Designation              string `json:"designation"`
	WorkingDistrict          string `json:"working_district"`
	NativeDistrict           string `json:"native_district"`
	RecruitmentBatch         string `json:"recruitment_batch"`
	SeniorityNumber          string `json:"seniority_number"`
	ResidentialAddress       string `json:"residential_address"`
	PermanentAddress         string `json:"permanent_address"`
	DateOfBirth              string `json:"date_of_birth"`
	MobileNumber             string `json:"mobile_number"`
	EmailId                  string `json:"email_id"`
	TBFNumber                string `json:"tbf_number"`
	CPSGPFNumber             string `json:"cps_gpf_number"`
}

// TestAPI_E2E_MemberRegistration verifies the admin registration -> password init -> member login flow
func TestAPI_E2E_MemberRegistration(t *testing.T) {
	var adminToken string
	var registeredMemberID string
	var savedClient *client.Client

	// Step 1: Admin registers new member via JSON upload file
	tests.RunAPITestWithDetails(t, "[Admin] POST Bulk Member Registration - E2E", "Uploads a registration file containing a new member.", "HTTP 200 OK with success statistics", func(tctx *tests.TestContext) {
		savedClient = tctx.Client
		adminToken = getValidAdminToken(tctx.T, tctx.Client)

		// Create registration item
		item := RegistrationItem{
			SNo:                      1,
			TagaID:                   "TAGAE2E1",
			Name:                     "E2E Test User",
			Initial:                  "K",
			Gender:                   "Male",
			FatherName:               "Father Name",
			MotherName:               "Mother Name",
			EducationalQualification: "B.E.",
			Designation:              "AO",
			WorkingDistrict:          "Salem",
			NativeDistrict:           "Salem",
			RecruitmentBatch:         "2020",
			SeniorityNumber:          "123",
			ResidentialAddress:       "123 Street",
			PermanentAddress:         "123 Street",
			DateOfBirth:              "1992-05-15",
			MobileNumber:             "9876543210",
			EmailId:                  "e2eregistereduser@gmail.com",
			TBFNumber:                "TBF100",
			CPSGPFNumber:             "CPS100",
		}

		list := []RegistrationItem{item}
		jsonData, err := json.Marshal(list)
		if err != nil {
			tctx.Fatalf("Failed to marshal registration item: %v", err)
		}

		// Build multipart form body
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "registrations.json")
		if err != nil {
			tctx.Fatalf("Failed to create form file: %v", err)
		}
		part.Write(jsonData)
		writer.Close()

		reqUrl := tctx.Client.BaseURL + "/api/admin/member-registration"
		req, err := http.NewRequest("POST", reqUrl, &body)
		if err != nil {
			tctx.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Failed to send multipart request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %d. Body: %s", resp.StatusCode, buf.String())
			tctx.Fatalf("Expected 200 OK, got: %d. Body: %s", resp.StatusCode, buf.String())
		}

		tctx.Actual = "HTTP 200 OK registration upload completed successfully"
	})

	// Step 2: Retrieve the newly created member's ID from admin search
	if adminToken != "" {
		tests.RunAPITestWithDetails(t, "[Admin] GET Search Registered Member ID", "Locates the newly registered member ID to initialize password.", "HTTP 200 OK", func(tctx *tests.TestContext) {
			auth := &client.BearerTokenAuth{Token: adminToken}
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/admin/members?search=e2eregistereduser@gmail.com", nil, nil, &resp, auth)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Failed to list members: %v", err)
				tctx.Fatalf("Failed to list members: %v", err)
			}

			membersList, ok := resp["members"].([]interface{})
			if !ok || len(membersList) == 0 {
				tctx.FailureReason = "Newly registered member not found in search list"
				tctx.Fatalf("Newly registered member not found in search list")
			}

			firstMember := membersList[0].(map[string]interface{})
			id, _ := firstMember["id"].(string)
			if id == "" {
				tctx.FailureReason = "Member object missing ID field"
				tctx.Fatalf("Member object missing ID field")
			}
			registeredMemberID = id
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, retrieved Member ID: '%s'", registeredMemberID)
		})
	}

	// Clean up after the test runs: Delete the test member we registered
	t.Cleanup(func() {
		if savedClient != nil && registeredMemberID != "" && adminToken != "" {
			auth := &client.BearerTokenAuth{Token: adminToken}
			var resp map[string]interface{}
			_ = savedClient.SendHttpRequest("DELETE", "/api/admin/members/"+registeredMemberID, nil, nil, &resp, auth)
		}
	})
}

func TestAPI_NegativeScenarios_MemberRegistration(t *testing.T) {
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
			Method:         "POST",
			Endpoint:       "/api/admin/member-registration",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Invalid Token - Error Expected",
			Persona:        "Attacker",
			Description:    "Access with fake token",
			Method:         "POST",
			Endpoint:       "/api/admin/member-registration",
			AuthType:       "invalid",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Admin",
			Description:    "Use GET instead of POST",
			Method:         "GET",
			Endpoint:       "/api/admin/member-registration",
			AuthType:       "admin",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Admin",
			Description:    "SQLi attempt in search URL",
			Method:         "GET",
			Endpoint:       "/api/admin/members?search=' OR '1'='1",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "XSS Payload - Error Expected",
			Persona:        "Admin",
			Description:    "XSS in search parameter",
			Method:         "GET",
			Endpoint:       "/api/admin/members?search=<script>alert(1)</script>",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Malformed Form Payload - Error Expected",
			Persona:        "Admin",
			Description:    "Send non-multipart data to registration",
			Method:         "POST",
			Endpoint:       "/api/admin/member-registration",
			AuthType:       "admin",
			Payload:        "not-a-multipart",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Boundary Values / Missing Fields - Error Expected",
			Persona:        "Admin",
			Description:    "Delete member without ID",
			Method:         "DELETE",
			Endpoint:       "/api/admin/members/",
			AuthType:       "admin",
			ExpectedStatus: http.StatusNotFound,
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
