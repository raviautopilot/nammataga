package api_tests

import (
	"os"
	"testing"

	"e2e-template/tests"
)

func TestMain(m *testing.M) {
	tests.SetupSuite()

	exitCode := m.Run()

	tests.TeardownSuite()

	os.Exit(exitCode)
}
