# E2E API Test Suite Execution Guide

This document outlines how to compile, execute, configure, and run individual API automation tests. It includes copy-paste ready commands for all individual API test suites.

---

## 1. Prerequisites & Environment Setup

Ensure that Go (1.20+) is installed and that the target backend API service is reachable.

### Check Go Installation:
```bash
go version
```

### Precompile the API Test Binary (Optional, Recommended for Fast Execution):
Building a precompiled test binary avoids recompiling test files on every invocation:
```bash
go test -c ./tests/api -o api.test
```

---

## 2. Test Configuration

The API test suite reads its configuration from [config.json](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/config.json). You can modify this file directly or override parameters via environment variables:

| config.json Field | Environment Variable | Default Value | Description |
|---|---|---|---|
| `baseUrl` | `E2E_BASE_URL` | `https://api.nammataga.com` | Target URL of the backend REST API |
| `timeout` | `E2E_TIMEOUT` | `10` | Request timeout in seconds |
| `logLevel` | `E2E_LOG_LEVEL` | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`) |

---

## 3. How to Execution Options

### Option A: Using the `run-api-tests.sh` Script (Recommended)
Executes tests using the standard test runner shell wrapper:
```bash
# Run all API tests
./run-api-tests.sh

# Run a specific API test
./run-api-tests.sh -run TestAPI_Health_ValidRequest
```

### Option B: Using the Precompiled Binary (`api.test`)
For maximum speed in local development and CI/CD pipelines:

> [!IMPORTANT]
> When running the precompiled binary directly, prefix `go test` flags with `test.` (e.g. `-test.v`, `-test.run`).

```bash
# Execute a specific test with the precompiled binary
./api.test -test.v -test.run TestAPI_Health_ValidRequest
```

### Option C: Using Standard `go test`
```bash
go test -v ./tests/api/... -run TestAPI_Health_ValidRequest
```

---

## 4. Individual Test Execution Commands (Copy-Paste Ready)

### 🏥 Health & System API Tests ([05-health-api_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/api/05-health-api_test.go))

#### 1. Health Status Endpoint (`/health`)
Verifies HTTP 200 OK and valid health status JSON payload.
```bash
./run-api-tests.sh -run TestAPI_Health_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_ValidRequest
```

#### 2. Root Endpoint (`/`)
Verifies root endpoint response and welcome message.
```bash
./run-api-tests.sh -run TestAPI_Root_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Root_ValidRequest
```

#### 3. Health Endpoint Headers Validation
Validates Server, CORS, and Content-Type headers on `/health`.
```bash
./run-api-tests.sh -run TestAPI_Health_ValidHeaders
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_ValidHeaders
```

#### 4. Health Endpoint Response Time Benchmark
Ensures `/health` API responds within target SLA threshold (< 500ms).
```bash
./run-api-tests.sh -run TestAPI_Health_ResponseTime
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_ResponseTime
```

#### 5. Root Endpoint Response Time Benchmark
Ensures `/` API responds within target SLA threshold (< 500ms).
```bash
./run-api-tests.sh -run TestAPI_Root_ResponseTime
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Root_ResponseTime
```

#### 6. Table-Driven Health Endpoints Sweep
Runs table-driven test covering all health & root endpoints in sequence.
```bash
./run-api-tests.sh -run TestAPI_Health_AllEndpoints_TableDriven
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_AllEndpoints_TableDriven
```

#### 7. Wrong HTTP Methods Restriction
Enforces HTTP method restrictions (POST/PUT/DELETE on GET-only health routes).
```bash
./run-api-tests.sh -run TestAPI_Health_WrongHTTPMethods
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_WrongHTTPMethods
```

#### 8. Invalid Health Routes (404 Handling)
Verifies 404 response on non-existent health paths.
```bash
./run-api-tests.sh -run TestAPI_Health_InvalidRoutes
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_InvalidRoutes
```

#### 9. Health Request Header Validations
Validates API behavior with custom, missing, or malformed request headers.
```bash
./run-api-tests.sh -run TestAPI_Health_HeaderValidations
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_HeaderValidations
```

#### 10. Health Security Injection Resistance
Verifies resistance against SQLi and XSS injection strings in query parameters.
```bash
./run-api-tests.sh -run TestAPI_Health_SecurityInjections
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_SecurityInjections
```

#### 11. Health Unexpected Payloads
Tests body payload handling on GET health endpoints.
```bash
./run-api-tests.sh -run TestAPI_Health_UnexpectedPayloads
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_Health_UnexpectedPayloads
```

---

### ℹ️ Public Info & About API Tests ([04-about-api_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/api/04-about-api_test.go))

#### 12. About Public Info Endpoint (`/api/public/about`)
Validates HTTP 200 and schema response for `/api/public/about`.
```bash
./run-api-tests.sh -run TestAPI_About_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_ValidRequest
```

#### 13. About Endpoint Headers
Validates headers and CORS settings on the About API.
```bash
./run-api-tests.sh -run TestAPI_About_ValidHeaders
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_ValidHeaders
```

#### 14. About Endpoint Response Time Benchmark
Validates latency SLA for `/api/public/about`.
```bash
./run-api-tests.sh -run TestAPI_About_ResponseTime
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_ResponseTime
```

#### 15. About Stats Endpoint (`/api/public/about/stats`)
Validates organization statistics endpoint.
```bash
./run-api-tests.sh -run TestAPI_AboutStats_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_AboutStats_ValidRequest
```

#### 16. About Objectives Endpoint (`/api/public/about/objectives`)
Validates organizational objectives endpoint.
```bash
./run-api-tests.sh -run TestAPI_AboutObjectives_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_AboutObjectives_ValidRequest
```

#### 17. About Services Endpoint (`/api/public/about/services`)
Validates services provided endpoint.
```bash
./run-api-tests.sh -run TestAPI_AboutServices_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_AboutServices_ValidRequest
```

#### 18. About Contact Endpoint (`/api/public/about/contact`)
Validates public contact information endpoint.
```bash
./run-api-tests.sh -run TestAPI_AboutContact_ValidRequest
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_AboutContact_ValidRequest
```

#### 19. Table-Driven Sweep of All About Endpoints
Runs comprehensive table-driven tests over all About sub-routes.
```bash
./run-api-tests.sh -run TestAPI_About_AllEndpoints_TableDriven
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_AllEndpoints_TableDriven
```

#### 20. About Wrong HTTP Methods Restriction
Ensures non-GET HTTP methods return method not allowed or route error.
```bash
./run-api-tests.sh -run TestAPI_About_WrongHTTPMethods
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_WrongHTTPMethods
```

#### 21. About Invalid Routes (404 Handling)
Verifies 404 responses for bad `/api/public/about/*` subpaths.
```bash
./run-api-tests.sh -run TestAPI_About_InvalidRoutes
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_InvalidRoutes
```

#### 22. About Header Validations
Validates Content-Type and custom header behavior for About APIs.
```bash
./run-api-tests.sh -run TestAPI_About_HeaderValidations
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_HeaderValidations
```

#### 23. About Security Injection Resistance
Tests SQLi and XSS injection payloads against About API parameters.
```bash
./run-api-tests.sh -run TestAPI_About_SecurityInjections
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_SecurityInjections
```

#### 24. About Unexpected Body Payloads
Tests API resiliency against malformed or oversized body payloads on GET routes.
```bash
./run-api-tests.sh -run TestAPI_About_UnexpectedPayloads
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_About_UnexpectedPayloads
```

---

### 🔒 Endpoint Security Audit ([06-endpoint-security_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/api/06-endpoint-security_test.go))

#### 25. Comprehensive Endpoint Security Suite
Audits all 70 Swagger-defined endpoints for public access vs protected JWT enforcement, header validation, and CORS security.
```bash
./run-api-tests.sh -run TestAPI_06_EndpointSecurity
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_06_EndpointSecurity
```

##### Filter Specific Subtests within Endpoint Security Suite:
Because `TestAPI_06_EndpointSecurity` runs individual endpoint subtests, you can filter by method or path using regex:

*Run only GET endpoint security checks:*
```bash
./run-api-tests.sh -run 'TestAPI_06_EndpointSecurity/.*GET'
```

*Run only Admin route security checks:*
```bash
./run-api-tests.sh -run 'TestAPI_06_EndpointSecurity/.*admin'
```

*Run security check for a single endpoint (e.g. `/api/public/about`):*
```bash
./run-api-tests.sh -run 'TestAPI_06_EndpointSecurity/.*about'
```
*Using Precompiled Binary:*
```bash
./api.test -test.v -test.run 'TestAPI_06_EndpointSecurity/.*about'
```

---

### 📝 Posts & Pointer Validation Tests ([01-get-post_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/api/01-get-post_test.go), [02-create-post_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/api/02-create-post_test.go), [03-pointer-validation_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/api/03-pointer-validation_test.go))

#### 26. Get Post Test
Validates basic GET request and response extraction.
```bash
./run-api-tests.sh -run TestAPI_01_GetPost
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_01_GetPost
```

#### 27. Create Post Test
Validates POST request payload submission and response parsing.
```bash
./run-api-tests.sh -run TestAPI_02_CreatePost
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_02_CreatePost
```

#### 28. Pointer Validation Framework Test
Validates pointer-passing architecture and result aggregation helper functions.
```bash
./run-api-tests.sh -run TestAPI_03_PointerValidation
```
*Precompiled Binary:*
```bash
./api.test -test.v -test.run TestAPI_03_PointerValidation
```

---

## 5. API Test Suite Architecture

The API test suite utilizes a **pointer-based, failure-skipping design pattern** in `tests/helpers.go`.

### Example Test Implementation Pattern:
```go
func TestAPI_Health_ValidRequest(t *testing.T) {
	tests.RunAPITest(t, "Verify Application Health Status", func(t *testing.T, api *tests.APIPersona) {
		result := tests.NewResult("Health Status Test")

		// Sequential actions using pointer passing
		resp := api.Get(result, "/health")
		api.AssertStatusCode(result, resp, 200)
		api.AssertJSONField(result, resp, "status", "ok")

		if result.Failed() {
			t.Fatalf("Test Failed: %v. Advice: %v", result.Error, result.Advice)
		}
	})
}
```

---

## 6. Test Reports & Artifacts

After every test execution, artifacts and reports are saved to `evidence/run-YYYY-MM-DD_HH-MM-SS/`:

* **HTML Dashboard**: Open `evidence/run-.../reports/report.html` in a web browser for visual test metrics, execution times, and HTTP status summaries.
* **JSON Report**: Machine-readable output stored in `evidence/run-.../reports/report.json`.
* **Markdown Report**: Summary written to `evidence/test-report.md`.
