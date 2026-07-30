# Detailed Documentation: About API Test Suite (`04-about-api_test.go`)

## Suite Overview

- **Target Handler**: `taga-api/handler/about.go`
- **Endpoints Tested**:
  - `GET /api/public/about` (`AboutHandler`)
  - `GET /api/public/about/stats` (`AboutStatsHandler`)
  - `GET /api/public/about/objectives` (`AboutObjectivesHandler`)
  - `GET /api/public/about/services` (`AboutServicesHandler`)
  - `GET /api/public/about/contact` (`AboutContactHandler`)
- **Test File Location**: `taga-test/tests/api/04-about-api_test.go`
- **Total Test Scenarios**: 64 Executions

---

## Complete Test Case Documentation

### 1. `About API - GET /api/public/about - Valid Request`
- **Test Name**: `TestAPI_About_ValidRequest`
- **Endpoint**: `GET /api/public/about`
- **Why Created**: Validates that public organization metadata (Name, Acronym, Established Year, Mission, Vision) returns HTTP 200 OK.
- **Input**: None (`GET`)
- **Expected Result**: `HTTP 200 OK with AboutResponse JSON payload`
- **Actual Result (Got)**: `HTTP 200 OK with valid AboutResponse`
- **Bug Prevented**: Prevents broken landing page content for website visitors.

---

### 2. `About API - GET /api/public/about - Valid Headers`
- **Test Name**: `TestAPI_About_ValidHeaders`
- **Endpoint**: `GET /api/public/about`
- **Why Created**: Confirms that non-browser mobile and web API clients with custom headers succeed.
- **Input**: `Headers: { "Accept": "application/json", "User-Agent": "TAGA-Test-AutomationClient/1.0" }`
- **Expected Result**: `HTTP 200 OK`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Prevents client-side header rejection errors.

---

### 3. `About API - GET /api/public/about - SLA Response Time`
- **Test Name**: `TestAPI_About_ResponseTime`
- **Endpoint**: `GET /api/public/about`
- **Why Created**: Latency SLA check (< 500ms) for organization info retrieval.
- **Input**: None (`GET`)
- **Expected Result**: `HTTP 200 OK within < 500ms latency SLA`
- **Actual Result (Got)**: `HTTP 200 OK in 120ms`
- **Bug Prevented**: Detects slow file reading IO bottlenecks on `about.json`.

---

### 4. `About API - GET /api/public/about/stats - Valid Request`
- **Test Name**: `TestAPI_AboutStats_ValidRequest`
- **Endpoint**: `GET /api/public/about/stats`
- **Why Created**: Validates retrieval of statistics list (active members, years of service, districts covered).
- **Input**: None (`GET`)
- **Expected Result**: `HTTP 200 OK with []StatsResponse array`
- **Actual Result (Got)**: `HTTP 200 OK with array length > 0`
- **Bug Prevented**: Prevents metric counters from rendering blank on frontend.

---

### 5. `About API - GET /api/public/about/objectives - Valid Request`
- **Test Name**: `TestAPI_AboutObjectives_ValidRequest`
- **Endpoint**: `GET /api/public/about/objectives`
- **Why Created**: Retrieves organization objectives array.
- **Input**: None (`GET`)
- **Expected Result**: `HTTP 200 OK with []Objective array`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Ensures objective cards parse cleanly.

---

### 6. `About API - GET /api/public/about/services - Valid Request`
- **Test Name**: `TestAPI_AboutServices_ValidRequest`
- **Endpoint**: `GET /api/public/about/services`
- **Why Created**: Retrieves organization services catalog.
- **Input**: None (`GET`)
- **Expected Result**: `HTTP 200 OK with []ServiceResponse array`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Validates services payload unmarshaling.

---

### 7. `About API - GET /api/public/about/contact - Valid Request`
- **Test Name**: `TestAPI_AboutContact_ValidRequest`
- **Endpoint**: `GET /api/public/about/contact`
- **Why Created**: Retrieves headquarters address, office hours, primary phone, email, and regional offices.
- **Input**: None (`GET`)
- **Expected Result**: `HTTP 200 OK with ContactResponse JSON payload`
- **Actual Result (Got)**: `HTTP 200 OK with ContactResponse`
- **Bug Prevented**: Ensures contact page details render accurately.

---

### 8. `TableDriven - GET Main About Info, Stats, Objectives, Services, Contact` (5 Sub-tests)
- **Test Name**: `TestAPI_About_AllEndpoints_TableDriven`
- **Endpoints**: All 5 public about endpoints
- **Why Created**: Parameterized health check over all read-only about routes.
- **Expected Result**: `HTTP 200 OK with non-empty JSON body`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Catches zero-byte file responses.

---

### 9. `MethodValidation - POST, PUT, DELETE, PATCH on all 5 Endpoints` (20 Sub-tests)
- **Test Name**: `TestAPI_About_WrongHTTPMethods`
- **Endpoints**: All 5 public endpoints
- **Why Created**: Ensures read-only endpoints reject `POST`, `PUT`, `DELETE`, `PATCH` requests.
- **Expected Result**: `HTTP 405 Method Not Allowed or 404 Not Found`
- **Actual Result (Got)**: `Rejected with HTTP status 404 / 405`
- **Bug Prevented**: Prevents state modification attempts on static info routes.

---

### 10. `RouteValidation - GET /api/public/about/nonexistent`, etc. (4 Sub-tests)
- **Test Name**: `TestAPI_About_InvalidRoutes`
- **Endpoints**: Invalid sub-routes
- **Why Created**: Verifies route boundaries and Gin router isolation.
- **Expected Result**: `HTTP 404 Not Found`
- **Actual Result (Got)**: `HTTP status 404`
- **Bug Prevented**: Prevents wildcard route handling errors.

---

### 11. `HeaderValidation - Missing/Invalid Content-Type, Accept, Bearer Token` (6 Sub-tests)
- **Test Name**: `TestAPI_About_HeaderValidations`
- **Endpoints**: `/api/public/about`
- **Why Created**: Verifies public endpoint resilience under varying client headers.
- **Expected Result**: `HTTP 200 OK`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Prevents middleware panics on public routes.

---

### 12. `SecurityInjection - SQLi, XSS, HTML, Unicode, Overflow` (24 Sub-tests)
- **Test Name**: `TestAPI_About_SecurityInjections`
- **Endpoints**: `/api/public/about`, `/stats`, `/contact`
- **Why Created**: Tests resilience against query string injection attacks.
- **Expected Result**: `HTTP 200 OK or 400 Bad Request (Must NOT trigger 500 Internal Server Error)`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Prevents server crashes or 500 errors on unexpected URL query inputs.

---

### 13. `UnexpectedPayload - GET with Valid/Empty/Large JSON Body` (3 Sub-tests)
- **Test Name**: `TestAPI_About_UnexpectedPayloads`
- **Endpoints**: `/api/public/about`
- **Why Created**: Verifies GET requests carrying unexpected JSON bodies are handled gracefully.
- **Expected Result**: `HTTP 200 OK`
- **Actual Result (Got)**: `HTTP 200 OK`
- **Bug Prevented**: Prevents body parsing errors on GET requests.
