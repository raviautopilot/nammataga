package tests

import (
	"os"
	"testing"

	"e2e-template/pkg/logger"
)

func TestMain(m *testing.M) {
	SetupSuite()

	logger.Info("Starting E2E Test Suite...")
	logger.Info("Test evidence will be compiled inside: %s", EvidenceDir)

	// Run all tests in the package
	exitCode := m.Run()

	TeardownSuite()

	os.Exit(exitCode)
}
