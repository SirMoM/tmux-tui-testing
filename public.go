package ttt

import (
	"bytes"
	"github.com/SirMoM/tmux-tui-testing/internal"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var tmuxWasRunning = true

// Setup ensures tmux is installed and running
// If tmux is not installed, the program will terminate with an error
func Setup(t *testing.T) {
	// Ensure tmux is installed
	_, err := exec.LookPath(internal.Tmux)
	if err != nil {
		t.Errorf("Tmux not found %s", err)
		os.Exit(1)
	}

	// Check if tmux server is running
	var out bytes.Buffer
	cmd := exec.Command(internal.Tmux, "list-sessions")
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		if strings.Contains(out.String(), "no server running") {
			tmuxWasRunning = false
		} else {
			t.Log("Consider killing the tmux (`tmux kill-server`) server before running the tests")
			t.Fatal(out.String(), "err", err)
		}
	}

	DestroyDefaultSession(t)

}

// CleanUp kills the all tmux sessions that are still open and leaves tmux in the state it was before the tests
func CleanUp(t *testing.T) {
	if !tmuxWasRunning {
		err := exec.Command(internal.Tmux, internal.TmuxKillServer).Run()
		if err != nil {
			t.Fatalf("Failed to kill server: %v", err)
		}
		// Don't need to kill session if server is killed ;)
		return
	}

	//TODO: Kill all sessions as soon as we can parallelize tests with different tmux sessions
	err := exec.Command(internal.Tmux, internal.TmuxKillSession, internal.TmuxFlagT, internal.SessionName).Run()
	if err != nil {
		t.Errorf("Failed to kill session: %v", err)
	}
}

// CreateSession creates a new tmux session with the given programToStart
// The session is created in detached mode
// The session is named ttt-session
//
// # It is recommended to use a shell as the programToStart
//
// The function will terminate the program if it fails to create the session
func CreateSession(programToStart string, t *testing.T) {
	Setup(t)
	cmd := exec.Command(internal.Tmux, internal.TmuxNewSession, internal.TmuxDetached, internal.TmuxSessionFlag, internal.SessionName, programToStart)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Log(string(out))
		t.Fatalf("Failed to create tmux session: %v", err)
	}
	// TODO: Add a timeout to wait for the session to be created or check with Tmux to see if the session is ready
	time.Sleep(500 * time.Millisecond)
}

// CreateShSession creates a new tmux session with the sh shell
func CreateShSession(t *testing.T) {
	CreateSession("sh", t)
}

// DestroySession this kills the Tmux session with the given name
func DestroySession(sessionName string, t *testing.T) {
	err := exec.Command(internal.Tmux, internal.TmuxKillSession, internal.TmuxFlagT, sessionName).Run()
	if err != nil {
		t.Logf("Failed to kill session: %v", err)
	}
}

// DestroyDefaultSession kills the default session with the name ttt-session
func DestroyDefaultSession(t *testing.T) {
	DestroySession(internal.SessionName, t)
}

// RunTestSpec reads a test specification from the provided file path, then executes the test by creating a tmux session,
// sending inputs, capturing the output, and comparing it with the expected output. If any error occurs during the process,
// it reports the failure and halts the test.
//
// The cleanupFunc is called after each test in the directory. If no cleanup function is wanted, pass nil.
func RunTestSpec(filePath string, cleanupFunc *func(), t *testing.T) {
	testSpec, err := internal.ReadTestSpec(filePath)
	if err != nil {
		t.Errorf("Failed to read test spec: %v", err)
		t.FailNow()
	}
	t.Run(testSpec.Name, func(ot *testing.T) {
		if cleanupFunc != nil {
			ot.Cleanup(*cleanupFunc)
		}

		CreateSession(testSpec.RootProgramm, ot)

		// Send inputs
		snapshots := internal.SendInputs(testSpec.Inputs, ot)

		// Capture and save output
		actualOutput := strings.TrimSpace(internal.CaptureOutput(ot))
		actualOutput = strings.Join(snapshots, "\n") + "\n" + actualOutput
		internal.CompareOutput(actualOutput, testSpec.ExpectedOutput, ot)

		DestroyDefaultSession(ot)
	})
}

// RunTestSpecDir reads and executes test specifications from all files in the specified directory.
// It iterates over the directory entries and calls RunTestSpec for each file. If an error occurs while reading the directory, it reports the failure.
//
// The cleanupFunc is called after each test in the directory. If no cleanup function is wanted, pass nil.
//
// **This doesn't read the subdirectories recursively!**
func RunTestSpecDir(dirPath string, cleanupFunc *func(), t *testing.T) {
	if dirEntries, err := os.ReadDir(dirPath); err == nil {
		for _, entry := range dirEntries {
			if entry.IsDir() {
				continue
			}
			filePath := filepath.Join(dirPath, entry.Name())
			RunTestSpec(filePath, cleanupFunc, t)
		}
	} else {
		t.Fatal(err)
	}
}
