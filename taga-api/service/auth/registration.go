package auth

import (
	"fmt"
	"taga-api/model"
	"taga-api/service/email"
	"taga-api/service/member"
	"taga-api/utils"
)

func ProcessRegistrationFile(path string) model.RegistrationResult {
	members, err := utils.ReadMembersFromFile(path)
	var result model.RegistrationResult

	if err != nil {
		result.Failed = append(result.Failed, model.RegistrationError{
			Username: "FILE",
			Errors:   []string{err.Error()},
		})
		return result
	}

	for _, m := range members {
		validationErrors := utils.ValidateMember(m)
		if len(validationErrors) > 0 {
			result.Failed = append(result.Failed, model.RegistrationError{
				Username: m.Username,
				Errors:   validationErrors,
			})
			continue
		}

		tempPassword := utils.GenerateTempPassword()
		m.Password = utils.HashPassword(tempPassword)
		m.FirstLogin = true
		m.PaymentStatus = "Unpaid"
		m.SubscriptionActive = false

		fmt.Println("Generated Password:", tempPassword)

		if err := member.SaveMember(m); err != nil {
			result.Failed = append(result.Failed, model.RegistrationError{
				Username: m.Username,
				Errors:   []string{err.Error()},
			})
			continue
		}

		email.SendRegistrationSuccess(m.EmailId, tempPassword)
		result.SuccessCount++
	}

	if len(result.Failed) > 0 {
		email.SendRegistrationError(result.Failed)
	}

	return result
}

// Delete the old SaveMember function from this file
