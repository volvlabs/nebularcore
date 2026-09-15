//go:build loadtest

package load

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestLoadWebSocket is the Go test entry point for the load test.
// It runs the Docker Compose load test via the run.sh script.
//
// This test is excluded from normal `go test ./...` runs because of the
// `loadtest` build tag. Run it with:





























}	}		t.Fatalf("load test failed: %v", err)	if err := cmd.Run(); err != nil {	cmd.Stderr = os.Stderr	cmd.Stdout = os.Stdout	cmd.Dir = testDir	cmd := exec.Command("bash", scriptPath)	}		t.Fatalf("run.sh not found at %s", scriptPath)	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {	scriptPath := filepath.Join(testDir, "run.sh")	testDir := filepath.Dir(filename)	_, filename, _, _ := runtime.Caller(0)	}		t.Skip("docker not found, skipping load test")	if _, err := exec.LookPath("docker"); err != nil {	// Ensure Docker is available.	}		t.Skip("skipping load test in short mode")	if testing.Short() {func TestLoadWebSocket(t *testing.T) {//	go test -tags loadtest -v -timeout 30m ./tests/load///