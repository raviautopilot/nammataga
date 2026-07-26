package service

import (
	"encoding/json"
	"os"
	"taga-api/model"
	"taga-api/utils"
)

var membershipFile = "data/memberlogin/membership.json"
var districtFile = "data/memberlogin/membershipdistricts.json"

func ApplyMembership(newMember model.Membership) error {
	file, _ := os.ReadFile(membershipFile)

	var members []model.Membership
	json.Unmarshal(file, &members)

	members = append(members, newMember)

	updated, _ := json.MarshalIndent(members, "", "  ")
	return os.WriteFile(membershipFile, updated, 0644)
}
func GetMembershipList() ([]model.Membership, error) {
	file, err := os.ReadFile("data/memberlogin/membership.json")
	if err != nil {
		return nil, err
	}

	var members []model.Membership
	json.Unmarshal(file, &members)

	return members, nil
}
func GetMembershipDistricts() ([]string, error) {
	var districts []string

	err := utils.ReadJSON(districtFile, &districts)
	if err != nil {
		return nil, err
	}

	return districts, nil
}
