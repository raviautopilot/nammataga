package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// WriteFileAtomic writes data to a temporary file in the same directory and renames it atomically.
func WriteFileAtomic(filePath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	return os.Rename(tmpName, filePath)
}
