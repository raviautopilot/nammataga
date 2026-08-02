# Nammataga Shell Scripts Directory

This directory contains shell scripts used for development, dependencies setup, production builds, data synchronization, and codebase indexing.

## Summary of Scripts

| Script Name | Purpose | Target Environment / Location | Permissions / Prerequisites |
| :--- | :--- | :--- | :--- |
| [`dev-publish-api.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/dev-publish-api.sh) | Builds API and deploys binary to dev server | `taga-prod` / `/apps/taga-api/dev` | SSH access, `rsync` |
| [`dev-publish-web.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/dev-publish-web.sh) | Validates environment, builds frontend, and deploys to dev server | `taga-prod` / `/apps/taga-web/dev` | SSH access, npm, `rsync` |
| [`install-graphify.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/install-graphify.sh) | Installs Python 3, Pip, and `graphify` globally | Local Workstation / VM | `sudo` / Root privileges |
| [`npm-install-web.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/npm-install-web.sh) | Helper to run npm installs in `taga-web` directory | Local Workstation | Node.js, npm |
| [`prod-docker-publish.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/prod-docker-publish.sh) | Smart production Docker builds, tarball packaging, and deploy | VPS `31.97.62.187` / `/apps/taga-api/prd` | SSH access, Docker daemon, git |
| [`sync-remote-data.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/sync-remote-data.sh) | Syncs database/data folders from production to local | Local Destination: `taga-api/data/` | SSH alias `sys-taga` configured, `rsync` |
| [`token-saver.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/token-saver.sh) | Installs graphify locally and indexes codebase to save tokens | Local workspace repositories | Python 3, pip, graphifyy CLI |

---

## Script Details

### 1. `dev-publish-api.sh`

Deploys the development backend binary to the development environment server.

*   **Location:** [`scripts/dev-publish-api.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/dev-publish-api.sh)
*   **Prerequisites:** 
    *   `rsync` installed on host and target machine.
    *   Configured SSH access to `taga-prod` for username `dev-taga`.
*   **Execution Flow:**
    1. Triggers the build script `taga-api/build.sh`.
    2. Builds the `taga-api` binary.
    3. Uploads the final binary file to the server using:
       ```bash
       rsync -az -e "ssh" taga-api/taga-api dev-taga@taga-prod:/apps/taga-api/dev/
       ```
*   **Usage:**
    ```bash
    ./scripts/dev-publish-api.sh
    ```

---

### 2. `dev-publish-web.sh`

Deploys the development frontend application to the development server.

*   **Location:** [`scripts/dev-publish-web.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/dev-publish-web.sh)
*   **Prerequisites:**
    *   Node.js and npm installed locally.
    *   Active configuration variables set inside `taga-web/.env`, `taga-web/.env.development`, and `taga-web/.env.production`.
*   **Validation Rules:**
    *   Verifies that the above environmental configuration files exist.
    *   Ensures that `VITE_API_BASE_URL` is defined and non-empty in each config file.
*   **Execution Flow:**
    1. Installs npm dependencies inside `taga-web` via `npm install`.
    2. Runs `npm run build:dev` to compile a development-configured production frontend package in `dist/`.
    3. Overwrites the client environment configuration file (`dist/env-config.js`) to target `https://devapi.nammataga.com/api` instead of local dev hosts.
    4. Syncs compilation directories (`dist/`) using `rsync` over SSH to the development VPS location `/apps/taga-web/dev`.
*   **Usage:**
    ```bash
    ./scripts/dev-publish-web.sh
    ```

---

### 3. `install-graphify.sh`

A provisioning script to globally install the `graphify` code knowledge-base library.

*   **Location:** [`scripts/install-graphify.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/install-graphify.sh)
*   **Prerequisites:**
    *   Must be run with **root** or `sudo` privileges.
    *   Target systems: Debian/Ubuntu systems using the `apt` package manager.
*   **Logging:** All logs (successes and errors) are written to `/var/log/graphify_install.log`.
*   **Execution Steps:**
    1. Performs `apt-get update`.
    2. Installs `python3` and `python3-pip` packages.
    3. Automatically appends `--break-system-packages` flag on newer versions of Debian/Ubuntu to bypass Python environment restrictions.
    4. Installs the `graphifyy[gemini]` package using `pip3`.
    5. Validates system environment paths to ensure `graphify` binary is discoverable and responds to `--help`.
*   **Usage:**
    ```bash
    sudo ./scripts/install-graphify.sh
    ```

---

### 4. `npm-install-web.sh`

A simple utility helper command that initializes the npm configurations for the frontend workspace.

*   **Location:** [`scripts/npm-install-web.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/npm-install-web.sh)
*   **Usage:**
    ```bash
    ./scripts/npm-install-web.sh
    ```

---

### 5. `prod-docker-publish.sh`

Performs smart and efficient production builds utilizing Docker image caching and Git diff optimizations.

*   **Location:** [`scripts/prod-docker-publish.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/prod-docker-publish.sh)
*   **Prerequisites:**
    *   Docker engine installed and active daemon on local machine.
    *   Git initialized inside repository.
    *   Configured SSH access to production VPS `31.97.62.187` for username `dev-taga`.
*   **Features:**
    *   **Incremental Builds:** Performs git diff status checks across directories `taga-api/` and `taga-web/`. If no changes are found, image rebuilding is bypassed to save disk resources and time.
    *   **Force Deployment:** Accepts `--force` option to ignore git changes and force complete reconstruction of images.
    *   **Archiving:** Compiles target images as local Gzip Tarballs in the project's root `dist/` directory:
        *   `dist/taga-api-prd.tar.gz`
        *   `dist/taga-web-prd.tar.gz`
    *   **VPS Deployment logs:** Automates release information log inserts on the production server under `/apps/taga-api/prd/deploy.log`.
*   **Usage:**
    *   *Standard build (smart rebuilding):*
        ```bash
        ./scripts/prod-docker-publish.sh
        ```
    *   *Forced build:*
        ```bash
        ./scripts/prod-docker-publish.sh --force
        ```
*   **Post-execution Action:** After transferring the package files, you must ssh into the VPS and execute the deploy command:
    ```bash
    ssh dev-taga@31.97.62.187
    sudo bash /apps/taga-api/prd/prd-deploy-docker.sh
    ```

---

### 6. `sync-remote-data.sh`

Data synchronization utility script to pull production data/sqlite database directories safely to the local workstation for test purposes.

*   **Location:** [`scripts/sync-remote-data.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/sync-remote-data.sh)
*   **Prerequisites:**
    *   SSH configuration alias `sys-taga` must be declared inside the local `~/.ssh/config` file.
    *   `rsync` installed locally and on VPS.
*   **Options Used:**
    *   Downloads files through `sudo rsync` on the VPS backend target.
    *   Enforces permission rules locally with `--chmod=ugo=rwX` and ignores target user metadata via `--no-owner --no-group --no-perms` to avoid local validation errors.
*   **Output Details:**
    *   Saves synchronization logs to: `/home/ubuntu/code/github/raviautopilot/nammataga/logs/sync-[DATE]-[TIME].log`.
    *   Destination folder: `taga-api/data/`.
*   **Usage:**
    ```bash
    ./scripts/sync-remote-data.sh
    ```

---

### 7. `token-saver.sh`

Codebase structural indexing helper file that saves LLM query token consumption by cataloging structure trees.

*   **Location:** [`scripts/token-saver.sh`](file:///home/ubuntu/code/github/raviautopilot/nammataga/scripts/token-saver.sh)
*   **Prerequisites:**
    *   Python 3 and Pip installed locally.
*   **Execution Flow:**
    1. Checks if local packages `graphifyy` and `openai` are present; if not, installs them via local pip setup options (`--user`).
    2. Runs `graphify extract . --backend gemini` in both `taga-api` and `taga-web` directory trees.
*   **Usage:**
    ```bash
    ./scripts/token-saver.sh
    ```
