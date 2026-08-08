# Implementation Plan — Industry-Level Audit Log System

This document defines the implementation plan for a modular, secure, persistent, and maintainable **Audit Log System** for the `nammataga` application.

The existing application uses:

- **Go** backend (`taga-api`)
- **React** frontend (`taga-web`)
- **JSON files** for application data
- **Docker / Docker Compose**
- **VPS** deployment
- Existing authentication and administrator access controls

## Important Architecture Constraint

**Do NOT introduce PostgreSQL, MongoDB, SQLite, Redis, or any other database for audit logs.**

Audit logs must remain **filesystem-based JSON files**.

The audit system must be implemented as a reusable service so that the storage implementation can be changed in the future without changing every business handler.

---

# 1. Goals

The system must:

1. Record important user and security actions.
2. Record who performed each action.
3. Record when the action happened.
4. Record the affected module/resource.
5. Record old and new values for modifications where appropriate.
6. Store logs as JSON files.
7. Separate logs by year, month, and authenticated user ID.
8. Persist logs on the VPS outside the temporary Docker filesystem.
9. Protect audit logs from unauthorized access.
10. Provide an internal administrator-only audit UI.
11. Support filtering, searching, and pagination.
12. Safely handle concurrent writes.
13. Never store passwords, tokens, secrets, or other sensitive credentials.
14. Automatically remove logs older than the configured retention period.
15. Default the retention period to **3 months**.
16. Keep the retention period configurable through environment variables.
17. Avoid unnecessary changes to the existing application.

---

# 2. Audit Log Storage

Audit logs must be stored using this structure:

```text
audit-logs/
└── YYYY/
    └── MM/
        ├── user_<user_id>.json
        ├── user_<user_id>.json
        └── ...
```

Example:

```text
audit-logs/
└── 2026/
    └── 08/
        ├── user_42.json
        ├── user_43.json
        └── user_57.json
```

## User Identification

Use the **actual authenticated user's existing ID**.

Do NOT use a shared ID such as:

```text
admin-001
```

for all administrators.

For example:

```text
user_1.json
user_2.json
user_15.json
```

An administrator should use their actual administrator/user ID.

For an operation performed without an authenticated user:

```text
user_anonymous.json
```

or an equivalent safe filename.

The exact naming should follow the existing application's user identity system.

---

# 3. Audit Record Structure

Each audit record should contain:

```go
type AuditRecord struct {
    AuditID      string      `json:"audit_id"`
    UserID       string      `json:"user_id"`
    Username     string      `json:"username"`
    Action       string      `json:"action"`
    Module       string      `json:"module"`
    ResourceType string      `json:"resource_type,omitempty"`
    ResourceID   string      `json:"resource_id,omitempty"`
    Description  string      `json:"description"`
    OldData      interface{} `json:"old_data,omitempty"`
    NewData      interface{} `json:"new_data,omitempty"`
    IPAddress    string      `json:"ip_address,omitempty"`
    UserAgent    string      `json:"user_agent,omitempty"`
    Timestamp    string      `json:"timestamp"`
}
```

Use UUID for `AuditID`.

Use RFC3339 timestamps.

Example:

```text
2026-08-08T16:30:22+05:30
```

Prefer UTC internally if the existing application follows UTC conventions.

---

# 4. Audit Actions

The system should audit meaningful business and security operations.

## Authentication

```text
LOGIN
LOGOUT
LOGIN_FAILED
PASSWORD_CHANGED
```

## User / Administrator Management

```text
USER_CREATED
USER_UPDATED
USER_DELETED
ROLE_CHANGED
PERMISSION_CHANGED
```

## General Data Operations

```text
CREATE
UPDATE
DELETE
```

## Application-Specific Business Actions

Inspect the existing application and identify important business operations that should be audited.

Examples may include:

```text
BOOKING_CREATED
BOOKING_UPDATED
BOOKING_CANCELLED
PAYMENT_CREATED
PAYMENT_CONFIRMED
```

Only use actions that actually exist in the application.

### Important

Do NOT blindly log every HTTP request.

Audit meaningful:

- business operations
- security operations
- permission changes
- important data modifications

---

# 5. Audit Service

Create a dedicated reusable audit package.

Recommended location:

```text
taga-api/service/audit/
```

Possible structure:

```text
service/
└── audit/
    ├── audit.go
    ├── service.go
    ├── storage.go
    ├── sanitizer.go
    └── cleanup.go
```

Follow the existing project architecture if it uses a different service structure.

---

# 6. Audit Service API

Business handlers should not directly manipulate audit files.

They should call a reusable service.

Conceptually:

```go
auditService.Log(...)
```

or:

```go
audit.LogAudit(...)
```

The exact API should follow the existing project's coding style.

Handlers should NOT contain code such as:

```go
os.OpenFile(...)
os.ReadFile(...)
os.WriteFile(...)
```

for audit files.

The audit service owns all audit storage logic.

---

# 7. Audit Write Behavior

Audit logging must NOT be fire-and-forget.

Do NOT use:

```go
go saveAuditRecord(record)
```

for critical audit records.

Reason:

```text
Business operation succeeds
        ↓
Audit goroutine starts
        ↓
Container crashes
        ↓
Audit record is lost
```

For important actions, the audit write must be explicitly handled.

The audit service should return an error:

```go
err := auditService.Log(...)
```

The calling code must decide how to handle the failure based on the operation.

At minimum:

- never silently ignore audit-write errors
- log failures through the existing application logger
- ensure critical audit events are not silently lost

Do not allow an audit failure to create inconsistent business data without an explicit strategy.

---

# 8. JSON File Format

For the current project, store each user's monthly audit records in a JSON array.

Example:

```json
[
  {
    "audit_id": "6d1a...",
    "user_id": "42",
    "username": "ravi",
    "action": "UPDATE",
    "module": "CUSTOMER",
    "resource_type": "customer",
    "resource_id": "125",
    "description": "Updated customer phone number",
    "old_data": {
      "phone": "9876543210"
    },
    "new_data": {
      "phone": "9876500000"
    },
    "ip_address": "192.168.1.20",
    "user_agent": "Mozilla/5.0",
    "timestamp": "2026-08-08T16:30:22+05:30"
  }
]
```

The audit storage implementation must remain isolated behind the audit service so that a different storage format can be introduced later without modifying business handlers.

---

# 9. Safe Concurrent File Writes

Multiple users or requests may write audit logs simultaneously.

The audit service must prevent corrupted JSON.

For the current single-process Docker deployment, use a mutex or equivalent safe file-write mechanism.

Important:

```text
Read file
    ↓
Parse JSON
    ↓
Append record
    ↓
Marshal JSON
    ↓
Write safely
```

The entire read-modify-write operation must be protected.

If possible, write to a temporary file and atomically replace the original file to reduce the chance of a partially written JSON file.

Example concept:

```text
user_42.json.tmp
       ↓
write complete JSON
       ↓
rename
       ↓
user_42.json
```

The implementation must not leave a partially written JSON file after a failed write.

---

# 10. Future Multi-Container Consideration

The current project uses a single Go application container.

A process-local mutex is sufficient for the current architecture.

However, do not design the service in a way that permanently assumes multiple containers can safely share the same JSON files.

Document this limitation.

If the application later runs multiple backend containers writing to the same filesystem, the storage mechanism must be upgraded to a cross-process locking/storage solution.

Do not introduce this complexity unless the current deployment requires it.

---

# 11. Sensitive Data Protection

Audit logs must NEVER contain:

```text
password
password_hash
pass
token
access_token
refresh_token
secret
api_key
apikey
authorization
jwt
otp
private_key
client_secret
```

The sanitization mechanism must be robust.

Do NOT use `filepath.Clean()` or filesystem path operations to determine whether a JSON field is sensitive.

Instead:

1. Convert field names to lowercase.
2. Compare against a controlled sensitive-field list/pattern.
3. Recursively sanitize nested objects.
4. Recursively sanitize arrays.
5. Sanitize data after converting structs to a generic JSON representation when necessary.
6. Ensure sensitive values are replaced before the data is written to disk.

Example:

```json
{
  "username": "ravi",
  "password": "[REDACTED]",
  "profile": {
    "phone": "9876543210",
    "api_key": "[REDACTED]"
  }
}
```

Never store actual secrets.

---

# 12. IP Address and User Agent

Where the request context is available, record:

```text
IP address
User agent
```

Use the existing Gin request context and existing proxy configuration.

Do not blindly trust forwarded IP headers unless the existing Nginx/proxy configuration is trusted and correctly configured.

Follow the application's existing `ClientIP()` / proxy configuration.

---

# 13. Docker Persistent Storage

Audit logs must survive:

```text
docker compose down
docker compose up
docker compose restart
container rebuild
container recreation
```

The audit directory must not exist only inside the container's writable layer.

Inside the container:

```text
/app/audit-logs
```

On the VPS:

```text
./taga-api/audit-logs
```

or the existing deployment's appropriate persistent directory.

Docker Compose should mount it:

```yaml
volumes:
  - ./taga-api/data:/app/data
  - ./taga-api/audit-logs:/app/audit-logs
```

Use the existing project structure if a different host path is already established.

Do not change unrelated volumes or ports.

---

# 14. Git Protection

Audit logs must never be committed to Git.

Update:

```text
taga-api/.gitignore
```

with:

```text
audit-logs/
```

Do not delete existing `.gitignore` entries.

Do not commit generated audit history.

---

# 15. Audit API

Create a protected backend API:

```http
GET /api/admin/audit
```

The API must use the existing authentication and administrator authorization system.

Do NOT rely on a secret URL for security.

The backend must reject unauthorized requests even if someone manually calls the API.

---

# 16. Audit API Filters

Support:

```text
user_id
year
month
action
module
search
date range
page
limit
```

Example:

```http
GET /api/admin/audit?user_id=42&year=2026&month=08&page=1&limit=50
```

Return structured JSON.

Example:

```json
{
  "data": [],
  "page": 1,
  "limit": 50,
  "total": 100
}
```

---

# 17. Efficient File Reading

Do NOT recursively scan and parse every audit file for every API request.

Use filters to identify the required files.

Example:

```text
user_id=42
year=2026
month=08
```

should directly target:

```text
audit-logs/2026/08/user_42.json
```

If only year/month are supplied:

```text
audit-logs/2026/08/
```

can be scanned.

Avoid loading unnecessary years/months/users.

If pagination requires reading multiple files, keep the implementation efficient and bounded.

---

# 18. Audit API Security

The audit API must:

- require authentication
- require administrator/audit permission
- never expose raw filesystem paths
- never expose secrets
- never allow arbitrary filesystem paths from query parameters
- validate user IDs
- validate year/month values
- prevent path traversal
- return safe error messages

Never construct a filesystem path directly from unchecked user input.

For example, do not allow:

```text
../../some-secret-file
```

to influence file access.

---

# 19. Internal Audit UI

Create:

```text
/internal/audit
```

in the React application.

The page should be hidden from normal navigation if appropriate.

However:

**Hidden URL is not a security mechanism.**

Both the frontend and backend must enforce administrator access.

---

# 20. React Audit UI Features

The page should contain:

### Filters

```text
User
Year
Month
Action
Module
Search
Date range
```

### Table

```text
Time
User
Action
Module
Resource
Description
```

### Detail View

Clicking an audit record should display:

```text
Audit ID
User
User ID
Action
Module
Resource Type
Resource ID
Description
IP Address
User Agent
Timestamp
Old Data
New Data
```

For UPDATE operations:

```text
OLD DATA
        ↓
NEW DATA
```

should be clearly distinguishable.

Use the existing UI components/design system.

---

# 21. React Route

Modify the existing routing architecture to support:

```text
/internal/audit
```

Do not introduce a second routing system if the project already has one.

If `App.tsx` currently uses a page-switching mechanism instead of React Router, follow the existing architecture.

Do not rewrite routing unnecessarily.

---

# 22. Authentication and Authorization

Before showing the audit page:

```text
User
 ↓
Authentication
 ↓
Administrator / Audit Permission
 ↓
Audit UI
```

The backend must perform the same authorization check:

```text
GET /api/admin/audit
 ↓
Authentication
 ↓
Authorization
 ↓
Audit data
```

Never rely only on frontend checks.

---

# 23. Three-Month Audit Retention

Audit logs should be automatically deleted when they become older than the configured retention period.

Default:

```env
AUDIT_LOG_RETENTION_MONTHS=3
```

Do not hard-code `3` throughout the code.

Make the retention period configurable.

Example:

```env
AUDIT_LOG_RETENTION_MONTHS=3
```

Future configuration could be:

```env
AUDIT_LOG_RETENTION_MONTHS=12
```

without changing application code.

---

# 24. Audit Cleanup

Create a dedicated cleanup function/service.

Conceptually:

```text
cleanupAuditLogs()
```

It should:

1. Find audit year/month directories.
2. Determine their date.
3. Compare them with the retention period.
4. Delete only directories older than the retention period.
5. Never delete the current month.
6. Never delete files outside the audit-log directory.
7. Handle missing directories safely.
8. Report cleanup errors through the existing logger.

Do NOT implement cleanup as:

```bash
rm -rf audit-logs/*
```

---

# 25. Cleanup Schedule

Do not wait exactly three months between cleanup executions.

Run cleanup periodically, preferably **once per month**.

Each execution checks:

```text
Is this audit data older than AUDIT_LOG_RETENTION_MONTHS?
       ↓
YES → delete
NO  → keep
```

The cleanup can be implemented using:

- VPS cron
- a dedicated Docker cleanup process
- another existing scheduled-task mechanism

Prefer the approach that best matches the existing deployment.

Do not introduce an unnecessary long-running goroutine if the VPS already supports cron/scheduled jobs.

---

# 26. Retention Safety

Before deleting logs:

- calculate the cutoff date correctly
- use year/month boundaries safely
- never delete current/future months
- handle January/year transitions
- handle missing folders
- log what was deleted

Example:

```text
Cleanup started
Retention: 3 months
Deleted:
  audit-logs/2026/03/
  audit-logs/2026/04/
```

Do not log the actual sensitive audit contents during cleanup.

---

# 27. Recommended Backend Files

Follow the existing project structure.

Possible implementation:

```text
taga-api/
├── service/
│   └── audit/
│       ├── audit.go
│       ├── storage.go
│       ├── sanitizer.go
│       └── cleanup.go
│
├── handler/
│   └── admin_audit.go
│
└── router/
    └── router.go
```

Do not create duplicate architecture if equivalent services already exist.

---

# 28. Existing Backend Integration

Inspect the existing handlers before modifying them.

Potential files may include:

```text
handler/admin_auth.go
handler/member_auth.go
handler/admin_members.go
```

But do not assume these exact files exist.

Add audit logging to the actual existing handlers/services after inspection.

Important operations include:

```text
LOGIN
LOGOUT
CREATE
UPDATE
DELETE
ROLE_CHANGED
PERMISSION_CHANGED
PASSWORD_CHANGED
```

Use the actual module/resource names from the existing application.

---

# 29. Update Audit Logging

For UPDATE operations:

1. Read the existing record before modification.
2. Store the old state.
3. Perform the modification.
4. Obtain the resulting new state.
5. Sanitize both.
6. Create the audit record.

Example:

```text
Before:

phone = 9876543210

Update

After:

phone = 9876500000
```

Audit:

```json
{
  "action": "UPDATE",
  "old_data": {
    "phone": "9876543210"
  },
  "new_data": {
    "phone": "9876500000"
  }
}
```

---

# 30. CREATE Audit

For CREATE:

```text
CREATE
 ↓
Resource created
 ↓
Audit record
```

Store the created resource data where safe.

Do not include sensitive fields.

---

# 31. DELETE Audit

For DELETE:

1. Read the existing resource before deletion.
2. Delete the resource.
3. Create the audit record containing the deleted resource information where safe.

Example:

```text
DELETE
Resource: Member #42
```

Audit should retain enough information to understand what was deleted without storing sensitive data.

---

# 32. Authentication Audit

For successful login:

```text
LOGIN
```

For failed login:

```text
LOGIN_FAILED
```

Do not store:

```text
password
password attempt
OTP
token
```

For logout:

```text
LOGOUT
```

For password changes:

```text
PASSWORD_CHANGED
```

Never store the old or new password.

---

# 33. Error Handling

Audit failures must not be silently ignored.

Use the existing application logger.

Example:

```text
Failed to write audit record:
<error>
```

Do not return internal filesystem paths or sensitive implementation details to the frontend.

For critical operations, determine whether the business operation should fail if the audit record cannot be persisted.

Use a consistent policy across the project.

---

# 34. File Permissions

Audit files contain sensitive operational information.

Use restrictive permissions appropriate for the Docker user/group.

Prefer:

```text
directories: 0750
files:       0600
```

unless the existing Docker/VPS permission model requires a different secure configuration.

Do not make audit files world-readable.

Do not expose them through Nginx static file serving.

---

# 35. Nginx

Inspect the existing Nginx configuration.

Ensure:

```text
/audit-logs/
```

is NOT publicly accessible.

Do not create a static route for the audit directory.

The only supported access path should be:

```text
/internal/audit
       ↓
GET /api/admin/audit
       ↓
Authorization
       ↓
Audit files
```

---

# 36. Verification — Backend

Run:

```bash
go test ./...
```

Also run the project's normal formatting/build commands.

Verify:

- audit package builds
- handlers build
- API builds
- tests pass

---

# 37. Verification — Audit Logging

Test:

### Login

```text
LOGIN
```

### Failed login

```text
LOGIN_FAILED
```

### Create

```text
CREATE
```

### Update

```text
UPDATE
```

Verify:

```text
old_data
new_data
```

### Delete

```text
DELETE
```

### Multiple users

Verify:

```text
user_1.json
user_2.json
```

are separate.

### Month change

Verify:

```text
2026/08/
2026/09/
```

are handled correctly.

---

# 38. Verification — Security

Test:

```text
Unauthenticated user
        ↓
/internal/audit
        ↓
DENIED
```

Test:

```text
Normal authenticated user
        ↓
/internal/audit
        ↓
DENIED
```

Test:

```text
Administrator
        ↓
/internal/audit
        ↓
ALLOWED
```

Also directly test:

```text
GET /api/admin/audit
```

without frontend access.

It must still enforce authorization.

---

# 39. Verification — Sensitive Data

Perform operations containing fields such as:

```text
password
token
secret
api_key
authorization
otp
```

Verify that actual values never appear in:

```text
audit-logs/
```

Nested sensitive values must also be sanitized.

---

# 40. Verification — Docker Persistence

Perform an action and verify:

```text
audit-logs/2026/08/user_42.json
```

exists on the host.

Then run:

```bash
docker compose restart
```

Verify the logs remain.

Then test container recreation according to the project's normal deployment process.

Verify the logs remain.

---

# 41. Verification — Retention

Temporarily configure a test retention period if the implementation supports it.

Verify that old directories are removed while current data remains.

Example:

```text
OLD
2026/01
2026/02

CURRENT
2026/08
```

After cleanup:

```text
2026/01 → deleted
2026/02 → deleted
2026/08 → retained
```

Do not perform destructive testing against real production audit data.

Use a test audit directory/environment.

---

# 42. Verification — JSON Integrity

Verify:

- every audit file contains valid JSON
- concurrent writes do not corrupt the file
- failed writes do not leave invalid JSON
- application restart does not corrupt existing logs

Use temporary-file + atomic rename where practical.

---

# 43. Verification — Audit UI

Verify:

- `/internal/audit` loads for authorized administrators
- unauthorized users cannot access it
- user filter works
- year filter works
- month filter works
- action filter works
- module filter works
- search works
- pagination works
- detail view works
- old/new data is readable
- loading state works
- empty state works
- API error state works

---

# 44. Final Code Quality Requirements

Before completing implementation:

1. Run `gofmt`.
2. Run `go test ./...`.
3. Run the frontend build/tests.
4. Verify Docker build.
5. Verify Docker Compose.
6. Check error handling.
7. Check race/concurrency behavior where practical.
8. Check path traversal protection.
9. Check sensitive-data sanitization.
10. Check Git ignore.
11. Check Nginx exposure.
12. Check Docker persistence.
13. Check retention cleanup.
14. Do not modify unrelated application behavior.

---

# 45. Agent Implementation Rules

The implementation must be done **phase by phase**.

Do NOT implement everything at once.

## Phase 1

Inspect the existing project.

Do not modify files.

Report:

- backend structure
- frontend structure
- authentication
- authorization
- JSON storage
- Docker
- Nginx
- existing routes
- recommended integration points

Wait for approval.

---

## Phase 2

Design the audit architecture.

Do not implement the complete system yet.

Confirm:

- AuditRecord structure
- file structure
- storage service
- sanitizer
- concurrency approach
- retention strategy
- Docker persistence
- API design
- UI design

Wait for approval.

---

## Phase 3

Implement the core Go audit service.

Test it independently.

Do not modify all existing handlers yet.

Wait for approval.

---

## Phase 4

Connect the audit service to real application actions.

Test:

```text
LOGIN
LOGOUT
CREATE
UPDATE
DELETE
ROLE_CHANGED
PERMISSION_CHANGED
PASSWORD_CHANGED
```

Use actual application modules discovered during Phase 1.

Wait for approval.

---

## Phase 5

Implement Docker/VPS persistence.

Verify logs survive container restart/recreation.

Wait for approval.

---

## Phase 6

Implement:

```http
GET /api/admin/audit
```

with authentication, authorization, filters, search, and pagination.

Wait for approval.

---

## Phase 7

Implement:

```text
/internal/audit
```

React UI.

Wait for approval.

---

## Phase 8

Implement retention cleanup and complete security testing.

Wait for approval before any destructive cleanup testing.

---

# 46. Final Expected Architecture

The final system should work like this:

```text
                         USER
                           │
                           ▼
                    React Application
                           │
                           ▼
                       Go API
                           │
                  ┌────────┴────────┐
                  │                 │
                  ▼                 ▼
             JSON Data        Audit Service
                                    │
                                    ▼
                              /app/audit-logs
                                    │
                              Docker Volume
                                    │
                                    ▼
                                  VPS
                                    │
                          audit-logs/YYYY/MM/
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
               user_1.json    user_2.json    user_3.json


Administrator
      │
      ▼
/internal/audit
      │
      ▼
GET /api/admin/audit
      │
      ▼
Authentication
      │
      ▼
Administrator Permission
      │
      ▼
Audit Service Reader
      │
      ▼
JSON Audit Files
      │
      ▼
Audit Dashboard
```

Retention:

```text
Monthly cleanup
       │
       ▼
Read AUDIT_LOG_RETENTION_MONTHS
       │
       ▼
Find audit data older than cutoff
       │
       ▼
Delete only expired YYYY/MM directories
       │
       ▼
Keep current audit history
```

---

# 47. Final Success Criteria

The implementation is considered complete only when:

- [ ] Audit service is modular.
- [ ] JSON is used instead of a database.
- [ ] Logs are separated by actual user ID.
- [ ] Year/month directory structure works.
- [ ] Anonymous actions are supported.
- [ ] CREATE is logged.
- [ ] UPDATE is logged with old/new data.
- [ ] DELETE is logged.
- [ ] LOGIN/LOGOUT are logged.
- [ ] Security-related actions are logged.
- [ ] Sensitive values are never stored.
- [ ] Concurrent writes are safe.
- [ ] Audit write errors are handled.
- [ ] Docker persistence works.
- [ ] Audit files are ignored by Git.
- [ ] Nginx does not expose audit files.
- [ ] Audit API requires authorization.
- [ ] Audit UI requires authorization.
- [ ] Filters work.
- [ ] Search works.
- [ ] Pagination works.
- [ ] Detail view works.
- [ ] Three-month retention works by default.
- [ ] Retention is configurable.
- [ ] Cleanup is scheduled safely.
- [ ] Old logs are deleted without affecting current logs.
- [ ] Docker restart does not lose logs.
- [ ] Go tests pass.
- [ ] Frontend build/tests pass.
- [ ] Docker build passes.
- [ ] No unrelated functionality is broken.

## Important Final Rule

**Do not introduce a database just for the audit system.**

Keep the implementation:

```text
Go
+
JSON
+
Filesystem
+
Docker persistent volume
+
Protected Admin API
+
React Audit UI
+
Configurable 3-month retention
```

The implementation should remain simple, maintainable, secure, and compatible with the existing `nammataga` architecture.