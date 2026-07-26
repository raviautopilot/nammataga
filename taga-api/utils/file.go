package utils

import (
	"encoding/json"
	"os"
	"taga-api/model"
)

func ReadMembersFromFile(path string) ([]model.Member, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var members []model.Member
	if err := json.Unmarshal(b, &members); err != nil {
		return nil, err
	}

	return members, nil
}
