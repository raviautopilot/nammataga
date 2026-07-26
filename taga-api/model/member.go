package model

type Member struct {
	ID                       string `json:"id,omitempty"`
	TagaID                   string `json:"tagaId,omitempty"`
	Username                 string `json:"username"`
	Password                 string `json:"password"`
	FirstLogin               bool   `json:"first_login"`
	CpsGpfNumber             string `json:"cps_gpf_number"`
	DateOfBirth              string `json:"date_of_birth"`
	Designation              string `json:"designation"`
	EducationalQualification string `json:"educational_qualification"`
	EmailId                  string `json:"emailId"`
	FatherName               string `json:"father_name"`
	Gender                   string `json:"gender"`
	Initial                  string `json:"initial"`
	MobileNumber             string `json:"mobile_number"`
	MotherName               string `json:"mother_name"`
	Name                     string `json:"name"`
	NativeDistrict           string `json:"native_district"`
	PermanentAddress         string `json:"permanent_address"`
	RecruitmentBatch         string `json:"recruitment_batch"`
	ResidentialAddress       string `json:"residential_address"`
	SNo                      int    `json:"sno"`
	SeniorityNumber          string `json:"seniority_number"`
	TbfNumber                string `json:"tbf_number"`
	WorkingDistrict          string `json:"working_district"`
	CreatedAt                string `json:"created_at,omitempty"`
	// Payment fields
	PaymentStatus      string `json:"payment_status,omitempty"`
	SubscriptionActive bool   `json:"subscription_active,omitempty"`
}
