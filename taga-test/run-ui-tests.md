# E2E UI Test Suite Execution Guide

This document outlines how to compile, execute, configure, and expand the UI automation tests. It includes instructions for running individual test cases, viewing the live browser execution (headed mode), and a complete catalog of all 22 available UI tests with ready-to-use commands.

---

## 1. Prerequisites & Environment Setup

Before running tests, ensure that `chromedriver` is installed on your system and running on port `9515`.

### Linux/Ubuntu Installation:
```bash
sudo apt-get update
sudo apt-get install -y chromium-browser chromium-chromedriver
```

### Start Chromedriver:
Ensure it runs in the background or a separate shell:
```bash
chromedriver --port=9515
```

---

## 2. Test Configuration

The test suite reads configuration from [config.json](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/config.json). You can modify this file directly or override values via environment variables:

| config.json Field | Environment Variable | Default Value | Description |
|---|---|---|---|
| `uiUrl` | `E2E_UI_URL` | `https://nammataga.com` | Target URL of the web frontend application |
| `baseUrl` | `E2E_BASE_URL` | `https://api.nammataga.com` | Target URL of the backend API |
| `seleniumUrl` | `E2E_SELENIUM_URL` | `http://localhost:9515` | Selenium WebDriver address |
| `headless` | `E2E_HEADLESS` | `false` (in config) / `true` (in run-tests.sh) | Toggle headless Chrome UI mode (`false` shows browser) |
| `timeout` | `E2E_TIMEOUT` | `10` | Timeout in seconds for page element waits |

---

## 3. How to Run the Tests

You can run the full suite or individual UI tests using the `run-tests.sh` script, standard `go test` commands, or the pre-compiled test binary.

### Viewing the Browser Execution Live (Headed Mode)

By default, `./run-tests.sh` executes Chrome in **headless mode** (`E2E_HEADLESS=true`).  
To see Chrome open, navigate, click, and execute test steps live on your desktop, set `E2E_HEADLESS=false`:

```bash
# Run an individual test with visible Chrome window (Headed Mode)
E2E_HEADLESS=false ./run-tests.sh -run TestUI_02_AdminAddDeleteMember

# Run all UI tests with visible Chrome window (Headed Mode)
E2E_HEADLESS=false ./run-tests.sh -run TestUI_
```

---

## 4. Individual Test Execution Commands (All 22 UI Test Files)

Below are copy-pasteable commands to execute each of the 22 UI test files individually.

### A. Visible Browser Mode (Headed Mode - Watch Chrome Live)
Use these commands to watch Chrome interact with the web app step-by-step:

```bash
# 01. Public Journeys Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_01_PublicJourneys

# 02. Admin Add & Delete Member Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_02_AdminAddDeleteMember

# 03. Bulk Member Upload Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_03_BulkUploadMembers

# 04. Send Announcement Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_04_SendAnnouncement

# 05. District Office Bearers Management Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_05_DistrictOfficeBearersManagement

# 06. Manage Resources Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_06_ManageResources

# 07. Manage Events Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_07_ManageEvents

# 08. Manage Gallery Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_08_ManageGallery

# 09. Edit Member Details Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_09_EditMemberDetails

# 10. Download Member Reports Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_10_DownloadReports

# 11. Member Login Happy Path Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_11_MemberLoginHappyPath

# 12. Member Annual Subscription Payment Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_12_MemberAnnualSubscriptionPayment

# 13. TAGA Tower Simple Booking Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_13_TAGATower_SimpleBooking

# 14. TAGA Tower Guest Booking Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_14_TAGATower_GuestBooking

# 15. TAGA Tower All Room Booking Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_15_TAGATower_AllRoomBooking

# 16. TAGA Tower Self Multibooking Negative Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_16_TAGATower_SelfMultibookingNegative

# 17. TAGA Tower Gender Restriction (Male) Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_17_TAGATower_GenderRestriction_Male

# 18. TAGA Tower Gender Restriction (Female) Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_18_TAGATower_GenderRestriction_Female

# 19. TAGA Tower Gents Dorm Restriction Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_19_TAGATower_GentsDormRestriction

# 20. TAGA Tower Ladies Dorm Restriction Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_20_TAGATower_LadiesDormRestriction

# 21. TAGA Tower 10-Days Booking Limit Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_21_TAGATower_TenDaysBooking

# 22. TAGA Tower Overlapping Booking Test
E2E_HEADLESS=false ./run-tests.sh -run TestUI_22_TAGATower_OverlappingBooking
```

### B. Headless Mode (Background Execution)
Use these commands to run individual tests headlessly in the background:

```bash
# 01. Public Journeys Test
./run-tests.sh -run TestUI_01_PublicJourneys

# 02. Admin Add & Delete Member Test
./run-tests.sh -run TestUI_02_AdminAddDeleteMember

# 03. Bulk Member Upload Test
./run-tests.sh -run TestUI_03_BulkUploadMembers

# 04. Send Announcement Test
./run-tests.sh -run TestUI_04_SendAnnouncement

# 05. District Office Bearers Management Test
./run-tests.sh -run TestUI_05_DistrictOfficeBearersManagement

# 06. Manage Resources Test
./run-tests.sh -run TestUI_06_ManageResources

# 07. Manage Events Test
./run-tests.sh -run TestUI_07_ManageEvents

# 08. Manage Gallery Test
./run-tests.sh -run TestUI_08_ManageGallery

# 09. Edit Member Details Test
./run-tests.sh -run TestUI_09_EditMemberDetails

# 10. Download Member Reports Test
./run-tests.sh -run TestUI_10_DownloadReports

# 11. Member Login Happy Path Test
./run-tests.sh -run TestUI_11_MemberLoginHappyPath

# 12. Member Annual Subscription Payment Test
./run-tests.sh -run TestUI_12_MemberAnnualSubscriptionPayment

# 13. TAGA Tower Simple Booking Test
./run-tests.sh -run TestUI_13_TAGATower_SimpleBooking

# 14. TAGA Tower Guest Booking Test
./run-tests.sh -run TestUI_14_TAGATower_GuestBooking

# 15. TAGA Tower All Room Booking Test
./run-tests.sh -run TestUI_15_TAGATower_AllRoomBooking

# 16. TAGA Tower Self Multibooking Negative Test
./run-tests.sh -run TestUI_16_TAGATower_SelfMultibookingNegative

# 17. TAGA Tower Gender Restriction (Male) Test
./run-tests.sh -run TestUI_17_TAGATower_GenderRestriction_Male

# 18. TAGA Tower Gender Restriction (Female) Test
./run-tests.sh -run TestUI_18_TAGATower_GenderRestriction_Female

# 19. TAGA Tower Gents Dorm Restriction Test
./run-tests.sh -run TestUI_19_TAGATower_GentsDormRestriction

# 20. TAGA Tower Ladies Dorm Restriction Test
./run-tests.sh -run TestUI_20_TAGATower_LadiesDormRestriction

# 21. TAGA Tower 10-Days Booking Limit Test
./run-tests.sh -run TestUI_21_TAGATower_TenDaysBooking

# 22. TAGA Tower Overlapping Booking Test
./run-tests.sh -run TestUI_22_TAGATower_OverlappingBooking
```

---

## 5. Available UI Tests Catalog Table

Below is the summary reference table for all 22 UI tests in [taga-test/tests/ui/](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/):

| # | Test Function Name | Test File | Description |
|---|---|---|---|
| 01 | `TestUI_01_PublicJourneys` | [01-public-journies_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/01-public-journies_test.go) | Public user navigation across home, office bearers, events, member/admin login pages |
| 02 | `TestUI_02_AdminAddDeleteMember` | [02-admin-add-delete-member_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/02-admin-add-delete-member_test.go) | Admin login, single member creation with 19 fields, and automated member deletion cleanup |
| 03 | `TestUI_03_BulkUploadMembers` | [03-bulk-upload-members_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/03-bulk-upload-members_test.go) | Admin bulk member import from CSV fixture and batch deletion cleanup |
| 04 | `TestUI_04_SendAnnouncement` | [04-send-announcement_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/04-send-announcement_test.go) | Admin announcement broadcast creation and delivery verification |
| 05 | `TestUI_05_DistrictOfficeBearersManagement` | [05-district-office-bearers_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/05-district-office-bearers_test.go) | District office bearer filter, selection, and management workflow |
| 06 | `TestUI_06_ManageResources` | [06-manage-resources_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/06-manage-resources_test.go) | Admin resource document upload (PDF) and category assignment |
| 07 | `TestUI_07_ManageEvents` | [07-manage-events_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/07-manage-events_test.go) | Event creation, date scheduling, location assignment, and event management |
| 08 | `TestUI_08_ManageGallery` | [08-manage-gallery_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/08-manage-gallery_test.go) | Admin photo gallery upload and album management |
| 09 | `TestUI_09_EditMemberDetails` | [09-edit-member_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/09-edit-member_test.go) | Admin editing existing member designation, district details, and verifying saved state |
| 10 | `TestUI_10_DownloadReports` | [10-download-reports_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/10-download-reports_test.go) | Admin panel member reports filter and CSV export download |
| 11 | `TestUI_11_MemberLoginHappyPath` | [11-member-login_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/11-member-login_test.go) | Member credential login, profile dashboard access, and logout |
| 12 | `TestUI_12_MemberAnnualSubscriptionPayment` | [12-member-payment_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/12-member-payment_test.go) | Member subscription payment workflow and payment status update verification |
| 13 | `TestUI_13_TAGATower_SimpleBooking` | [13-tagatower-simple-booking_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/13-tagatower-simple-booking_test.go) | TAGA Tower single room booking workflow and confirmation |
| 14 | `TestUI_14_TAGATower_GuestBooking` | [14-tagatower-guest-booking_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/14-tagatower-guest-booking_test.go) | TAGA Tower guest room booking workflow with guest details |
| 15 | `TestUI_15_TAGATower_AllRoomBooking` | [15-tagatower-allroombooking_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/15-tagatower-allroombooking_test.go) | TAGA Tower booking across multiple room types |
| 16 | `TestUI_16_TAGATower_SelfMultibookingNegative` | [16-tagatower-self-multibooking-negative_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/16-tagatower-self-multibooking-negative_test.go) | Negative validation: preventing duplicate self-bookings for overlapping dates |
| 17 | `TestUI_17_TAGATower_GenderRestriction_Male` | [17-tagatower-gender-restriction-male_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/17-tagatower-gender-restriction-male_test.go) | Gender restriction validation: Male booking rules enforcement |
| 18 | `TestUI_18_TAGATower_GenderRestriction_Female` | [18-tagatower-gender-restriction-female_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/18-tagatower-gender-restriction-female_test.go) | Gender restriction validation: Female booking rules enforcement |
| 19 | `TestUI_19_TAGATower_GentsDormRestriction` | [19-tagatower-gentsdorm-restriction_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/19-tagatower-gentsdorm-restriction_test.go) | Dormitory restriction validation: Gents dorm access control |
| 20 | `TestUI_20_TAGATower_LadiesDormRestriction` | [20-tagatower-ladiesdorm-restriction_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/20-tagatower-ladiesdorm-restriction_test.go) | Dormitory restriction validation: Ladies dorm access control |
| 21 | `TestUI_21_TAGATower_TenDaysBooking` | [21-tagatower-tendaysbooking_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/21-tagatower-tendaysbooking_test.go) | Maximum stay duration validation: 10-day booking limit enforcement |
| 22 | `TestUI_22_TAGATower_OverlappingBooking` | [22-tagatower-overlappingbooking_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/22-tagatower-overlappingbooking_test.go) | Date overlap validation: preventing double booking of identical rooms |
| 23 | `TestUI_23_TAGATower_IncompleteGuestDetails` | [23-tagatower-incomplete-guest-details_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/23-tagatower-incomplete-guest-details_test.go) | Guest details validation: required fields enforcement |
| 28 | `TestUI_28_TAGATower_MixedGenderCoupleBooking` | [28-tagatower-mixed-gender-couple-booking_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/28-tagatower-mixed-gender-couple-booking_test.go) | Mixed gender couple booking passes & blocks 3rd bed in Apex Suite |

---

## 6. Pointer-Based UI Test Architecture

To make UI test scripts easily readable and editable, tests implement a **declarative, pointer-passing style** (e.g. in [01-public-journies_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/01-public-journies_test.go)).

### Step-by-Step Scenario Example:
```go
func TestUI_01_PublicJourneys(t *testing.T) {
	tests.RunUITest(t, "Verify Public User Journeys", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		
		// 1. Initialize Persona & Result pointers
		pubPersona := actions.NewPublicPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_01_PublicJourneys")

		// 2. Simple sequential action flows
		actions.GoToHome(pubPersona, result)
		actions.GoToOfficeBeaers(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToEvents(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToMemberLogin(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToAdminLogin(pubPersona, result)

		// 3. Evaluate results
		if result.Failed() {
			t.Fatalf("Journey failed! Error: %v. Advice: %v", result.Error, result.Advice)
		}
	})
}
```

### How Evidence & Advice are Captured:
Each action function (e.g. `GoToOfficeBeaers`) performs the following under the hood:
1. **Actions History**: Appends the action step name to `result.Actions`.
2. **Execution Check**: Instantly skips execution if a previous step in the journey has already failed.
3. **Defensive Validation**: Ensures the page driver and elements are ready.
4. **Remediation & Advice**: If a step fails, it appends friendly debugging tips to `result.Advice` (e.g., verifying if the button's testID is present).
5. **Screenshots & Evidence**: Captures and saves screenshots to the active evidence path, attaching them to `result.Evidence` to document successes or failures.

---

## 7. Test Reports & Artifacts

Every test execution generates reports and collects evidence in the `evidence/run-YYYY-MM-DD_HH-MM-SS/` directory.

* **HTML Visual Dashboard**: Open `evidence/run-.../reports/report.html` in any web browser to see detailed pass/fail status, execution duration, and embedded screenshots.
* **JSON Results**: Parsed test data is stored in `evidence/run-.../reports/report.json`.
* **Markdown Summary**: A quick test report summary is written to `evidence/test-report.md`.
* **Screenshots Directory**: All screenshots captured by tests (successes and failures) are located under `evidence/run-.../screenshots/`.
