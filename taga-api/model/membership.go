package model

type Membership struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	DateOfBirth       string `json:"dateOfBirth"`
	Gender            string `json:"gender"`
	Address           string `json:"address"`
	District          string `json:"district"`
	Qualification     string `json:"qualification"`
	GraduationYear    string `json:"graduationYear"`
	CurrentEmployment string `json:"currentEmployment"`
	Organization      string `json:"organization"`
	Experience        string `json:"experience"`
	Specialization    string `json:"specialization"`
	References        string `json:"references"`
}
