package model

type EditRequest struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	MobileNumber       string `json:"mobileNumber"`
	MailId             string `json:"mailId"`
	Designation        string `json:"designation"`
	WorkingDistrict    string `json:"workingDistrict"`
	ResidentialAddress string `json:"residentialAddress"`
	PermanentAddress   string `json:"permanentAddress"`
	Remarks            string `json:"remarks"`
	Status             string `json:"status"` // pending / approved / rejected
	CreatedAt          string `json:"createdAt"`
}
