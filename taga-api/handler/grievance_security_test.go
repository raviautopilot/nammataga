package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"taga-api/model"
)

func setupGrievanceTest(t *testing.T) func() {
	gin.SetMode(gin.TestMode)
	_ = os.MkdirAll("data/grievance", 0755)

	backupFile := "data/grievance/grievanceg_test_backup.json"
	origFile := "data/grievance/grievanceg.json"

	if data, err := os.ReadFile(origFile); err == nil {
		_ = os.WriteFile(backupFile, data, 0644)
	}

	return func() {
		if data, err := os.ReadFile(backupFile); err == nil {
			_ = os.WriteFile(origFile, data, 0644)
			_ = os.Remove(backupFile)
		}
	}
}

func TestGrievance_IDOR_And_UninitializedSliceFix(t *testing.T) {
	cleanup := setupGrievanceTest(t)
	defer cleanup()

	// Seed grievances
	aliceGrievance := model.Grievance{
		ID:            "GRV-ALICE-01",
		MemberName:    "Alice",
		MemberEmail:   "alice@example.com",
		Subject:       "Alice Issue",
		Description:   "Alice private description",
		Status:        "Pending",
		SubmittedDate: time.Now(),
		LastUpdate:    time.Now(),
	}

	bobGrievance := model.Grievance{
		ID:            "GRV-BOB-01",
		MemberName:    "Bob",
		MemberEmail:   "bob@example.com",
		Subject:       "Bob Issue",
		Description:   "Bob private description",
		Status:        "Pending",
		SubmittedDate: time.Now(),
		LastUpdate:    time.Now(),
	}

	seedData, _ := json.MarshalIndent([]model.Grievance{aliceGrievance, bobGrievance}, "", "  ")
	_ = os.WriteFile("data/grievance/grievanceg.json", seedData, 0644)

	// Router setup
	r := gin.New()
	r.GET("/api/grievances", GetGrievances)
	r.GET("/api/grievances/:id", GetGrievanceByID)
	r.PUT("/api/grievances/:id", UpdateGrievance)
	r.DELETE("/api/grievances/:id", DeleteGrievance)

	// --- 1. GET /api/grievances Member Isolation ---
	// Alice should only see Alice's grievance
	reqAlice, _ := http.NewRequest("GET", "/api/grievances", nil)
	wAlice := httptest.NewRecorder()
	cAlice, _ := gin.CreateTestContext(wAlice)
	cAlice.Request = reqAlice
	cAlice.Set("member_email", "alice@example.com")
	cAlice.Set("role", "member")
	GetGrievances(cAlice)

	assert.Equal(t, http.StatusOK, wAlice.Code)
	var aliceList []model.Grievance
	_ = json.Unmarshal(wAlice.Body.Bytes(), &aliceList)
	assert.Len(t, aliceList, 1)
	assert.Equal(t, "GRV-ALICE-01", aliceList[0].ID)

	// Admin should see both grievances
	reqAdmin, _ := http.NewRequest("GET", "/api/grievances", nil)
	wAdmin := httptest.NewRecorder()
	cAdmin, _ := gin.CreateTestContext(wAdmin)
	cAdmin.Request = reqAdmin
	cAdmin.Set("username", "superadmin")
	cAdmin.Set("role", "admin")
	GetGrievances(cAdmin)

	assert.Equal(t, http.StatusOK, wAdmin.Code)
	var adminList []model.Grievance
	_ = json.Unmarshal(wAdmin.Body.Bytes(), &adminList)
	assert.Len(t, adminList, 2)

	// --- 2. GET /api/grievances/:id (Uninitialized Slice Fix & Authorization) ---
	// Alice viewing her own grievance (Previously failed with 404 due to uninitialized slice)
	wAliceGetOwn := httptest.NewRecorder()
	cAliceGetOwn, _ := gin.CreateTestContext(wAliceGetOwn)
	cAliceGetOwn.Request, _ = http.NewRequest("GET", "/api/grievances/GRV-ALICE-01", nil)
	cAliceGetOwn.Params = gin.Params{{Key: "id", Value: "GRV-ALICE-01"}}
	cAliceGetOwn.Set("member_email", "alice@example.com")
	cAliceGetOwn.Set("role", "member")
	GetGrievanceByID(cAliceGetOwn)

	assert.Equal(t, http.StatusOK, wAliceGetOwn.Code, "Alice should be able to view her own grievance from disk")
	var aliceFetched model.Grievance
	_ = json.Unmarshal(wAliceGetOwn.Body.Bytes(), &aliceFetched)
	assert.Equal(t, "GRV-ALICE-01", aliceFetched.ID)

	// Alice viewing Bob's grievance -> 403 Forbidden
	wAliceGetBob := httptest.NewRecorder()
	cAliceGetBob, _ := gin.CreateTestContext(wAliceGetBob)
	cAliceGetBob.Request, _ = http.NewRequest("GET", "/api/grievances/GRV-BOB-01", nil)
	cAliceGetBob.Params = gin.Params{{Key: "id", Value: "GRV-BOB-01"}}
	cAliceGetBob.Set("member_email", "alice@example.com")
	cAliceGetBob.Set("role", "member")
	GetGrievanceByID(cAliceGetBob)

	assert.Equal(t, http.StatusForbidden, wAliceGetBob.Code, "Alice should be forbidden from viewing Bob's grievance")

	// Admin viewing Bob's grievance -> 200 OK
	wAdminGetBob := httptest.NewRecorder()
	cAdminGetBob, _ := gin.CreateTestContext(wAdminGetBob)
	cAdminGetBob.Request, _ = http.NewRequest("GET", "/api/grievances/GRV-BOB-01", nil)
	cAdminGetBob.Params = gin.Params{{Key: "id", Value: "GRV-BOB-01"}}
	cAdminGetBob.Set("username", "admin")
	cAdminGetBob.Set("role", "admin")
	GetGrievanceByID(cAdminGetBob)

	assert.Equal(t, http.StatusOK, wAdminGetBob.Code, "Admin can view any grievance")

	// --- 3. PUT /api/grievances/:id (Update Authorization) ---
	// Alice attempts to update Bob's grievance -> 403 Forbidden
	updateBody, _ := json.Marshal(model.Grievance{Subject: "Hacked Subject"})
	wAlicePutBob := httptest.NewRecorder()
	cAlicePutBob, _ := gin.CreateTestContext(wAlicePutBob)
	cAlicePutBob.Request, _ = http.NewRequest("PUT", "/api/grievances/GRV-BOB-01", bytes.NewBuffer(updateBody))
	cAlicePutBob.Params = gin.Params{{Key: "id", Value: "GRV-BOB-01"}}
	cAlicePutBob.Set("member_email", "alice@example.com")
	cAlicePutBob.Set("role", "member")
	UpdateGrievance(cAlicePutBob)

	assert.Equal(t, http.StatusForbidden, wAlicePutBob.Code, "Alice cannot update Bob's grievance")

	// Admin marks Bob's grievance as Read -> 200 OK
	adminUpdateBody, _ := json.Marshal(model.Grievance{Status: "Read"})
	wAdminPutBob := httptest.NewRecorder()
	cAdminPutBob, _ := gin.CreateTestContext(wAdminPutBob)
	cAdminPutBob.Request, _ = http.NewRequest("PUT", "/api/grievances/GRV-BOB-01", bytes.NewBuffer(adminUpdateBody))
	cAdminPutBob.Params = gin.Params{{Key: "id", Value: "GRV-BOB-01"}}
	cAdminPutBob.Set("username", "admin")
	cAdminPutBob.Set("role", "admin")
	UpdateGrievance(cAdminPutBob)

	assert.Equal(t, http.StatusOK, wAdminPutBob.Code, "Admin can update grievance status")

	// --- 4. DELETE /api/grievances/:id (Delete Authorization & Uninitialized Slice Fix) ---
	// Alice attempts to delete Bob's grievance -> 403 Forbidden
	wAliceDelBob := httptest.NewRecorder()
	cAliceDelBob, _ := gin.CreateTestContext(wAliceDelBob)
	cAliceDelBob.Request, _ = http.NewRequest("DELETE", "/api/grievances/GRV-BOB-01", nil)
	cAliceDelBob.Params = gin.Params{{Key: "id", Value: "GRV-BOB-01"}}
	cAliceDelBob.Set("member_email", "alice@example.com")
	cAliceDelBob.Set("role", "member")
	DeleteGrievance(cAliceDelBob)

	assert.Equal(t, http.StatusForbidden, wAliceDelBob.Code, "Alice cannot delete Bob's grievance")

	// Alice deletes her own grievance -> 200 OK
	wAliceDelOwn := httptest.NewRecorder()
	cAliceDelOwn, _ := gin.CreateTestContext(wAliceDelOwn)
	cAliceDelOwn.Request, _ = http.NewRequest("DELETE", "/api/grievances/GRV-ALICE-01", nil)
	cAliceDelOwn.Params = gin.Params{{Key: "id", Value: "GRV-ALICE-01"}}
	cAliceDelOwn.Set("member_email", "alice@example.com")
	cAliceDelOwn.Set("role", "member")
	DeleteGrievance(cAliceDelOwn)

	assert.Equal(t, http.StatusOK, wAliceDelOwn.Code, "Alice can delete her own grievance")

	_ = filepath.Clean("")
}
