package tests

import (
	"testing"

	"e2e-template/pkg/client"
)

// Post models the structure used for the API placeholder mock tests.
type Post struct {
	ID     int    `json:"id,omitempty"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

func TestAPI_GetPost(t *testing.T) {
	RunAPITest(t, "Get Post Details", func(t *testing.T, c *client.Client) {
		var post Post
		err := c.SendHttpRequest("GET", "/posts/1", nil, nil, &post, nil)
		if err != nil {
			t.Fatalf("Failed to retrieve post: %v", err)
		}

		if post.ID != 1 {
			t.Errorf("Expected ID 1, got %d", post.ID)
		}
		if post.Title == "" {
			t.Errorf("Expected non-empty Title")
		}
	})
}

func TestAPI_CreatePost(t *testing.T) {
	RunAPITest(t, "Create Post Payload", func(t *testing.T, c *client.Client) {
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

func TestAPI_PointerValidation(t *testing.T) {
	RunAPITest(t, "Enforce Pointer Type Safety Check", func(t *testing.T, c *client.Client) {
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
