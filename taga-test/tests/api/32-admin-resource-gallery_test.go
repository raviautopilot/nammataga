package api_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type GalleryImageUploadResponse struct {
	Message string `json:"message"`
	Image   struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"image"`
}

func TestAPI_AdminContent_Uploads(t *testing.T) {
	var adminToken string
	var savedClient *client.Client
	var uploadedImageID string

	// Step 1: Admin uploads a PDF document resource (Multipart Form)
	tests.RunAPITestWithDetails(t, "[Admin] POST Upload Resource Document", "Uploads a new PDF resource under the establishment category.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		savedClient = tctx.Client
		adminToken = getValidAdminToken(tctx.T, tctx.Client)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		// Form fields
		_ = writer.WriteField("categoryId", "establishment")
		_ = writer.WriteField("title", "E2E Temp Document")
		_ = writer.WriteField("year", "2026")
		_ = writer.WriteField("subcategory", "E2E Test")

		// File part
		part, err := writer.CreateFormFile("file", "test-document.pdf")
		if err != nil {
			tctx.Fatalf("Failed to create form file: %v", err)
		}
		part.Write([]byte("%PDF-1.4 dummy pdf content for E2E validation"))
		writer.Close()

		reqUrl := tctx.Client.BaseURL + "/api/admin/resources/upload"
		req, err := http.NewRequest("POST", reqUrl, &body)
		if err != nil {
			tctx.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Failed to send upload request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %d", resp.StatusCode)
			tctx.Fatalf("Expected 200 OK, got: %d", resp.StatusCode)
		}
		tctx.Actual = "HTTP 200 OK, PDF document uploaded successfully"
	})

	// Step 2: Admin uploads a gallery image (Multipart Form)
	tests.RunAPITestWithDetails(t, "[Admin] POST Upload Gallery Image", "Uploads a photo to the public gallery.", "HTTP 200 OK containing image metadata", func(tctx *tests.TestContext) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		// Form fields
		_ = writer.WriteField("title", "E2E Temp Gallery Photo")
		_ = writer.WriteField("event", "E2E Gala")
		_ = writer.WriteField("date", "2026-08-11")
		_ = writer.WriteField("year", "2026")

		// Image file
		part, err := writer.CreateFormFile("image", "gala.jpg")
		if err != nil {
			tctx.Fatalf("Failed to create image field: %v", err)
		}
		part.Write([]byte("fake jpeg binary bytes"))
		writer.Close()

		reqUrl := tctx.Client.BaseURL + "/api/admin/gallery/upload"
		req, err := http.NewRequest("POST", reqUrl, &body)
		if err != nil {
			tctx.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Failed to send image upload: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %d", resp.StatusCode)
			tctx.Fatalf("Expected 200 OK, got: %d", resp.StatusCode)
		}

		var uploadResp GalleryImageUploadResponse
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			tctx.Fatalf("Failed to parse gallery response: %v", err)
		}

		uploadedImageID = uploadResp.Image.ID
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Image ID='%s'", uploadedImageID)
	})

	// Cleanup at the end: Delete the resource and the gallery photo we created
	t.Cleanup(func() {
		if savedClient != nil && adminToken != "" {
			auth := &client.BearerTokenAuth{Token: adminToken}
			var resp map[string]interface{}

			// Delete document resource
			_ = savedClient.SendHttpRequest("DELETE", "/api/admin/resources/establishment/E2E Temp Document", nil, nil, &resp, auth)

			// Delete gallery image
			if uploadedImageID != "" {
				_ = savedClient.SendHttpRequest("DELETE", "/api/admin/gallery/"+uploadedImageID, nil, nil, &resp, auth)
			}
		}
	})
}

func TestAPI_AdminContent_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Admin] POST Upload Resource - Missing File",
			Description: "Attempts to upload a resource without attaching a file.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				adminToken := getValidAdminToken(tc.T, tc.Client)
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				_ = writer.WriteField("categoryId", "establishment")
				_ = writer.WriteField("title", "Missing File Doc")
				writer.Close()

				reqUrl := tc.Client.BaseURL + "/api/admin/resources/upload"
				req, _ := http.NewRequest("POST", reqUrl, &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+adminToken)

				resp, err := tc.Client.HTTPClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "Expected failure due to missing file"
					tc.Fatalf("Expected 400 Bad Request but got 200 OK")
				}
				tc.Actual = "Correctly rejected missing file"
			},
		},
		{
			Name:        "[Public] POST Upload Gallery Image - Unauthenticated",
			Description: "Attempts to upload gallery image without auth token.",
			Expected:    "HTTP 401 Unauthorized",
			TestFn: func(tc *tests.TestContext) {
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake jpeg binary bytes"))
				writer.Close()

				reqUrl := tc.Client.BaseURL + "/api/admin/gallery/upload"
				req, _ := http.NewRequest("POST", reqUrl, &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				resp, err := tc.Client.HTTPClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusUnauthorized {
					tc.FailureReason = fmt.Sprintf("Expected 401, got %d", resp.StatusCode)
					tc.Fatalf("Expected 401 Unauthorized")
				}
				tc.Actual = "Correctly blocked unauthenticated upload"
			},
		},
		{
			Name:        "[Admin] POST Upload Resource - 0 Byte File",
			Description: "Attempts to upload a 0-byte file to verify storage boundary checks.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				adminToken := getValidAdminToken(tc.T, tc.Client)
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				_ = writer.WriteField("categoryId", "establishment")
				_ = writer.WriteField("title", "Zero Byte Doc")
				
				part, _ := writer.CreateFormFile("file", "zerobyte.pdf")
				part.Write([]byte{}) // 0 bytes
				writer.Close()

				reqUrl := tc.Client.BaseURL + "/api/admin/resources/upload"
				req, _ := http.NewRequest("POST", reqUrl, &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+adminToken)

				resp, err := tc.Client.HTTPClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "Expected failure due to 0-byte file"
					tc.Fatalf("Expected 400 Bad Request but got 200 OK")
				}
				tc.Actual = "Correctly rejected 0-byte file"
			},
		},
		{
			Name:        "[Admin] POST Upload Resource - Mismatched Mimetype",
			Description: "Attempts to upload an executable disguised as a PDF.",
			Expected:    "HTTP 400 Bad Request or HTTP 415 Unsupported Media Type",
			TestFn: func(tc *tests.TestContext) {
				adminToken := getValidAdminToken(tc.T, tc.Client)
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				_ = writer.WriteField("categoryId", "establishment")
				_ = writer.WriteField("title", "Fake PDF")
				
				part, _ := writer.CreateFormFile("file", "fake.pdf")
				part.Write([]byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xFF\xFF\x00\x00")) // MZ header
				writer.Close()

				reqUrl := tc.Client.BaseURL + "/api/admin/resources/upload"
				req, _ := http.NewRequest("POST", reqUrl, &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+adminToken)

				resp, err := tc.Client.HTTPClient.Do(req)
				if err != nil {
					tc.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					tc.FailureReason = "Expected failure due to mismatched mimetype"
					tc.Fatalf("Expected 400 Bad Request but got 200 OK")
				}
				tc.Actual = "Correctly rejected mismatched mimetype"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
