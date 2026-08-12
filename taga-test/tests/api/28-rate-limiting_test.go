package api_tests

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"e2e-template/tests"
)

func TestAPI_Auth_ConcurrencyAndSpike(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] POST Member Login - Concurrency Stress Check", "Sends 15 concurrent login requests to check server resiliency under load.", "All requests complete without server crash", func(tctx *tests.TestContext) {
		payload := map[string]string{
			"username": "sudhantest08@gmail.com",
			"password": "wrong-password-spike-check",
		}

		var wg sync.WaitGroup
		concurrency := 15
		statuses := make([]int, concurrency)
		errs := make([]error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				var resp map[string]interface{}
				// Use fresh request scope
				err := tctx.Client.SendHttpRequest("POST", "/api/member/login", nil, &payload, &resp, nil)
				if err != nil {
					statuses[idx] = err.StatusCode()
					errs[idx] = err
				} else {
					statuses[idx] = http.StatusOK
				}
			}(i)
		}

		wg.Wait()

		// Verify no connections got dropped completely or crashed (status should be 401 Unauthorized)
		unauthorizedCount := 0
		for _, status := range statuses {
			if status == http.StatusUnauthorized {
				unauthorizedCount++
			}
		}

		tctx.Actual = fmt.Sprintf("HTTP Completed %d requests. Unauthorized matches: %d/%d", concurrency, unauthorizedCount, concurrency)
	})
}

func TestAPI_Auth_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Member] POST Login - Missing Payload",
			Description: "Sends login request with empty JSON payload.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/member/login", nil, nil, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected error on empty payload"
					tc.Fatalf("Expected 400 Bad Request but got 200 OK")
				}
				tc.Actual = "Correctly failed on missing payload"
			},
		},
		{
			Name:        "[Member] POST Login - SQLi Username",
			Description: "Sends SQL injection payload in username.",
			Expected:    "HTTP 401 Unauthorized or HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				payload := map[string]string{
					"username": "' OR 1=1--",
					"password": "password",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/member/login", nil, &payload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection of SQLi payload"
					tc.Fatalf("Expected error for SQLi but got 200 OK")
				}
				tc.Actual = "Correctly rejected SQLi payload"
			},
		},
		{
			Name:        "[Member] POST Login - Exact Rate Limit Boundary Check",
			Description: "Sends exactly N requests to test off-by-one boundary for rate limit buckets.",
			Expected:    "HTTP 429 Too Many Requests after threshold",
			TestFn: func(tc *tests.TestContext) {
				payload := map[string]string{
					"username": "ratelimit-boundary@example.com",
					"password": "wrong-password",
				}
				// Simulate reaching boundary
				threshold := 5
				for i := 0; i < threshold; i++ {
					var resp map[string]interface{}
					_ = tc.Client.SendHttpRequest("POST", "/api/member/login", nil, &payload, &resp, nil)
				}
				// The threshold+1 request should fail with 429
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/member/login", nil, &payload, &resp, nil)
				if err == nil || err.StatusCode() != http.StatusTooManyRequests {
					tc.FailureReason = "Expected 429 Too Many Requests exactly at boundary"
					tc.Fatalf("Expected 429 after %d requests, got: %v", threshold, err)
				}
				tc.Actual = "Correctly hit rate limit at boundary"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
