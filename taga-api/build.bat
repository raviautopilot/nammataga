@echo off
setlocal enabledelayedexpansion

REM ---------------- Configuration ----------------
set MAIN_FILE=main.go
set OUTPUT_BINARY=taga-api

REM Force Linux build (same as your script)
set OS=linux

for /f %%i in ('go env GOARCH') do set ARCH=%%i

REM ---------------- Functions ----------------

echo Ensuring module dependencies are up to date...
go mod tidy
go clean -cache

IF %ERRORLEVEL% NEQ 0 (
    echo ❌ Dependency preparation failed!
    exit /b 1
)

echo Running swag init...
swag init ^
  --generalInfo %MAIN_FILE% ^
  --dir .\,handler ^
  --output ./docs

IF %ERRORLEVEL% NEQ 0 (
    echo ❌ Swagger generation failed!
    exit /b 1
)

echo.
echo Building for %OS%/%ARCH%...

REM Build Linux binary (same as bash)
set TARGET=%OUTPUT_BINARY%

set GOOS=%OS%
set GOARCH=%ARCH%

go build -o %TARGET% %MAIN_FILE%

IF %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed!
    exit /b 1
)

echo ========================================
echo ✅ Build completed successfully!
echo Target: %OS%/%ARCH%
echo Binary: %TARGET%
echo ========================================

endlocal