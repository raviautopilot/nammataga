package utils_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taga-api/utils"
)

func TestWriteFileAtomic_BasicSuccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atomic_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	targetPath := filepath.Join(tmpDir, "sub", "target.json")
	content := []byte(`{"status":"success","data":[1,2,3]}`)

	err = utils.WriteFileAtomic(targetPath, content, 0644)
	require.NoError(t, err)

	readData, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, content, readData)

	// Verify no stray .tmp files left in directory
	entries, err := os.ReadDir(filepath.Dir(targetPath))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "target.json", entries[0].Name())
}

func TestWriteFileAtomic_ConcurrentWritesWithMutex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atomic_concurrent_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	targetPath := filepath.Join(tmpDir, "concurrent.txt")
	var mu sync.Mutex
	var wg sync.WaitGroup

	numGoroutines := 20
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			payload := []byte(fmt.Sprintf("Goroutine-%d-Payload", idx))
			writeErr := utils.WriteFileAtomic(targetPath, payload, 0644)
			assert.NoError(t, writeErr)
		}(i)
	}

	wg.Wait()

	// Read final contents, must not be corrupted
	finalData, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Contains(t, string(finalData), "Goroutine-")
	assert.Contains(t, string(finalData), "-Payload")

	// Ensure no temp files leaked
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, filepath.Ext(entry.Name()) == ".tmp", "Temp file was not cleaned up: %s", entry.Name())
	}
}
