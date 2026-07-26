package model

type StateOfficer struct {
	Sno           int    `json:"sno"`
	Name          string `json:"name"`
	Designation   string `json:"designation"`
	Phone         string `json:"phone"`
	Location      string `json:"location"`
	Qualification string `json:"qualification"`
	Experience    int    `json:"experience"`
	Description   string `json:"description"`
	Image         string `json:"image,omitempty"` // optional
}

type DistrictOfficer struct {
	Name       string `json:"name"`
	Title      string `json:"title"`
	Contact    string `json:"contact"`
	Department string `json:"department"`
}

type OfficeData struct {
	StateOfficers    []StateOfficer               `json:"state_officers"`
	DistrictOfficers map[string][]DistrictOfficer `json:"district_officers"`
}
