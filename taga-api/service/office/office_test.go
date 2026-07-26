package office

import (
	"encoding/json"
	"os"
	"path/filepath"
	"taga-api/model"
	"testing"
)

func TestGetOfficeData(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()

	// Set up test data using model.OfficeData
	testData := model.OfficeData{
		StateOfficers: []model.StateOfficer{
			{
				Sno:           1,
				Name:          "Test Name",
				Designation:   "State President",
				Phone:         "123-456-7890",
				Location:      "Test Location",
				Qualification: "M.Sc Agriculture",
				Experience:    10,
				Description:   "Test description",
			},
		},
		DistrictOfficers: map[string][]model.DistrictOfficer{
			"test": {
				{
					Name:       "District Officer",
					Title:      "Test Title",
					Contact:    "123-456-7890",
					Department: "Test Department",
				},
			},
		},
	}

	// Write test data to a temporary file
	expectedDir := filepath.Join(tempDir, "data", "office")
	os.MkdirAll(expectedDir, 0755)
	testFilePath := filepath.Join(expectedDir, "taga-office.json")

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(testFilePath, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Temporarily change the working directory to use our temp data directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	os.Chdir(tempDir)

	// Test cases
	tests := []struct {
		name       string
		pathParam  string
		wantErr    bool
		wantLength int
	}{
		{
			name:       "Valid state",
			pathParam:  "state",
			wantErr:    false,
			wantLength: 1,
		},
		{
			name:       "Valid district",
			pathParam:  "test",
			wantErr:    false,
			wantLength: 1,
		},
		{
			name:      "Non-existent district",
			pathParam: "nonexistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetOfficeData(tt.pathParam)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOfficeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check the type of result
				switch v := result.(type) {
				case []model.StateOfficer:
					if len(v) != tt.wantLength {
						t.Errorf("GetOfficeData() length = %v, want %v", len(v), tt.wantLength)
					}
				case []model.DistrictOfficer:
					if len(v) != tt.wantLength {
						t.Errorf("GetOfficeData() length = %v, want %v", len(v), tt.wantLength)
					}
				default:
					t.Errorf("GetOfficeData() returned unexpected type")
				}
			}
		})
	}
}
