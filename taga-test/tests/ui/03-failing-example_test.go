package ui_tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

func TestUI_03_FailingExample(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Mock Target Content</h1></body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tests.RunUITest(t, "Intentionally Failing UI Test", func(t *testing.T, page *ui.Page) {
		if err := page.Driver.Get(server.URL); err != nil {
			t.Fatalf("Failed to navigate to mock url: %v", err)
		}

		// Try to verify header text
		text, err := page.GetText("css:h1", 3*time.Second)
		if err != nil {
			t.Fatalf("Failed to get header text: %v", err)
		}

		// Assertion will fail, triggering the deferred failure screenshot handler
		if text != "Expected Correct Header Text" {
			t.Errorf("Header validation failed: expected 'Expected Correct Header Text', got '%s'", text)
		}
	})
}
