# 🌾 Nammataga — Monorepo & Operations Guide

Welcome to the **Nammataga** project repository. This repository contains the complete ecosystem for the Nammataga Employee Association platform, including the Frontend Web Application, Backend REST API, End-to-End Test Suite, Docker orchestration, and automated deployment pipelines for Development and Production environments.

---

## 📑 Table of Contents

1. [Architecture & Project Structure](#-architecture--project-structure)
2. [Tech Stack](#-tech-stack)
3. [Local Development & Docker Setup](#-local-development--docker-setup)
4. [Running End-to-End Tests (UI & API)](#-running-end-to-end-tests-ui--api)
5. [Deploying to Development (Dev)](#-deploying-to-development-dev)
6. [Deploying to Production (Prod)](#-deploying-to-production-prod)
7. [Data Synchronization & Remote Maintenance](#-data-synchronization--remote-maintenance)
8. [Configuration & Environment Variables](#-configuration--environment-variables)

---

## 🏛 Architecture & Project Structure

The repository is organized into dedicated sub-packages:

```
nammataga/
├── taga-web/                  # React + Vite + Tailwind frontend application
├── taga-api/                  # Go (Gin) REST API backend
├── taga-test/                 # Go-based E2E Test Suite (Selenium UI + API + HTML Reports)
├── dev_environment/           # Development Docker Compose & deployment scripts
├── prod_environment/          # Production Docker Compose & deployment scripts
├── scripts/                   # Utility scripts (publishing, sync, data migration, tools)
├── docker-compose.yml         # Root local Docker Compose definition
├── docker-setup-local.sh      # Local Docker container management script
├── config.json                # Root / fallback configuration
└── README.md                  # Master repository documentation
```

### Submodule Overview

| Module | Description | Primary Language / Framework |
| :--- | :--- | :--- |
| **`taga-web`** | User & Admin portal (Membership, TAGA Towers booking, Grievances, Events, Content) | React 18, TypeScript, Vite, Tailwind CSS, Radix / Shadcn UI |
| **`taga-api`** | Backend service exposing authenticated endpoints, business logic, and JSON DB storage | Go 1.22+, Gin, Zap Logger, JWT, Razorpay Integration |
| **`taga-test`** | Comprehensive automated test framework covering API endpoints and Chrome UI flows | Go, Selenium WebDriver, Page Object Model (POM), HTML Reporter |
| **`scripts`** | Deployment, synchronizing production databases, and operational automation | Bash, Python |

---

## 🛠 Tech Stack

- **Frontend**: React, TypeScript, Vite, Tailwind CSS, Lucide Icons, Radix UI primitives, Sonner toasts.
- **Backend**: Go, Gin HTTP framework, Uber Zap structured logging, Bcrypt, JWT auth, Razorpay API.
- **Storage / Database**: Structured JSON & Flat File DB under `taga-api/data/` with automated file-locking.
- **Audit Logging**: System-wide actor & transaction audit logger recording all member and admin actions under `taga-api/audit-logs/`.
- **Testing**: Go `testing` framework, Tebeka Selenium WebDriver, ChromeDriver, custom HTML/JSON reporters.
- **Containerization**: Multi-stage Dockerfiles, Nginx Alpine, Docker Compose.

---

## 💻 Local Development & Docker Setup

You can run the entire application locally either using the automated Docker setup or by running the services directly on your host machine.

### Option A: Local Docker Setup (Recommended)

Use the root management script [`docker-setup-local.sh`](file:///home/sudhan_dev/Downloads/code/nammataga/docker-setup-local.sh):

```bash
# 1. Start all containers in the background
./docker-setup-local.sh start

# 2. Check status and health logs
./docker-setup-local.sh status

# 3. View live backend logs
./docker-setup-local.sh logs

# 4. Rebuild images after code modifications
./docker-setup-local.sh build

# 5. Full rebuild without cache (if needed)
docker compose build --no-cache && docker compose up -d --force-recreate

# 6. Stop and clean up containers
./docker-setup-local.sh stop
```

#### Local Endpoints:
- **Frontend UI**: [http://localhost:1701](http://localhost:1701)
- **Backend API**: [http://localhost:1801](http://localhost:1801)
- **Health Check**: [http://localhost:1801/health](http://localhost:1801/health)

---

### Option B: Running Directly on Host

#### 1. Backend (`taga-api`):
```bash
cd taga-api
go mod tidy
go run main.go
# Backend starts on http://localhost:8080 (or configured PORT)
```

#### 2. Frontend (`taga-web`):
```bash
cd taga-web
npm install
npm run dev
# Frontend dev server starts on http://localhost:5173
```

---

## 🧪 Running End-to-End Tests (UI & API)

The [`taga-test`](file:///home/sudhan_dev/Downloads/code/nammataga/taga-test) framework runs both **API** and **Selenium UI** tests, generating comprehensive interactive HTML reports with screenshots, request/response logging, and pass/fail metrics.

### Prerequisites for Testing:
1. **Google Chrome** & **Chromedriver** installed on your system:
   ```bash
   sudo apt-get update && sudo apt-get install -y chromium-browser chromium-chromedriver
   ```
2. Target backend/frontend servers running (locally or on dev/prod).

---

### 1. Run API Tests Only
Execute all API tests or target specific scenarios using regular expressions:

```bash
cd taga-test

# Run entire API test suite
./run-api-tests.sh

# Run specific API test file or test pattern
./run-api-tests.sh -run "TestAPI_Public"
./run-api-tests.sh -run "TestAPI_SessionSecurity"
./run-api-tests.sh -run "TestAPI_Tower"
```

---

### 2. Run UI Tests Only
UI tests launch Chromedriver, automate browser interactions (Page Object Model), and capture screenshots at every step:

```bash
cd taga-test

# Run entire UI test suite (headed browser)
./run-ui-tests.sh

# Run specific UI workflow by name/number
./run-ui-tests.sh -run "TestUI_01_MemberLogin"
./run-ui-tests.sh -run "TestUI_14_TAGATowers"

# Run in Headless mode (CI / background)
E2E_HEADLESS=true ./run-ui-tests.sh
```

---

### 3. Run Full Combined Suite (UI + API)
```bash
cd taga-test
./run-tests.sh
```

---

### 4. Viewing Test Reports
After any test run, the framework generates:
- **Interactive HTML Dashboard**: `taga-test/evidence/run-<timestamp>/reports/report.html`
- **Machine-readable JSON**: `taga-test/evidence/run-<timestamp>/reports/report.json`
- **Markdown Summary**: `taga-test/evidence/run-<timestamp>/test-report.md`
- **Full API Request/Response Payloads**: `taga-test/evidence/run-<timestamp>/requests/`
- **Step Screenshots**: `taga-test/evidence/run-<timestamp>/screenshots/`

In non-headless mode, the test runner automatically opens `report.html` in your default browser. The exact path is also printed in the terminal:
```text
=========================================
 API Tests finished with exit code 0
📊 HTML Report: /home/.../taga-test/evidence/run-2026-08-27_17-21-01/reports/report.html
=========================================
```

---

## 🚀 Deploying to Development (Dev)

### Target Environment:
- **Frontend**: [https://dev.nammataga.com](https://dev.nammataga.com)
- **Backend API**: [https://devapi.nammataga.com](https://devapi.nammataga.com)
- **VPS Destination**: `/apps/taga-api/dev` and `/apps/taga-web/dev`

---

### Smart Dev Deployment (Dockerized)

The [`dev_environment/dev-publish.sh`](file:///home/sudhan_dev/Downloads/code/nammataga/dev_environment/dev-publish.sh) script automatically detects git changes, builds the changed services into tarball packages, and deploys them to the server:

```bash
# Standard smart build (only builds services with git modifications)
./dev_environment/dev-publish.sh

# Force rebuild & deploy of both frontend and backend
./dev_environment/dev-publish.sh --force
```

#### Remote Activation:
After uploading, activate the new containers on the dev VPS:
```bash
ssh sys-taga@taga-prod
sudo bash /apps/taga-api/dev/dev-deploy-docker.sh
```

---

### Component-Specific Script Publishing (Direct Binary / Static Build)

If you prefer building and deploying individual components without Docker tarballs:

```bash
# 1. Deploy API binary directly to dev server
./scripts/dev-publish-api.sh

# 2. Build and sync frontend dist files to dev server
./scripts/dev-publish-web.sh
```

---

## 🚢 Deploying to Production (Prod)

### Target Environment:
- **Frontend**: [https://nammataga.com](https://nammataga.com)
- **Backend API**: [https://api.nammataga.com](https://api.nammataga.com)
- **VPS Destination**: `31.97.62.187` / `/apps/taga-api/prd`

---

### Production Deployment Workflow

1. **Build & Package Production Docker Images**:
   ```bash
   # Smart incremental build based on git diffs
   ./prod_environment/prod-docker-publish.sh

   # Or force rebuild of both images
   ./prod_environment/prod-docker-publish.sh --force
   ```
   *This compiles optimized production Docker images, generates `dist/taga-api-prd.tar.gz` & `dist/taga-web-prd.tar.gz`, and syncs compose files to the VPS.*

2. **Deploy on Production VPS**:
   ```bash
   ssh dev-taga@31.97.62.187
   sudo bash /apps/taga-api/prd/prd-deploy-docker.sh
   ```

3. **Verify Production Services**:
   - Check container status: `sudo docker ps`
   - Check backend health: `curl -f https://api.nammataga.com/health`

---

## 🔄 Data Synchronization & Remote Maintenance

To safely pull data, database JSON files, and uploads from the remote production server to your local development environment:

```bash
./scripts/sync-remote-data.sh
```

*This downloads remote data files into `taga-api/data/` while maintaining correct local user permissions and logging the synchronization timestamp.*

---

## ⚙️ Configuration & Environment Variables

### Backend Configuration (`taga-api`)
Environment variables are defined in `taga-api/data/.env` or passed via Docker Compose:

| Variable | Description | Example / Default |
| :--- | :--- | :--- |
| `PORT` | API Server Port | `1801` (Docker) / `8080` (Local) |
| `ENVIRONMENT` | Runtime mode (`development` / `production`) | `development` |
| `JWT_SECRET` | Secret key used for signing session tokens | `[secure-random-string]` |
| `RAZORPAY_KEY_ID` | Razorpay Key ID for payments | `rzp_test_...` |
| `RAZORPAY_KEY_SECRET` | Razorpay Key Secret | `...` |

---

### Frontend Configuration (`taga-web`)
Configured in `taga-web/.env`, `.env.development`, and `.env.production`:

| Variable | Description | Dev Value | Prod Value |
| :--- | :--- | :--- | :--- |
| `VITE_API_BASE_URL` | Base endpoint for REST API calls | `https://devapi.nammataga.com/api` | `https://api.nammataga.com/api` |

---

### Test Suite Configuration (`taga-test/config.json`)
Settings in `taga-test/config.json` can be overridden using environment variables:

| Setting | Environment Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| `baseUrl` | `E2E_BASE_URL` | `https://devapi.nammataga.com` | API host target |
| `uiUrl` | `E2E_UI_URL` | `https://dev.nammataga.com/` | Web frontend target |
| `seleniumUrl` | `E2E_SELENIUM_URL` | `http://localhost:9515` | ChromeDriver URL |
| `headless` | `E2E_HEADLESS` | `true` | Toggle headless browser execution |
| `timeout` | `E2E_TIMEOUT` | `10` | Default timeout (seconds) |

---

## 🛡️ Security & Quality Assurance

- **Audit Trail**: Every administrative action, member profile modification, grievance status update, and booking transaction is securely logged with actor IDs (`tagaId` / email), IP addresses, timestamps, and before/after state diffs.
- **Negative & Security Test Suites**: Includes automated testing for SQL injection, XSS inputs, parameter tampering, expired/revoked JWT tokens, future timestamps, and concurrent booking race conditions.

---

