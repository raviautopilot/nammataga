package api_tests

// Post models the structure used for the API placeholder mock tests.
type Post struct {
	ID     int    `json:"id,omitempty"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

// AboutResponse represents organization about details
type AboutResponse struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Acronym         string `json:"acronym"`
	EstablishedYear int    `json:"established_year"`
	Tagline         string `json:"tagline"`
	HeroImageURL    string `json:"hero_image_url"`
	Mission         string `json:"mission"`
	Vision          string `json:"vision"`
	Description     string `json:"description"`
}

// StatsResponse represents an organization statistic item
type StatsResponse struct {
	ID         int    `json:"id"`
	Label      string `json:"label"`
	Value      int    `json:"value"`
	IconName   string `json:"icon_name"`
	ColorClass string `json:"color_class"`
}

// Objective represents an organization objective item
type Objective struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
}

// ServiceResponse represents an organization service item
type ServiceResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconName    string `json:"icon_name"`
	Category    string `json:"category"`
}

// Headquarters represents the main office address
type Headquarters struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// RegionalOffice represents a regional office contact item
type RegionalOffice struct {
	ID      int    `json:"id"`
	City    string `json:"city"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// ContactResponse represents full contact information
type ContactResponse struct {
	Headquarters    Headquarters     `json:"headquarters"`
	OfficeHours     string           `json:"office_hours"`
	PrimaryPhone    string           `json:"primary_phone"`
	PrimaryEmail    string           `json:"primary_email"`
	SecondaryEmail  string           `json:"secondary_email"`
	Website         string           `json:"website"`
	RegionalOffices []RegionalOffice `json:"regional_offices"`
}

