package api_tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// ============================================================================
// 1. GET /api/resources & /api/resources/all - Parameterized Tests
// ============================================================================

func TestAPI_Resources_List(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		endpoint       string
		authType       string // "none", "member"
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Get Resource Categories",
			persona:        "Authenticated Member",
			description:    "Retrieves names of resource categories with member JWT.",
			endpoint:       "/api/resources",
			authType:       "member",
			expectedStatus: http.StatusOK,
			expectedSub:    "id",
		},
		{
			name:           "Happy Path - Get All Resources with Documents",
			persona:        "Authenticated Member",
			description:    "Retrieves full categories and document hierarchies with member JWT.",
			endpoint:       "/api/resources/all",
			authType:       "member",
			expectedStatus: http.StatusOK,
			expectedSub:    "documents",
		},
		{
			name:           "Security - Get Categories Unauthenticated",
			persona:        "Anonymous Guest",
			description:    "Attempts fetching categories without Bearer Authorization header.",
			endpoint:       "/api/resources",
			authType:       "none",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET %s - %s", tc.persona, tc.endpoint, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp []interface{}
			err := tctx.Client.SendHttpRequest("GET", tc.endpoint, nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d elements", len(resp))
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 2. GET /api/resources/:id - Parameterized Tests
// ============================================================================

func TestAPI_Resources_GetByCategory(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		categoryID     string
		queryParams    string
		authType       string // "none", "member"
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Retrieve Valid Category Documents",
			persona:        "Authenticated Member",
			description:    "Retrieves documents for the 'establishment' category with valid credentials.",
			categoryID:     "establishment",
			queryParams:    "",
			authType:       "member",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Happy Path - Filter Category by Subcategory",
			persona:        "Authenticated Member",
			description:    "Retrieves filtered documents using query param subcategory=State.",
			categoryID:     "establishment",
			queryParams:    "?subcategory=State",
			authType:       "member",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Status - Nonexistent Category ID",
			persona:        "Authenticated Member",
			description:    "Attempts fetching documents for a bogus/nonexistent category ID.",
			categoryID:     "invalid-category-name",
			queryParams:    "",
			authType:       "member",
			expectedStatus: http.StatusNotFound,
			expectedSub:    "Category not found",
		},
		{
			name:           "Security - Get Category Unauthenticated",
			persona:        "Anonymous Guest",
			description:    "Attempts document retrieval without authorization header.",
			categoryID:     "establishment",
			queryParams:    "",
			authType:       "none",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
		{
			name:           "Validation - Special Characters in Category",
			persona:        "Fuzzer",
			description:    "Attempts fetching documents for a category with special characters.",
			categoryID:     "est@blish!ment",
			queryParams:    "",
			authType:       "member",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - SQL Injection in Subcategory",
			persona:        "Malicious User",
			description:    "Attempts SQL injection via subcategory query param.",
			categoryID:     "establishment",
			queryParams:    "?subcategory=' OR 1=1--",
			authType:       "member",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Business Logic - Restricted Category Access",
			persona:        "Authenticated Member",
			description:    "Attempts fetching documents for an admin-only category.",
			categoryID:     "admin-only-docs",
			queryParams:    "",
			authType:       "member",
			expectedStatus: http.StatusNotFound, // or Forbidden, treating as not found for safety
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET Category %s - %s", tc.persona, tc.categoryID, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp []interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/resources/"+tc.categoryID+tc.queryParams, nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d documents", len(resp))
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 3. GET /api/resources/external-links - Parameterized Tests
// ============================================================================

func TestAPI_Resources_ExternalLinks(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "member"
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Retrieve CSV Links List",
			persona:        "Authenticated Member",
			description:    "Retrieves the parsed external links list with member JWT.",
			authType:       "member",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Security - Get Links Unauthenticated",
			persona:        "Anonymous Guest",
			description:    "Attempts fetching external links list without Bearer header.",
			authType:       "none",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET External Links - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp []interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/resources/external-links", nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d external links", len(resp))
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 4. GET /api/resources-banner - Parameterized Tests
// ============================================================================

func TestAPI_Resources_Banner(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Resources Banner", "Checks the resources banner image URL path.", "HTTP 200 OK with banner URL", func(tctx *tests.TestContext) {
		var resp map[string]string
		err := tctx.Client.SendHttpRequest("GET", "/api/resources-banner", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		image := resp["image"]
		if !strings.Contains(image, "banner") {
			tctx.FailureReason = fmt.Sprintf("Expected banner path, got: %s", image)
			tctx.Errorf("Expected banner path, got: %s", image)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Image Path='%s'", image)
	})
}
