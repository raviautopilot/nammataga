package model

// AboutResponse represents the structure of the about data
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

// StatsResponse represents a single statistics item
// @Description Organization statistics item with label, value, and display properties
type StatsResponse struct {
	ID         int    `json:"id"`
	Label      string `json:"label"`
	Value      int    `json:"value"`
	IconName   string `json:"icon_name"`
	ColorClass string `json:"color_class"`
}

// Objective represents a single objective item
type Objective struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
}

// ServiceResponse represents a single service offered by the organization.
// It's used for the Swagger documentation and JSON response structure.
type ServiceResponse struct {
	ID          int    `json:"id" example:"1"`
	Title       string `json:"title" example:"Professional Development"`
	Description string `json:"description" example:"Training programs, workshops, and skill enhancement opportunities for career growth."`
	IconName    string `json:"icon_name" example:"Users"`
	Category    string `json:"category" example:"member-support"`
}

// Headquarters represents the main office information.
type Headquarters struct {
	Name    string `json:"name" example:"TAGA Headquarters"`
	Address string `json:"address" example:"Tamil Nadu Agricultural University Campus\nCoimbatore - 641003\nTamil Nadu, India"`
}

// RegionalOffice represents a regional office.
type RegionalOffice struct {
	ID      int    `json:"id" example:"1"`
	City    string `json:"city" example:"Chennai"`
	Phone   string `json:"phone" example:"+91-44-2345-6789"`
	Address string `json:"address" example:"123 Agricultural Complex, GST Road, Chennai - 600032"`
}

// ContactResponse represents the full contact information payload.
type ContactResponse struct {
	Headquarters    Headquarters     `json:"headquarters"`
	OfficeHours     string           `json:"office_hours" example:"Monday - Friday: 9:00 AM - 5:30 PM\nSaturday: 9:00 AM - 1:00 PM\nSunday: Closed"`
	PrimaryPhone    string           `json:"primary_phone" example:"+91-422-6611-200"`
	PrimaryEmail    string           `json:"primary_email" example:"info@taga-tn.org"`
	SecondaryEmail  string           `json:"secondary_email" example:"info2@taga-tn.org"`
	Website         string           `json:"website" example:"www.taga-tn.org"`
	RegionalOffices []RegionalOffice `json:"regional_offices"`
}
