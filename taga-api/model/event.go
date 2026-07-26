package model

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Attendees   int    `json:"attendees,omitempty"`
	Status      string `json:"status"`
}

type EventItem struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type GalleryImage struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Event    string `json:"event"`
	ImageURL string `json:"imageUrl"`
	Date     string `json:"date"`
	Year     int    `json:"year"`
}
