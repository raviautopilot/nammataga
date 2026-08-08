# Audit Log System — Technical Guide & Executive Q&A

This document serves as a complete testing manual and an executive explanation guide to present the implemented Audit Log System to management.

---

## 1. System Architecture Overview

The system is built as a **lightweight, modular, filesystem-based JSON logging engine** integrated directly into the `nammataga` Go API and React UI, conforming to all architecture guidelines:

```
                      ┌───────────────────────┐
                      │   React Frontend UI   │
                      │   (/internal/audit)   │
                      └───────────┬───────────┘
                                  │ (JWT Admin Token)
                                  ▼
                      ┌───────────────────────┐
                      │    Go Backend API     │
                      │  (GET /api/admin/audit)│
                      └───────────┬───────────┘
                                  │
                                  ▼
                      ┌───────────────────────┐
                      │  Audit Log Service    │
                      │ (taga-api/service/aud)│
                      └───────────┬───────────┘
                                  │
                                  ▼
                     [VPS Host Path: /audit-logs]
                     └── 2026/
                         └── 08/
                             ├── user_T-001.json
                             ├── user_admin.json
                             └── user_anonymous.json
```

---

## 2. Key Implemented Features

1. **Partitioned File Storage**: Files are categorized monthly (`audit-logs/YYYY/MM/`) and isolated by the user's identifier (Member's `tagaId` e.g., `user_T-042.json`, `user_admin.json` for administrators, or `user_anonymous.json` for unauthenticated requests).
2. **Safe Concurrent Writes**: Utilizes process-level write locks (Mutex) and writes to a temporary file (`.tmp`) before swapping it atomically using a POSIX rename. This prevents partial JSON serialization or file corruption during heavy load.
3. **Recursive Sensitive-Data Redaction**: Intercepts values recursively and replaces sensitive fields (passwords, hashes, tokens, authorization headers) with `[REDACTED]`.
4. **Three-Month Automatic Retention**: Automatically scans and deletes monthly audit folders older than the configured threshold (default `3` months), via a background monthly scheduler.
5. **Secure Internal Dashboard**: Located at the frontend route `/internal/audit`, protected by JWT authentication and admin role enforcement on both frontend and backend.
6. **Path Traversal Protection**: Any search for files restricts inputs to alphanumeric characters and validates matches to ensure absolute safety.

---

## 3. How to Test the System (Step-by-Step)

Follow these steps to demonstrate and verify the functionality:

### Step 1: Run Automated Backend Tests
Before starting the servers, verify the logic using the new test suite:
```bash
cd taga-api/service/audit
go test -v
```
*Expected Output:* All tests (`TestSanitize`, `TestSanitizeFilenameID`, `TestLogAndCleanup`, and `TestRetentionMonths`) should pass.

### Step 2: Spin Up the Application
Start the system inside Docker:
```bash
# From the root directory containing docker-compose.yml
docker compose down
docker compose up --build
```
*Verification:* Inspect the `taga-api/` folder. You will see that `audit-logs/` is ignored by Git but mounted to the backend container under `/app/audit-logs`.

### Step 3: Trigger Audit Logs (Manual Actions)
1. **Unauthenticated Failed Login Attempt:**
   - Go to the Member Login page.
   - Enter an incorrect password for an email.
   - *Result:* Open `taga-api/audit-logs/YYYY/MM/user_anonymous.json`. You will see a `LOGIN_FAILED` record recording the email and IP address.
2. **Successful Member Login & Profile Update:**
   - Log in with valid member credentials.
   - Go to the profile page as a member and update your profile details.
   - *Result:* `LOGIN` and `UPDATE` records are logged in `user_T-XXX.json` (using their actual `tagaId`).
3. **Administrator Member CRUD:**
   - Log in as the administrator.
   - Go to the Admin Dashboard and update or add a member.
   - *Result:* A `CREATE` or `UPDATE` record is appended to `user_admin.json`.
4. **Administrator Content Management (Events, Resources, Gallery):**
   - As an admin, create an Event, upload a Resource document, or upload a Gallery image.
   - Delete an event, resource, or gallery image.
   - *Result:* Actions are logged inside `user_admin.json` under modules `EVENT`, `RESOURCE`, or `GALLERY` with exact details of created, modified, or deleted data.
5. **Administrator Announcements & Reminders:**
   - As an admin, send an Announcement to members or trigger manual renewal reminders.
   - *Result:* Actions are logged inside `user_admin.json` detailing the recipient filters or triggered processes.
6. **Administrator Office Bearers Management:**
   - As an admin, update district office bearers or restore from backup files.
   - *Result:* Actions are logged inside `user_admin.json` capturing old and new district bearers or restored backup filename.

### Step 4: Verify the Admin Dashboard
1. Log in to the application as the administrator.
2. Navigate to: `http://localhost:1701/internal/audit` (or the corresponding VPS domain).
3. Verify that you can:
   - Filter by user ID (`tagaId`).
   - Filter by Year, Month, Action, and Module.
   - Use the text search input.
   - Click the "Eye" icon on an `UPDATE` record to inspect the side-by-side green/red (old/new) comparative diff modal.
4. Try accessing `/internal/audit` from a private window or as a non-admin member. Verify that the UI redirects you to Home or displays an Access Denied message.

---

## 4. Q&A Guide for Your Boss (Answering Key Questions)

Here are the answers to questions your boss or security audit teams are likely to ask:

### Q1: Why did we build a custom filesystem JSON approach instead of using a database?
**Answer:**
- **Zero-Dependency Constraint:** It perfectly complies with the instruction to avoid introducing external databases (PostgreSQL, MongoDB, etc.) to keep the VPS footprint simple and lightweight.
- **Easy Archiving & Portability:** Monthly JSON files are incredibly simple to back up, archive, compress, or transfer to cold storage.
- **Fast Lookup:** Because logs are grouped by `year/month/user_id`, the API does not need to query a huge table of millions of rows. It simply reads the specific user's monthly file directly, resulting in sub-millisecond response times.

### Q2: What happens if multiple users perform actions at the exact same millisecond? Will the JSON file get corrupted?
**Answer:** No. We implemented two layers of protection:
1. **Process-Level Mutex:** In Go, we use a global sync Mutex (`mu.Lock()`) protecting the read-modify-write cycle, so only one write request accesses a user's JSON file at any given moment.
2. **Atomic Write Replacement:** We write the new JSON structure to a temporary file (`.tmp`) first. Once the write is successfully finished, we perform an atomic rename operation (`os.Rename`) on the operating system level to overwrite the old file. If the system crashes midway, the old file remains 100% intact and valid.

### Q3: How are we protecting sensitive credentials (like passwords or tokens) from leaking into the logs?
**Answer:**
We built a recursive whitelist-based sanitizer that inspects every incoming payload. If any field matches or contains sensitive keys (such as `password`, `token`, `secret`, `jwt`, `otp`, `hash`), the value is immediately replaced with `[REDACTED]` in-memory before it is written to the VPS. Furthermore, Go struct tags (like `json:"-"`) are respected automatically because payloads are serialized through a JSON marshalling round-trip.

### Q4: If the Docker container crashes, gets updated, or is rebuilt, will we lose all audit logs?
**Answer:** No. We configured a persistent Docker bind mount in the `docker-compose.yml` file mapping `./taga-api/audit-logs` (VPS directory) to `/app/audit-logs` (inside the container). The logs reside on the VPS host system, completely independent of the Docker container lifecycle.

### Q5: How do we prevent users or hackers from accessing or modifying the audit files?
**Answer:**
- **OS Level Permissions:** Audit directories are created with `0750` permissions and log files with `0600` permissions, meaning only the app runner process can read/write them.
- **Web Server Protection:** The Nginx proxy config does not expose the `audit-logs/` folder.
- **API Protection:** The `/api/admin/audit` endpoints require a valid JWT signature and an admin role check. Any non-admin attempt is rejected by the backend with a `401 Unauthorized` status.
- **Path Traversal Protection:** The API query parameters (like `taga_id` or `year/month`) are sanitized and validated against safe regular expressions, preventing path traversal attacks (like `?taga_id=../../`).

### Q6: How is log cleanup handled? Will the VPS run out of disk space?
**Answer:**
We built an in-process monthly scheduler running on the backend. Every month, it calculates a cutoff date using the configurable retention environment variable (default: 3 months). It deletes year/month folders older than the cutoff, ensuring the storage footprint remains flat and predictable without any manual maintenance.

### Q7: If we scale to multiple backend container instances in the future, what changes will be needed?
**Answer:**
The current implementation assumes a single container instance (supported by a process-local mutex). If we scale horizontally to multiple API containers in the future, we would upgrade the process mutex in `service/audit/storage.go` to an advisory file lock (e.g. using `flock` or a shared lock manager) to coordinate writes to the shared network-attached VPS filesystem. The rest of the business logic and handlers will not need any changes.

### Q8: Why did the local Docker build fail with "permission denied" on audit-logs, and how is it fixed?
**Answer:**
When unit tests or local execution write audit log files, they can be created under ownership that the local docker runner doesn't have read access to. During `docker compose build`, Docker tries to send the entire `taga-api/` directory context to the build daemon, hitting this directory and failing. We fixed this by creating `taga-api/.dockerignore` and adding `audit-logs/` to it. This tells Docker to completely ignore the local logs folder during builds, preventing any context-transfer errors.

### Q9: How do we feed external environment variables like Razorpay keys into Docker Compose?
**Answer:**
We added the `env_file` block pointing to `./taga-api/data/.env` under the `taga-backend` service in `docker-compose.yml`. This tells Docker Compose to automatically load all environment variables (including Razorpay keys and SMTP credentials) from the `.env` file at container startup and inject them into the Go process context, solving any "keys not loaded" configuration issues.
