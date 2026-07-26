package model

type Document struct {
	Title       string `json:"title"`
	Year        string `json:"year"`
	CategoryID  string `json:"category_id"`           // links to ResourceCategory
	Subcategory string `json:"subcategory,omitempty"` // "Central" or "State", optional
}

type ResourceCategory struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Subcategories []string `json:"subcategories,omitempty"` // e.g., ["Central", "State"]

}
