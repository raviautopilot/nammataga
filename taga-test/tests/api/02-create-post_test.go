package api_tests

import (
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_02_CreatePost(t *testing.T) {
	tests.RunAPITest(t, "Create Post Payload", func(t *testing.T, c *client.Client) {
		newPost := &Post{
			Title:  "Go E2E Framework Test",
			Body:   "Verifying API custom wrapper with pointer and JSON logging",
			UserID: 99,
		}
		var createdPost Post

		// Setup custom Bearer auth token injection
		auth := &client.BearerTokenAuth{Token: "super-secret-e2e-token"}

		err := c.SendHttpRequest("POST", "/posts", nil, newPost, &createdPost, auth)
		if err != nil {
			t.Fatalf("Failed to create post: %v", err)
		}

		if createdPost.Title != newPost.Title {
			t.Errorf("Expected Title '%s', got '%s'", newPost.Title, createdPost.Title)
		}
		if createdPost.ID == 0 {
			t.Errorf("Expected returned ID to be populated")
		}
	})
}
