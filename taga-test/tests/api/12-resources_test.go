package api_test

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

// ============================================================================
// 5. POST/DELETE /api/admin/resources/external-links - Admin CRUD Tests
// ============================================================================

func TestAPI_Admin_ExternalLinks_CRUD(t *testing.T) {
	testTitle := "E2E Test Portal Link"
	testURL := "https://e2e-test-portal.gov.in"

	// 1. Security: Attempt adding link without token -> 401
	tests.RunAPITestWithDetails(t, "[Anonymous] POST External Link Without Auth", "Attempts adding external link without auth header.", "HTTP 401 Unauthorized", func(tctx *tests.TestContext) {
		payload := map[string]string{
			"title": testTitle,
			"url":   testURL,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/resources/external-links", nil, &payload, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusUnauthorized, "Authorization header required")
	})

	// 2. Admin adds new link -> 200 OK
	tests.RunAPITestWithDetails(t, "[Admin] POST Add External Link", "Admin adds a new external link directly into CSV file.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		adminToken := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: adminToken}

		payload := map[string]string{
			"title": testTitle,
			"url":   testURL,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/resources/external-links", nil, &payload, &resp, auth)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK, External link added successfully"
	})

	// 3. Admin attempts adding duplicate link -> 409 Conflict
	tests.RunAPITestWithDetails(t, "[Admin] POST Add Duplicate External Link", "Admin attempts adding link with existing title.", "HTTP 409 Conflict", func(tctx *tests.TestContext) {
		adminToken := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: adminToken}

		payload := map[string]string{
			"title": testTitle,
			"url":   testURL,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/resources/external-links", nil, &payload, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusConflict, "already exists")
	})

	// 4. Member verifies link is in GET /api/resources/external-links
	tests.RunAPITestWithDetails(t, "[Member] GET External Links Contains New Link", "Member fetches external links and verifies newly added link.", "HTTP 200 OK with new link", func(tctx *tests.TestContext) {
		memberToken := getValidMemberToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: memberToken}

		var links []map[string]string
		err := tctx.Client.SendHttpRequest("GET", "/api/resources/external-links", nil, nil, &links, auth)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}

		found := false
		for _, l := range links {
			if strings.EqualFold(l["title"], testTitle) {
				found = true
				break
			}
		}
		if !found {
			tctx.FailureReason = fmt.Sprintf("New link '%s' not found in external links list", testTitle)
			tctx.Fatalf("New link '%s' not found", testTitle)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, found link '%s'", testTitle)
	})

	// 5. Admin deletes link -> 200 OK
	tests.RunAPITestWithDetails(t, "[Admin] DELETE External Link", "Admin deletes the external link from the CSV file.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		adminToken := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: adminToken}

		path := fmt.Sprintf("/api/admin/resources/external-links?title=%s", strings.ReplaceAll(testTitle, " ", "%20"))
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("DELETE", path, nil, nil, &resp, auth)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK, External link deleted successfully"
	})

	// 6. Admin deletes already deleted link -> 404 Not Found
	tests.RunAPITestWithDetails(t, "[Admin] DELETE Already Deleted External Link", "Admin attempts deleting non-existent link.", "HTTP 404 Not Found", func(tctx *tests.TestContext) {
		adminToken := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: adminToken}

		path := fmt.Sprintf("/api/admin/resources/external-links?title=%s", strings.ReplaceAll(testTitle, " ", "%20"))
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("DELETE", path, nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusNotFound, "not found")
	})
}
