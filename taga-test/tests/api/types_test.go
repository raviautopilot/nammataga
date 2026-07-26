package api_tests

// Post models the structure used for the API placeholder mock tests.
type Post struct {
	ID     int    `json:"id,omitempty"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}
