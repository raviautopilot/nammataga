package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// Credentials holds a username/password pair.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// MemberFormData holds all input field data for creating a new member.
type MemberFormData struct {
	TagaID                    string `json:"tagaId"`
	Name                      string `json:"name"`
	Initial                   string `json:"initial"`
	Gender                    string `json:"gender"`
	FatherName                string `json:"fatherName"`
	MotherName                string `json:"motherName"`
	EducationalQualification string `json:"educationalQualification"`
	Designation               string `json:"designation"`
	WorkingDistrict           string `json:"workingDistrict"`
	NativeDistrict            string `json:"nativeDistrict"`
	RecruitmentBatch          string `json:"recruitmentBatch"`
	SeniorityNumber           string `json:"seniorityNumber"`
	DateOfBirth               string `json:"dateOfBirth"`
	MobileNumber              string `json:"mobileNumber"`
	Email                     string `json:"email"`
	TbfNumber                 string `json:"tbfNumber"`
	CpsGpfNumber              string `json:"cpsGpfNumber"`
	ResidentialAddress        string `json:"residentialAddress"`
	PermanentAddress          string `json:"permanentAddress"`
}

// Config holds the configuration values for the testing framework.
type Config struct {
	BaseURL                                   string         `json:"baseUrl"`
	UiURL                                     string         `json:"uiUrl"`
	SeleniumURL                               string         `json:"seleniumUrl"`
	Headless                                  bool           `json:"headless"`
	Timeout                                   int            `json:"timeout"`
	AdminLoginButtonTestID                    string         `json:"adminLoginButtonTestID"`
	MemberLoginButtonTestID                   string         `json:"memberLoginButtonTestID"`
	AdminLoginTestIDs                         []string       `json:"adminLoginTestIDs"`
	MemberLoginTestIDs                        []string       `json:"memberLoginTestIDs"`
	AdminCredentials                          Credentials    `json:"adminCredentials"`
	MemberCredentials                         Credentials    `json:"memberCredentials"`
	AdminLoginUsernameInputTestID             string         `json:"adminLoginUsernameInputTestID"`
	AdminLoginPasswordInputTestID             string         `json:"adminLoginPasswordInputTestID"`
	AdminLoginSubmitButtonTestID              string         `json:"adminLoginSubmitButtonTestID"`
	MemberLoginUsernameInputTestID            string         `json:"memberLoginUsernameInputTestID"`
	MemberLoginPasswordInputTestID            string         `json:"memberLoginPasswordInputTestID"`
	MemberLoginSubmitButtonTestID             string         `json:"memberLoginSubmitButtonTestID"`
	LogoutButtonTestID                        string         `json:"logoutButtonTestID"`
	NewMemberEmail                            string         `json:"newMemberEmail"`
	NewMemberFormData                         MemberFormData `json:"newMemberFormData"`
	AdminAddMemberButtonTestID                 string         `json:"adminAddMemberButtonTestID"`
	AdminAddMemberTagaIdInputTestID           string         `json:"adminAddMemberTagaIdInputTestID"`
	AdminAddMemberNameInputTestID             string         `json:"adminAddMemberNameInputTestID"`
	AdminAddMemberInitialInputTestID          string         `json:"adminAddMemberInitialInputTestID"`
	AdminAddMemberGenderSelectTestID          string         `json:"adminAddMemberGenderSelectTestID"`
	AdminAddMemberFatherNameInputTestID       string         `json:"adminAddMemberFatherNameInputTestID"`
	AdminAddMemberMotherNameInputTestID       string         `json:"adminAddMemberMotherNameInputTestID"`
	AdminAddMemberEduQualInputTestID          string         `json:"adminAddMemberEduQualInputTestID"`
	AdminAddMemberDesignationInputTestID      string         `json:"adminAddMemberDesignationInputTestID"`
	AdminAddMemberWorkingDistrictSelectTestID string         `json:"adminAddMemberWorkingDistrictSelectTestID"`
	AdminAddMemberNativeDistrictSelectTestID  string         `json:"adminAddMemberNativeDistrictSelectTestID"`
	AdminAddMemberRecruitmentBatchInputTestID string         `json:"adminAddMemberRecruitmentBatchInputTestID"`
	AdminAddMemberSeniorityNumInputTestID     string         `json:"adminAddMemberSeniorityNumInputTestID"`
	AdminAddMemberDobInputTestID              string         `json:"adminAddMemberDobInputTestID"`
	AdminAddMemberMobileInputTestID           string         `json:"adminAddMemberMobileInputTestID"`
	AdminAddMemberEmailInputTestID            string         `json:"adminAddMemberEmailInputTestID"`
	AdminAddMemberTbfNumInputTestID           string         `json:"adminAddMemberTbfNumInputTestID"`
	AdminAddMemberCpsGpfNumInputTestID        string         `json:"adminAddMemberCpsGpfNumInputTestID"`
	AdminAddMemberResAddressInputTestID       string         `json:"adminAddMemberResAddressInputTestID"`
	AdminAddMemberPermAddressInputTestID      string         `json:"adminAddMemberPermAddressInputTestID"`
	AdminAddMemberSubmitButtonTestID          string         `json:"adminAddMemberSubmitButtonTestID"`
	MemberSearchInputTestID                    string         `json:"memberSearchInputTestID"`
	MemberRefreshButtonTestID                 string         `json:"memberRefreshButtonTestID"`
	MemberDeleteButtonTestID                  string         `json:"memberDeleteButtonTestID"`
	MemberConfirmDeleteButtonTestID           string         `json:"memberConfirmDeleteButtonTestID"`
	AdminPanelButtonTestID                    string         `json:"adminPanelButtonTestID"`
	AdminBulkUploadButtonTestID               string         `json:"adminBulkUploadButtonTestID"`
	AdminBulkUploadFileInputTestID            string         `json:"adminBulkUploadFileInputTestID"`
	AdminBulkUploadSubmitButtonTestID         string         `json:"adminBulkUploadSubmitButtonTestID"`
	BulkMemberEmails                          []string       `json:"bulkMemberEmails"`
	BulkMemberMobiles                         []string       `json:"bulkMemberMobiles"`
	AdminSendAnnouncementButtonTestID         string         `json:"adminSendAnnouncementButtonTestID"`
	AdminAnnouncementTitleInputTestID         string         `json:"adminAnnouncementTitleInputTestID"`
	AdminAnnouncementMessageInputTestID       string         `json:"adminAnnouncementMessageInputTestID"`
	AdminAnnouncementPrioritySelectTestID     string         `json:"adminAnnouncementPrioritySelectTestID"`
	AdminAnnouncementSendToSelectTestID       string         `json:"adminAnnouncementSendToSelectTestID"`
	AdminAnnouncementSubmitButtonTestID       string         `json:"adminAnnouncementSubmitButtonTestID"`
}

// LoadConfig reads the configuration file from path and applies environment overrides.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		BaseURL:     "https://api.nammataga.com",
		UiURL:       "https://nammataga.com",
		SeleniumURL: "http://localhost:9515",
		Headless:    false,
		Timeout:     10,
	}

	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
	}

	// Environment overrides
	if val := os.Getenv("E2E_BASE_URL"); val != "" {
		cfg.BaseURL = val
	}
	if val := os.Getenv("E2E_UI_URL"); val != "" {
		cfg.UiURL = val
	}
	if val := os.Getenv("E2E_SELENIUM_URL"); val != "" {
		cfg.SeleniumURL = val
	}
	if val := os.Getenv("E2E_HEADLESS"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			cfg.Headless = boolVal
		}
	}
	if val := os.Getenv("E2E_TIMEOUT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.Timeout = intVal
		}
	}

	return cfg, nil
}
