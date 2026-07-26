package api_tests

import (
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_03_PointerValidation(t *testing.T) {
	tests.RunAPITest(t, "Enforce Pointer Type Safety Check", func(t *testing.T, c *client.Client) {
		valuePost := Post{
			Title: "Passed by value",
		}
		var resp Post

		// Intentionally pass a value struct instead of pointer for request
		err := c.SendHttpRequest("POST", "/posts", nil, valuePost, &resp, nil)
		if err == nil {
			t.Fatalf("Expected SendHttpRequest to fail on value struct argument")
		}
		expectedErr := "HTTP Error: status=0, err=request body must be a pointer to a struct/value"
		if err.Error() != expectedErr {
			t.Errorf("Expected validation error '%s', got '%s'", expectedErr, err.Error())
		}
	})
}
