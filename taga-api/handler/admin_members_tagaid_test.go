package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
)

func TestAddMember_TagaIDValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup temp test environment
	tmpDir := t.TempDir()
	membersFile := filepath.Join(tmpDir, "members.json")

	initialMembers := []map[string]interface{}{
		{
			"id":            "member-uid-1",
			"tagaId":        "TAGA001",
			"emailId":       "existing@taga.com",
			"mobile_number": "9876543210",
			"name":          "Existing Member",
		},
	}
	data, _ := json.Marshal(initialMembers)
	_ = os.WriteFile(membersFile, data, 0644)

	oldFile := config.Config.MembersFile
	config.Config.MembersFile = membersFile
	defer func() {
		config.Config.MembersFile = oldFile
	}()

	r := gin.New()
	r.POST("/api/admin/members/add", AddMember)

	t.Run("Rejects creation with duplicate tagaId", func(t *testing.T) {
		payload := AddMemberRequest{
			TagaID:                   "TAGA001", // duplicate
			Name:                     "New User",
			Initial:                  "S",
			Gender:                   "Male",
			EducationalQualification: "B.Tech",
			Designation:              "Engineer",
			WorkingDistrict:          "Chennai",
			NativeDistrict:           "Madurai",
			DateOfBirth:              "1990-01-01",
			Email:                    "newuser@taga.com",
			MobileNumber:             "9876543211",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/admin/members/add", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "TAGA ID already registered", resp["error"])
	})

	t.Run("Rejects creation with duplicate tagaId case-insensitively", func(t *testing.T) {
		payload := AddMemberRequest{
			TagaID:                   "taga001", // lower-case duplicate
			Name:                     "New User 2",
			Initial:                  "M",
			Gender:                   "Female",
			EducationalQualification: "M.Sc",
			Designation:              "Specialist",
			WorkingDistrict:          "Coimbatore",
			NativeDistrict:           "Salem",
			DateOfBirth:              "1992-05-15",
			Email:                    "newuser2@taga.com",
			MobileNumber:             "9876543212",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/admin/members/add", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "TAGA ID already registered", resp["error"])
	})
}
