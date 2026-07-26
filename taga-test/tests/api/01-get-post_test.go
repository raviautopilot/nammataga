package api_tests

import (
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_01_GetPost(t *testing.T) {
	tests.RunAPITest(t, "Get Post Details", func(t *testing.T, c *client.Client) {
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
