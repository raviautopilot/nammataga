package utils

import (
	"taga-api/model"
)

func ValidateMember(m model.Member) []string {
	var errs []string

	if m.Username == "" {
		errs = append(errs, "username missing")
	}
	if m.EmailId == "" {
		errs = append(errs, "emailId missing")
	}
	if m.CpsGpfNumber == "" {
		errs = append(errs, "cpsGpfNumber missing")
	}

	return errs
}


