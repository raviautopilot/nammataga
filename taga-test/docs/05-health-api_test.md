# Detailed Documentation: Health API Test Suite (`05-health-api_test.go`)

## Suite Overview

- **Target Handler**: `taga-api/handler/health.go`
- **Endpoints Tested**:
  - `GET /health` (`HealthHandler`)
  - `GET /` (`RootHandler`)
- **Test File Location**: `taga-test/tests/api/05-health-api_test.go`
- **Total Test Scenarios**: 40 Executions

---

## Complete Test Case Documentation

### 1. `Health API - GET /health - Valid Request`
- **Test Name**: `TestAPI_Health_ValidRequest`
- **Endpoint**: `GET /health`
- **Why Created**: Validates that the primary application health check endpoint returns HTTP `200 OK` with JSON body `{"status": "healthy"}` so Kubernetes, AWS ALB, and load balancers can verify pod liveness.
- **Input**: None (`GET /health`)
- **Expected Result**: `HTTP 200 OK with JSON payload {"status": "healthy"}`
- **Actual Result (Got)**: `HTTP 200 OK with status = 'healthy'`
- **Bug Prevented**: Prevents unmonitored container failures or load balancer traffic blackholing.

---

### 2. `Root API - GET / - Valid Request`
- **Test Name**: `TestAPI_Root_ValidRequest`
- **Endpoint**: `GET /`
- **Why Created**: Verifies that the root landing route (`/`) returns a valid welcome payload with `status: "success"` and non-empty `message`.
- **Input**: None (`GET /`)
- **Expected Result**: `HTTP 200 OK with JSON payload {"message": "Hello from Gin + Zap!", "status": "success"}`
- **Actual Result (Got)**: `HTTP 200 OK with status = 'success', message = 'Hello from Gin + Zap!'`
- **Bug Prevented**: Prevents broken root endpoint returns or 500 errors when pinging base domain.

---

### 3. `Health API - GET /health - Valid Headers`
- **Test Name**: `TestAPI_Health_ValidHeaders`
- **Endpoint**: `GET /health`
- **Why Created**: Confirms that non-browser clients (monitoring scripts, cURL, Prometheus) sending `Accept: application/json` and `User-Agent: HealthMonitor/2.0` succeed.
- **Input**: `Headers: { "Accept": "application/json", "User-Agent": "HealthMonitor/2.0" }`
- **Expected Result**: `HTTP 200 OK with status 'healthy'`
- **Actual Result (Got)**: `HTTP 200 OK, status = 'healthy'`
- **Bug Prevented**: Prevents HTTP 406 or header rejection on automated health probes.

---

### 4. `Health API - GET /health - SLA Response Time`
- **Test Name**: `TestAPI_Health_ResponseTime`
- **Endpoint**: `GET /health`
- **Why Created**: Enforces strict latency SLA (< 200ms) for high-frequency health probes.
- **Input**: None (`GET /health`)
- **Expected Result**: `HTTP 200 OK within < 200ms latency SLA`
- **Actual Result (Got)**: `HTTP 200 OK in 14ms`
- **Bug Prevented**: Detects thread blocking or IO delays during health monitoring.

---

### 5. `Root API - GET / - SLA Response Time`
- **Test Name**: `TestAPI_Root_ResponseTime`
- **Endpoint**: `GET /`
- **Why Created**: Enforces latency SLA (< 200ms) for base root domain.
- **Input**: None (`GET /`)
- **Expected Result**: `HTTP 200 OK within < 200ms latency SLA`
- **Actual Result (Got)**: `HTTP 200 OK in 15ms`
- **Bug Prevented**: Detects slow root response rendering.

---

### 6. `TableDriven - GET Root Endpoint` & `TableDriven - GET Health Check Endpoint`
- **Test Name**: `TestAPI_Health_AllEndpoints_TableDriven`
- **Endpoints**: `GET /`, `GET /health`
- **Why Created**: Parameterized execution across all health handlers to verify non-empty JSON body returns.
- **Input**: Table of endpoints `[ "/", "/health" ]`
- **Expected Result**: `HTTP 200 OK with non-empty JSON response`
- **Actual Result (Got)**: `HTTP 200 OK, JSON body length > 0 bytes`
- **Bug Prevented**: Catches zero-byte body responses.

---

### 7. `MethodValidation - POST /`, `PUT /`, `DELETE /`, `PATCH /` (8 Sub-tests)
- **Test Name**: `TestAPI_Health_WrongHTTPMethods`
- **Endpoints**: `GET /`, `GET /health`
- **Why Created**: Ensures read-only health routes reject state-changing HTTP methods (`POST`, `PUT`, `DELETE`, `PATCH`).
- **Input**: `POST /health`, `PUT /health`, `DELETE /health`, `PATCH /health`, etc.
- **Expected Result**: `HTTP 405 Method Not Allowed or 404 Not Found`
- **Actual Result (Got)**: `Rejected with HTTP status 404 / 405`
- **Bug Prevented**: Prevents accidental data modification or method tampering vulnerabilities.

---

### 8. `RouteValidation - GET /health/status`, `/health/check`, etc. (5 Sub-tests)
- **Test Name**: `TestAPI_Health_InvalidRoutes`
- **Endpoints**: Non-existent paths (`/health/status`, `/health/123`, etc.)
- **Why Created**: Validates route boundaries and Gin router isolation.
- **Input**: `GET /health/status`, `GET /health/check`, `GET /api/health`
- **Expected Result**: `HTTP 404 Not Found`
- **Actual Result (Got)**: `Returned HTTP status 404`
- **Bug Prevented**: Prevents wildcard route leakage or unintended handler executions.

---

### 9. `HeaderValidation - Missing Content-Type`, `Invalid Content-Type`, `Dummy Bearer Token`, etc. (5 Sub-tests)
- **Test Name**: `TestAPI_Health_HeaderValidations`
- **Endpoints**: `GET /health`
- **Why Created**: Ensures public health routes safely handle missing headers, text/xml Content-Types, and unexpected Authorization Bearer tokens.
- **Input**: `Content-Type: text/xml`, `Authorization: Bearer fake_token`
- **Expected Result**: `HTTP 200 OK (Public endpoint safely handles non-standard headers)`
- **Actual Result (Got)**: `HTTP 200 OK, status = 'healthy'`
- **Bug Prevented**: Prevents auth middleware panics on unauthenticated health probes.

---

### 10. `SecurityInjection - SQLi, XSS, HTML, Unicode, Overflow` (16 Sub-tests)
- **Test Name**: `TestAPI_Health_SecurityInjections`
- **Endpoints**: `GET /`, `GET /health`
- **Why Created**: Tests resilience against malicious query parameters (`?id=1' OR '1'='1`, `?q=<script>alert(1)</script>`, UTF-8 unicode `?q=ஆனா`, 2KB string overflow).
- **Input**: Various security injection strings in URL query string
- **Expected Result**: `HTTP 200 OK or 400 Bad Request (Must NOT trigger 500 Internal Server Error)`
- **Actual Result (Got)**: `HTTP 200 OK (Query parameters safely ignored without 500 error)`
- **Bug Prevented**: Prevents server crashes, memory leaks, or log injection panics.

---

### 11. `UnexpectedPayload - GET /health with Valid/Empty/Large JSON Body` (3 Sub-tests)
- **Test Name**: `TestAPI_Health_UnexpectedPayloads`
- **Endpoints**: `GET /health`
- **Why Created**: Verifies that sending unexpected JSON body in a GET request does not break HTTP server.
- **Input**: `GET /health` with JSON body payload
- **Expected Result**: `HTTP 200 OK (GET handler ignores body without crashing)`
- **Actual Result (Got)**: `HTTP 200 OK, status = 'healthy'`
- **Bug Prevented**: Prevents request body unmarshaling panics on GET requests.
