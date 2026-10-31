#!/usr/bin/env bash
# ==============================================================================
#  _   _                                _____ _    ____    _   
# | \ | | __ _ _ __ ___  _ __ ___   __ |_   _/ \  / ___|  / \  
# |  \| |/ _` | '_ ` _ \| '_ ` _ \ / _` || |/ _ \| |  _  / _ \ 
# | |\  | (_| | | | | | | | | | | | (_| || / ___ \ |_| |/ ___ \
# |_| \_|\__,_|_| |_| |_|_| |_| |_|\__,_||_/_/   \_\____/_/   \_\
#
# Comprehensive Mobile Management & Automation CLI
# Target: Android & iOS | Package: com.ravi.taga | Directory: nammataga-mobile/
# ==============================================================================

set -eo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="$APP_DIR/.web_server.pid"
LOG_FILE="$APP_DIR/.web_server.log"
ENV_FILE="$APP_DIR/.last_env"
WEB_PORT=1702
PACKAGE_NAME="com.ravi.taga"

# ------------------------------------------------------------------------------
# Rich Color Palette (ANSI 256 / True Color)
# ------------------------------------------------------------------------------
CLR_RESET="\033[0m"
CLR_BOLD="\033[1m"
CLR_DIM="\033[2m"

# Brand Greens
CLR_FOREST="\033[38;5;28m"
CLR_EMERALD="\033[38;5;35m"
CLR_MINT="\033[38;5;120m"
CLR_SOFT_MINT="\033[38;5;157m"

# Accents
CLR_CYAN="\033[38;5;45m"
CLR_GOLD="\033[38;5;220m"
CLR_RED="\033[38;5;196m"
CLR_WHITE="\033[38;5;255m"
CLR_GRAY="\033[38;5;244m"

# Background Highlights
BG_EMERALD="\033[48;5;28m"
BG_RED="\033[48;5;160m"
BG_DARK="\033[48;5;234m"

# ------------------------------------------------------------------------------
# Banner & UI Components
# ------------------------------------------------------------------------------
print_banner() {
    clear 2>/dev/null || true
    echo -e "${CLR_EMERALD}${CLR_BOLD}"
    cat << 'EOF'
  ╔══════════════════════════════════════════════════════════════════╗
  ║   _   _                                _____ _    ____    _      ║
  ║  | \ | | __ _ _ __ ___  _ __ ___   __ |_   _/ \  / ___|  / \     ║
  ║  |  \| |/ _` | '_ ` _ \| '_ ` _ \ / _` || |/ _ \| |  _  / _ \    ║
  ║  | |\  | (_| | | | | | | | | | | | (_| || / ___ \ |_| |/ ___ \   ║
  ║  |_| \_|\__,_|_| |_| |_|_| |_| |_|\__,_||_/_/   \_\____/_/   \_\  ║
  ║                                                                  ║
  ║      OFFICIAL MOBILE APPLICATION ARCHITECT & BUILD MANAGER       ║
  ║         Package: com.ravi.taga  •  Flutter Cross-Platform        ║
  ╚══════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${CLR_RESET}"
}

print_header() {
    local title="$1"
    echo ""
    echo -e "  ${CLR_MINT}●${CLR_RESET} ${CLR_BOLD}${CLR_WHITE}$title${CLR_RESET}"
    echo -e "  ${CLR_GRAY}──────────────────────────────────────────────────────────────────${CLR_RESET}"
}

print_success() {
    echo -e "  ${CLR_EMERALD}${CLR_BOLD}✓${CLR_RESET} ${CLR_WHITE}$1${CLR_RESET}"
}

print_info() {
    echo -e "  ${CLR_CYAN}${CLR_BOLD}ℹ${CLR_RESET} ${CLR_WHITE}$1${CLR_RESET}"
}

print_warn() {
    echo -e "  ${CLR_GOLD}${CLR_BOLD}⚠${CLR_RESET} ${CLR_WHITE}$1${CLR_RESET}"
}

print_error() {
    echo -e "  ${CLR_RED}${CLR_BOLD}✖${CLR_RESET} ${CLR_WHITE}$1${CLR_RESET}"
}

# ------------------------------------------------------------------------------
# Environment Resolution Utility
# ------------------------------------------------------------------------------
select_environment() {
    local default_env="${1:-local}"
    echo ""
    echo -e "  ${CLR_BOLD}${CLR_WHITE}Select Target Backend API Environment:${CLR_RESET}"
    echo -e "    ${CLR_EMERALD}[1] Localhost${CLR_RESET}   ${CLR_GRAY}→  http://localhost:1801/api (Default for local testing)${CLR_RESET}"
    echo -e "    ${CLR_CYAN}[2] Development${CLR_RESET} ${CLR_GRAY}→  https://devapi.nammataga.com/api (Dev Cloud Server)${CLR_RESET}"
    echo -e "    ${CLR_GOLD}[3] Production${CLR_RESET}  ${CLR_GRAY}→  https://api.nammataga.com/api (Live Member Services)${CLR_RESET}"
    echo ""
    read -rp "  Enter environment choice [1-3] (Default: 1): " env_choice
    env_choice="${env_choice:-1}"

    case "$env_choice" in
        2)
            TARGET_ENV="dev"
            TARGET_API="https://devapi.nammataga.com/api"
            ;;
        3)
            TARGET_ENV="prod"
            TARGET_API="https://api.nammataga.com/api"
            ;;
        *)
            TARGET_ENV="local"
            TARGET_API="http://localhost:1801/api"
            ;;
    esac

    echo "$TARGET_ENV" > "$ENV_FILE"
}

# ------------------------------------------------------------------------------
# 1. Start Web View Simulation (Builds & Serves 19.5:9 Phone Mockup)
# ------------------------------------------------------------------------------
start_web() {
    print_banner
    print_header "19.5:9 MOBILE WEB SIMULATION LAUNCHER"

    local env_arg="$1"
    if [ -n "$env_arg" ]; then
        case "$env_arg" in
            dev)
                TARGET_ENV="dev"
                TARGET_API="https://devapi.nammataga.com/api"
                ;;
            prod)
                TARGET_ENV="prod"
                TARGET_API="https://api.nammataga.com/api"
                ;;
            *)
                TARGET_ENV="local"
                TARGET_API="http://localhost:1801/api"
                ;;
        esac
    else
        select_environment "local"
    fi

    echo ""
    print_info "Target Configuration:"
    echo -e "    ${CLR_GRAY}• Environment:${CLR_RESET} ${CLR_EMERALD}${CLR_BOLD}$TARGET_ENV${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• API URL:${CLR_RESET}     ${CLR_CYAN}${CLR_BOLD}$TARGET_API${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Local Port:${CLR_RESET}  ${CLR_MINT}${CLR_BOLD}$WEB_PORT${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Layout:${CLR_RESET}    ${CLR_WHITE}${CLR_BOLD}19.5:9 Ultra-Realistic Mobile Device Chassis${CLR_RESET}"

    # Kill any existing server on the target port
    stop_web_silent

    print_header "COMPILING MOBILE WEB ARTIFACT"
    echo -e "  ${CLR_GRAY}Building optimized Flutter web output with 19.5:9 phone mockup...${CLR_RESET}"
    echo -e "  ${CLR_GRAY}Please wait a few moments...${CLR_RESET}"

    cd "$APP_DIR"
    local build_start
    build_start=$(date +%s)

    flutter build web \
        --dart-define=API_ENV="$TARGET_ENV" \
        --dart-define=API_BASE_URL="$TARGET_API" \
        --dart-define=IS_WEB_PREVIEW=true \
        --no-wasm-dry-run \
        > "$LOG_FILE" 2>&1

    local build_end
    build_end=$(date +%s)
    local duration=$((build_end - build_start))

    print_success "Mobile Web compilation completed in ${duration}s!"

    # Start high-performance HTTP server in the background
    print_header "STARTING LOCAL SIMULATION SERVER"
    setsid nohup python3 -m http.server "$WEB_PORT" --directory "$APP_DIR/build/web" >> "$LOG_FILE" 2>&1 &
    local server_pid=$!
    disown 2>/dev/null || true
    echo "$server_pid" > "$PID_FILE"

    sleep 1

    # Verify server is listening and responding
    if curl -s -I "http://localhost:$WEB_PORT/" | grep -q "200 OK"; then
        echo ""
        echo -e "  ${BG_EMERALD}${CLR_WHITE}${CLR_BOLD}  ● SIMULATION ONLINE  ${CLR_RESET} ${CLR_EMERALD}Mobile layout successfully running!${CLR_RESET}"
        echo ""
        echo -e "  ${CLR_BOLD}${CLR_WHITE}Access the 19.5:9 Mobile Preview on your computer:${CLR_RESET}"
        echo -e "  ${CLR_EMERALD}${CLR_BOLD}➜ http://localhost:$WEB_PORT${CLR_RESET}"
        echo ""
        echo -e "  ${CLR_GRAY}• PID:           ${server_pid}${CLR_RESET}"
        echo -e "  ${CLR_GRAY}• Logs:          $LOG_FILE${CLR_RESET}"
        echo -e "  ${CLR_GRAY}• Stop Command:  ./manage.sh stop-web${CLR_RESET}"
        echo ""

        # Try to open browser automatically if graphical desktop environment is active
        if command -v xdg-open >/dev/null 2>&1 && [ -n "$DISPLAY" ]; then
            xdg-open "http://localhost:$WEB_PORT" 2>/dev/null || true
        fi
    else
        print_error "Web server started with PID $server_pid but failed to respond on port $WEB_PORT."
        echo -e "  ${CLR_RED}Check the error logs at: $LOG_FILE${CLR_RESET}"
        exit 1
    fi
}

# ------------------------------------------------------------------------------
# 2. Stop Web View Simulation
# ------------------------------------------------------------------------------
stop_web_silent() {
    # Free port 1702 if any process is occupying it
    if fuser "$WEB_PORT/tcp" >/dev/null 2>&1; then
        fuser -k -9 "$WEB_PORT/tcp" >/dev/null 2>&1 || true
    fi

    if [ -f "$PID_FILE" ]; then
        local pid
        pid=$(cat "$PID_FILE" 2>/dev/null || true)
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
        rm -f "$PID_FILE"
    fi
}

stop_web() {
    print_banner
    print_header "STOPPING 19.5:9 WEB SIMULATION"

    local stopped_anything=false

    if fuser "$WEB_PORT/tcp" >/dev/null 2>&1; then
        fuser -k -9 "$WEB_PORT/tcp" >/dev/null 2>&1 || true
        stopped_anything=true
    fi

    if [ -f "$PID_FILE" ]; then
        local target_pid
        target_pid=$(cat "$PID_FILE" 2>/dev/null || true)
        if [ -n "$target_pid" ] && kill -0 "$target_pid" 2>/dev/null; then
            kill -9 "$target_pid" 2>/dev/null || true
            stopped_anything=true
        fi
        rm -f "$PID_FILE"
    fi

    if [ "$stopped_anything" = true ]; then
        print_success "Web server on port $WEB_PORT stopped successfully."
        echo -e "  ${CLR_GRAY}Port $WEB_PORT has been completely freed.${CLR_RESET}"
    else
        print_info "No active web simulation process was running on port $WEB_PORT."
    fi
}

# ------------------------------------------------------------------------------
# 3. Status Web View Simulation
# ------------------------------------------------------------------------------
status_web() {
    print_banner
    print_header "19.5:9 MOBILE WEB SIMULATION STATUS"

    local is_listening=false
    if fuser "$WEB_PORT/tcp" >/dev/null 2>&1; then
        is_listening=true
    fi

    if [ "$is_listening" = true ]; then
        local pid
        pid=$(fuser "$WEB_PORT/tcp" 2>/dev/null | awk '{print $1}')
        local last_env="local"
        if [ -f "$ENV_FILE" ]; then
            last_env=$(cat "$ENV_FILE")
        fi

        echo -e "  ${BG_EMERALD}${CLR_WHITE}${CLR_BOLD}  ● ONLINE  ${CLR_RESET} ${CLR_EMERALD}${CLR_BOLD}Web view simulation is active and serving!${CLR_RESET}"
        echo ""
        echo -e "  ${CLR_GRAY}• URL:         ${CLR_RESET}${CLR_CYAN}${CLR_BOLD}http://localhost:$WEB_PORT${CLR_RESET}"
        echo -e "  ${CLR_GRAY}• Port:        ${CLR_RESET}${CLR_WHITE}${WEB_PORT}${CLR_RESET}"
        echo -e "  ${CLR_GRAY}• Process PID: ${CLR_RESET}${CLR_WHITE}${pid:-Unknown}${CLR_RESET}"
        echo -e "  ${CLR_GRAY}• Environment: ${CLR_RESET}${CLR_EMERALD}${last_env^^}${CLR_RESET}"
        echo -e "  ${CLR_GRAY}• Logs:        ${CLR_RESET}${CLR_WHITE}$LOG_FILE${CLR_RESET}"

        if [ -f "$LOG_FILE" ]; then
            echo ""
            echo -e "  ${CLR_BOLD}${CLR_WHITE}Recent Web Activity Logs (last 5 requests):${CLR_RESET}"
            tail -n 5 "$LOG_FILE" | while IFS= read -r line; do
                echo -e "    ${CLR_GRAY}$line${CLR_RESET}"
            done
        fi
    else
        echo -e "  ${BG_RED}${CLR_WHITE}${CLR_BOLD}  ○ OFFLINE  ${CLR_RESET} ${CLR_RED}No simulation server is running on port $WEB_PORT.${CLR_RESET}"
        echo ""
        echo -e "  ${CLR_GRAY}To start the web simulation, execute:${CLR_RESET}"
        echo -e "  ${CLR_EMERALD}${CLR_BOLD}➜ ./manage.sh start-web${CLR_RESET}"
    fi
}

# ------------------------------------------------------------------------------
# 4. Interactive Live Logs
# ------------------------------------------------------------------------------
logs_web() {
    print_banner
    print_header "LIVE WEB SIMULATION ACCESS LOGS"
    if [ ! -f "$LOG_FILE" ]; then
        print_warn "No log file found at $LOG_FILE"
        return 0
    fi

    echo -e "  ${CLR_GRAY}Streaming $LOG_FILE (Press Ctrl+C to exit)...${CLR_RESET}"
    echo ""
    tail -f "$LOG_FILE"
}

# ------------------------------------------------------------------------------
# 5. Interactive Build Pipeline (APK & Play Store AAB)
# ------------------------------------------------------------------------------
build_playstore_deployment() {
    print_banner
    print_header "GOOGLE PLAY STORE AAB DEPLOYMENT BUNDLER"

    echo -e "  ${CLR_BOLD}${CLR_WHITE}Select Target Environment for Play Store Bundle:${CLR_RESET}"
    echo -e "    ${CLR_GOLD}${CLR_BOLD}[1] Production${CLR_RESET}  ${CLR_GRAY}→  https://api.nammataga.com/api (Official Live Release)${CLR_RESET}"
    echo -e "    ${CLR_CYAN}${CLR_BOLD}[2] Development${CLR_RESET} ${CLR_GRAY}→  https://devapi.nammataga.com/api (Internal Testing / QA Track)${CLR_RESET}"
    echo ""
    read -rp "  Select environment [1-2] (Default: 1 - Production): " target_choice
    target_choice="${target_choice:-1}"

    if [ "$target_choice" = "2" ]; then
        TARGET_ENV="dev"
        TARGET_API="https://devapi.nammataga.com/api"
    else
        TARGET_ENV="prod"
        TARGET_API="https://api.nammataga.com/api"
    fi

    echo "$TARGET_ENV" > "$ENV_FILE"

    echo ""
    echo -e "  ${CLR_BOLD}${CLR_WHITE}Play Store Application Release Version:${CLR_RESET}"
    echo -e "  ${CLR_GRAY}Enter the version for this build. Examples:${CLR_RESET}"
    echo -e "    ${CLR_CYAN}• 1.0.0+1${CLR_RESET}   ${CLR_GRAY}(Version Name: 1.0.0, Version Code: 1)${CLR_RESET}"
    echo -e "    ${CLR_CYAN}• 1.0.1${CLR_RESET}     ${CLR_GRAY}(Version Name: 1.0.1, Version Code will prompt or default to 1)${CLR_RESET}"
    echo ""
    read -rp "  Enter release version (e.g., 1.0.0+1) [Default: 1.0.0+1]: " input_version
    input_version="${input_version:-1.0.0+1}"

    local build_name=""
    local build_number=""

    if [[ "$input_version" == *"+"* ]]; then
        build_name="${input_version%%+*}"
        build_number="${input_version##*+}"
    else
        build_name="$input_version"
        read -rp "  Enter build number / version code [Default: 1]: " input_code
        build_number="${input_code:-1}"
    fi

    # Fallback sanity
    build_name="${build_name:-1.0.0}"
    build_number="${build_number:-1}"
    local app_full_version="${build_name}+${build_number}"

    local DEPLOY_ROOT="$APP_DIR/playstore_deployment"
    local DEPLOY_DIR="$DEPLOY_ROOT/$TARGET_ENV"
    local TIMESTAMP
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)

    print_header "PLAY STORE BUNDLE SPECIFICATIONS"
    echo -e "    ${CLR_GRAY}• Target Track:${CLR_RESET}     ${CLR_GOLD}${CLR_BOLD}Google Play Console Release (AAB)${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Package ID:${CLR_RESET}       ${CLR_MINT}${CLR_BOLD}$PACKAGE_NAME${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• App Version:${CLR_RESET}      ${CLR_WHITE}${CLR_BOLD}$app_full_version${CLR_RESET} ${CLR_GRAY}(VersionName: $build_name, VersionCode: $build_number)${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Environment:${CLR_RESET}      ${CLR_EMERALD}${CLR_BOLD}${TARGET_ENV^^}${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Backend API URL:${CLR_RESET}  ${CLR_CYAN}${CLR_BOLD}$TARGET_API${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Deployment Dir:${CLR_RESET}   ${CLR_WHITE}${CLR_BOLD}$DEPLOY_DIR${CLR_RESET}"
    echo ""

    # Ensure deployment directories exist
    mkdir -p "$DEPLOY_DIR"

    cd "$APP_DIR"
    local start_time
    start_time=$(date +%s)

    print_info "Executing: flutter build appbundle --release --build-name=$build_name --build-number=$build_number --dart-define=API_ENV=$TARGET_ENV --dart-define=API_BASE_URL=$TARGET_API"
    flutter build appbundle --release \
        --build-name="$build_name" \
        --build-number="$build_number" \
        --dart-define=API_ENV="$TARGET_ENV" \
        --dart-define=API_BASE_URL="$TARGET_API"

    local output_file="$APP_DIR/build/app/outputs/bundle/release/app-release.aab"
    local end_time
    end_time=$(date +%s)
    local elapsed=$((end_time - start_time))

    if [ ! -f "$output_file" ]; then
        print_error "Play Store AAB build failed! File not found at $output_file"
        return 1
    fi

    local file_size
    file_size=$(du -h "$output_file" | awk '{print $1}')

    # Clean up previous AAB builds in the target environment directory
    # so only the latest, newly built AAB exists
    rm -f "$DEPLOY_DIR"/*.aab

    # Package deployment artifacts into target folder
    local target_aab="$DEPLOY_DIR/app-${TARGET_ENV}-release.aab"
    cp "$output_file" "$target_aab"

    # Copy official store assets into deployment folder
    if [ -f "$APP_DIR/assets/images/app_icon.png" ]; then
        cp "$APP_DIR/assets/images/app_icon.png" "$DEPLOY_DIR/playstore_icon_512x512.png"
    fi
    if [ -f "$APP_DIR/assets/images/logo.png" ]; then
        cp "$APP_DIR/assets/images/logo.png" "$DEPLOY_DIR/official_logo.png"
    fi

    # Copy signing certificate and security artifacts
    if [ -f "$APP_DIR/android/keystore/nammataga_release.jks" ]; then
        cp "$APP_DIR/android/keystore/nammataga_release.jks" "$DEPLOY_DIR/nammataga_release.jks"
    fi
    if [ -f "$APP_DIR/android/key.properties" ]; then
        cp "$APP_DIR/android/key.properties" "$DEPLOY_DIR/key.properties"
    fi

    # Extract SHA-1 and SHA-256 fingerprints for Google Play Console / Firebase
    local sha1_fingerprint=""
    local sha256_fingerprint=""
    if [ -f "$DEPLOY_DIR/nammataga_release.jks" ]; then
        sha1_fingerprint=$(keytool -list -v -keystore "$DEPLOY_DIR/nammataga_release.jks" -storepass "nammataga@2026" -alias "nammataga" 2>/dev/null | grep "SHA1:" | awk '{print $2}')
        sha256_fingerprint=$(keytool -list -v -keystore "$DEPLOY_DIR/nammataga_release.jks" -storepass "nammataga@2026" -alias "nammataga" 2>/dev/null | grep "SHA256:" | awk '{print $2}')
    fi

    # Generate complete deployment manifest and metadata
    cat << EOF > "$DEPLOY_DIR/deployment_info.txt"
================================================================================
NAMMATAGA MOBILE - GOOGLE PLAY STORE PRODUCTION DEPLOYMENT PACKAGE
================================================================================
Generated At:      $(date)
Environment:       ${TARGET_ENV^^}
App Version:       $app_full_version (Version Name: $build_name, Version Code: $build_number)
Backend API URL:   $TARGET_API
Package Name:      $PACKAGE_NAME
AAB File:          $(basename "$target_aab")
File Size:         $file_size
Build Duration:    ${elapsed}s

KEYSTORE & SIGNING DETAILS:
--------------------------------------------------------------------------------
Keystore File:     nammataga_release.jks
Key Alias:         nammataga
Key / Store Pass:  nammataga@2026
Store Type:        PKCS12
Certificate Owner: CN=NammaTAGA Admin, OU=IT, O=Tamil Nadu Agricultural Graduates Association
Validity:          10,000 days (Until 2054)
SHA-1 Fingerprint:   ${sha1_fingerprint:-Unknown}
SHA-256 Fingerprint: ${sha256_fingerprint:-Unknown}

COMPLETE DEPLOYMENT ASSETS INCLUDED IN THIS FOLDER:
--------------------------------------------------------------------------------
1. app-${TARGET_ENV}-release.aab     -> Official Signed Release Android App Bundle (AAB)
2. playstore_icon_512x512.png       -> High-resolution 512x512 Play Store icon (no white padding)
3. official_logo.png                -> Official association logo
4. nammataga_release.jks            -> Cryptographic release keystore for updates
5. key.properties                   -> Gradle signing configuration file
6. deployment_info.txt              -> Full manifest and upload instructions

GOOGLE PLAY CONSOLE STEP-BY-STEP UPLOAD:
--------------------------------------------------------------------------------
1. Log in to Google Play Console: https://play.google.com/console
2. Select your application ($PACKAGE_NAME).
3. Under Release > Production (or Internal / Closed testing):
   - Click 'Create new release'.
   - Upload 'app-${TARGET_ENV}-release.aab'.
   - Add release notes.
4. Under Store presence > Main store listing:
   - App icon: Upload 'playstore_icon_512x512.png' (512x512 32-bit PNG).
5. For Google Cloud / OAuth / Firebase / Maps integration (if needed):
   - SHA-1:   $sha1_fingerprint
   - SHA-256: $sha256_fingerprint
================================================================================
EOF

    echo ""
    echo -e "  ${BG_EMERALD}${CLR_WHITE}${CLR_BOLD}  ✓ PLAY STORE DEPLOYMENT READY  ${CLR_RESET} ${CLR_EMERALD}AAB bundle generated in ${elapsed}s!${CLR_RESET}"
    echo ""
    echo -e "  ${CLR_BOLD}${CLR_WHITE}All Play Store Deployment Artifacts Staged At:${CLR_RESET}"
    echo -e "  ${CLR_GOLD}${CLR_BOLD}$DEPLOY_DIR${CLR_RESET}"
    echo ""
    echo -e "  ${CLR_GRAY}Artifacts:${CLR_RESET}"
    echo -e "    ${CLR_WHITE}• AAB Bundle:   ${CLR_GOLD}${CLR_BOLD}$target_aab${CLR_RESET} ${CLR_GRAY}($file_size)${CLR_RESET}"
    echo -e "    ${CLR_WHITE}• Store Icon:   ${CLR_MINT}$DEPLOY_DIR/playstore_icon_512x512.png${CLR_RESET}"
    echo -e "    ${CLR_WHITE}• Official Logo:${CLR_MINT}$DEPLOY_DIR/official_logo.png${CLR_RESET}"
    echo -e "    ${CLR_WHITE}• Manifest:     ${CLR_CYAN}$DEPLOY_DIR/deployment_info.txt${CLR_RESET}"
    echo ""
}

build_app() {
    print_banner
    print_header "MOBILE APPLICATION BUILD PIPELINE"

    echo -e "  ${CLR_BOLD}${CLR_WHITE}Select Target Artifact Type:${CLR_RESET}"
    echo -e "    ${CLR_GOLD}[1] Google Play Store AAB Package${CLR_RESET} ${CLR_GRAY}→ Generates playstore_deployment/(dev|prod) with AAB + store icons${CLR_RESET}"
    echo -e "    ${CLR_EMERALD}[2] Universal Release APK${CLR_RESET}        ${CLR_GRAY}→ Direct installable file for physical phone sideloading${CLR_RESET}"
    echo -e "    ${CLR_CYAN}[3] Development Debug APK${CLR_RESET}        ${CLR_GRAY}→ Quick testing build with debug logging enabled${CLR_RESET}"
    echo ""
    read -rp "  Enter artifact choice [1-3] (Default: 1): " build_type
    build_type="${build_type:-1}"

    if [ "$build_type" = "1" ]; then
        build_playstore_deployment
        return 0
    fi

    select_environment "prod"

    local artifact_name="Release APK"
    case "$build_type" in
        3) artifact_name="Debug APK" ;;
        *) artifact_name="Release APK" ;;
    esac

    print_header "BUILD SPECIFICATIONS"
    echo -e "    ${CLR_GRAY}• Target Artifact: ${CLR_RESET}${CLR_WHITE}${CLR_BOLD}$artifact_name${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Package ID:      ${CLR_RESET}${CLR_MINT}${CLR_BOLD}$PACKAGE_NAME${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Environment:     ${CLR_RESET}${CLR_EMERALD}${CLR_BOLD}$TARGET_ENV${CLR_RESET}"
    echo -e "    ${CLR_GRAY}• Backend API:     ${CLR_RESET}${CLR_CYAN}${CLR_BOLD}$TARGET_API${CLR_RESET}"
    echo ""

    cd "$APP_DIR"
    local start_time
    start_time=$(date +%s)

    case "$build_type" in
        2)
            print_info "Executing: flutter build apk --release..."
            flutter build apk --release \
                --dart-define=API_ENV="$TARGET_ENV" \
                --dart-define=API_BASE_URL="$TARGET_API"

            local output_file="$APP_DIR/build/app/outputs/flutter-apk/app-release.apk"
            local end_time
            end_time=$(date +%s)
            local elapsed=$((end_time - start_time))

            if [ -f "$output_file" ]; then
                local file_size
                file_size=$(du -h "$output_file" | awk '{print $1}')
                echo ""
                echo -e "  ${BG_EMERALD}${CLR_WHITE}${CLR_BOLD}  ✓ BUILD SUCCEEDED  ${CLR_RESET} ${CLR_EMERALD}Installable APK created in ${elapsed}s!${CLR_RESET}"
                echo ""
                echo -e "  ${CLR_BOLD}${CLR_WHITE}Artifact Location:${CLR_RESET}"
                echo -e "  ${CLR_CYAN}${CLR_BOLD}$output_file${CLR_RESET}"
                echo -e "  ${CLR_GRAY}File Size: ${CLR_WHITE}${file_size}${CLR_RESET}"
            fi
            ;;
        3)
            print_info "Executing: flutter build apk --debug..."
            flutter build apk --debug \
                --dart-define=API_ENV="$TARGET_ENV" \
                --dart-define=API_BASE_URL="$TARGET_API"

            local output_file="$APP_DIR/build/app/outputs/flutter-apk/app-debug.apk"
            local end_time
            end_time=$(date +%s)
            local elapsed=$((end_time - start_time))

            if [ -f "$output_file" ]; then
                local file_size
                file_size=$(du -h "$output_file" | awk '{print $1}')
                echo ""
                echo -e "  ${BG_EMERALD}${CLR_WHITE}${CLR_BOLD}  ✓ BUILD SUCCEEDED  ${CLR_RESET} ${CLR_EMERALD}Debug APK created in ${elapsed}s!${CLR_RESET}"
                echo ""
                echo -e "  ${CLR_BOLD}${CLR_WHITE}Artifact Location:${CLR_RESET}"
                echo -e "  ${CLR_CYAN}${CLR_BOLD}$output_file${CLR_RESET}"
                echo -e "  ${CLR_GRAY}File Size: ${CLR_WHITE}${file_size}${CLR_RESET}"
            fi
            ;;
    esac
}

# ------------------------------------------------------------------------------
# 6. Quality Checks & Diagnostics
# ------------------------------------------------------------------------------
run_analyze() {
    print_banner
    print_header "FLUTTER STATIC CODE HEALTH ANALYSIS"
    cd "$APP_DIR"
    flutter analyze
    print_success "Zero static analysis issues found!"
}

run_test() {
    print_banner
    print_header "AUTOMATED UNIT & WIDGET TEST RUNNER"
    cd "$APP_DIR"
    flutter test
    print_success "All test suites executed cleanly!"
}

run_clean() {
    print_banner
    print_header "CLEANING BUILD ARTIFACTS & FETCHING DEPENDENCIES"
    cd "$APP_DIR"
    flutter clean
    flutter pub get
    print_success "Project cache cleaned and dependencies refreshed!"
}

run_doctor() {
    print_banner
    print_header "FLUTTER SYSTEM DOCTOR DIAGNOSTICS"
    flutter doctor -v
}

# ------------------------------------------------------------------------------
# 7. Interactive Main Menu
# ------------------------------------------------------------------------------
interactive_menu() {
    print_banner

    # Check live server status for the menu badge
    local status_badge="${CLR_RED}[ OFFLINE ]${CLR_RESET}"
    if fuser "$WEB_PORT/tcp" >/dev/null 2>&1; then
        status_badge="${CLR_EMERALD}[ ONLINE: http://localhost:$WEB_PORT ]${CLR_RESET}"
    fi

    echo -e "  ${CLR_BOLD}${CLR_WHITE}Current Web Simulation Status:${CLR_RESET} $status_badge"
    echo ""
    echo -e "  ${CLR_BOLD}${CLR_WHITE}Web Simulation Controls (19.5:9 Phone Frame):${CLR_RESET}"
    echo -e "    ${CLR_EMERALD}${CLR_BOLD}1)${CLR_RESET} Start Web Simulation  ${CLR_GRAY}(Builds mobile & serves on localhost:1702)${CLR_RESET}"
    echo -e "    ${CLR_RED}${CLR_BOLD}2)${CLR_RESET} Stop Web Simulation   ${CLR_GRAY}(Kills server & completely frees port 1702)${CLR_RESET}"
    echo -e "    ${CLR_CYAN}${CLR_BOLD}3)${CLR_RESET} Check Web Status      ${CLR_GRAY}(Inspect PID, memory, port & recent logs)${CLR_RESET}"
    echo -e "    ${CLR_MINT}${CLR_BOLD}4)${CLR_RESET} View Live Web Logs    ${CLR_GRAY}(Tail real-time access logs)${CLR_RESET}"
    echo ""
    echo -e "  ${CLR_BOLD}${CLR_WHITE}Build Pipeline & Device Operations:${CLR_RESET}"
    echo -e "    ${CLR_GOLD}${CLR_BOLD}5)${CLR_RESET} Build Mobile App      ${CLR_GRAY}(Interactive APK / Play Store AAB Generator)${CLR_RESET}"
    echo -e "    ${CLR_CYAN}${CLR_BOLD}6)${CLR_RESET} Run Code Analyzer     ${CLR_GRAY}(Static linting & zero-warning verification)${CLR_RESET}"
    echo -e "    ${CLR_EMERALD}${CLR_BOLD}7)${CLR_RESET} Run Test Suites       ${CLR_GRAY}(Unit & widget regression tests)${CLR_RESET}"
    echo -e "    ${CLR_WHITE}${CLR_BOLD}8)${CLR_RESET} Clean & Fetch Pub     ${CLR_GRAY}(Reset build cache & dependencies)${CLR_RESET}"
    echo -e "    ${CLR_GRAY}${CLR_BOLD}9)${CLR_RESET} Flutter Doctor        ${CLR_GRAY}(SDK & environment diagnostics)${CLR_RESET}"
    echo ""
    echo -e "    ${CLR_RED}${CLR_BOLD}0) Exit${CLR_RESET}"
    echo ""
    read -rp "  Select an action [0-9]: " menu_choice

    case "$menu_choice" in
        1) start_web ;;
        2) stop_web ;;
        3) status_web ;;
        4) logs_web ;;
        5) build_app ;;
        6) run_analyze ;;
        7) run_test ;;
        8) run_clean ;;
        9) run_doctor ;;
        0)
            echo ""
            echo -e "  ${CLR_EMERALD}Goodbye!${CLR_RESET}"
            exit 0
            ;;
        *)
            print_warn "Invalid option selected."
            sleep 1
            interactive_menu
            ;;
    esac
}

# ------------------------------------------------------------------------------
# 8. Command-Line Argument Dispatcher
# ------------------------------------------------------------------------------
case "${1:-}" in
    start-web)
        start_web "${2:-}"
        ;;
    stop-web)
        stop_web
        ;;
    status-web)
        status_web
        ;;
    logs-web)
        logs_web
        ;;
    build)
        build_app
        ;;
    build-playstore|playstore)
        build_playstore_deployment
        ;;
    analyze)
        run_analyze
        ;;
    test)
        run_test
        ;;
    clean)
        run_clean
        ;;
    doctor)
        run_doctor
        ;;
    *)
        interactive_menu
        ;;
esac
