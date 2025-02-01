package ttt

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var TmuxWasRunning = true

// Setup ensures tmux is installed and running
// If tmux is not installed, the program will terminate with an error
func Setup(t *testing.T) {
	// Ensure tmux is installed
	_, err := exec.LookPath(Tmux)
	if err != nil {
		t.Errorf("Tmux not found %s", err)
		os.Exit(1)
	}

	// Check if tmux server is running
	var out bytes.Buffer
	cmd := exec.Command(Tmux, TmuxInfo)
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		if strings.Contains(out.String(), "no server running") {
			TmuxWasRunning = false
		} else {
			t.Log("Consider killing the tmux (`tmux kill-server`) server before running the tests")
			t.Fatal(out.String(), "err", err)
		}
	}
}

// CleanUp kills the all tmux sessions that are still open and leaves tmux in the state it was before the tests
func CleanUp(t *testing.T) {
	if !TmuxWasRunning {
		err := exec.Command(Tmux, TmuxKillServer).Run()
		if err != nil {
			t.Fatalf("Failed to kill server: %v", err)
		}
		// Don't need to kill session if server is killed ;)
		return
	}

	//TODO: Kill all sessions as soon as we can parallelize tests with different tmux sessions
	err := exec.Command(Tmux, TmuxKillSession, TmuxFlagT, SessionName).Run()
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
	cmd := exec.Command(Tmux, TmuxNewSession, TmuxDetached, TmuxSessionFlag, SessionName, programToStart)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create tmux session: %v", err)
	}
}

// CreateShSession creates a new tmux session with the sh shell
func CreateShSession(t *testing.T) {
	CreateSession("sh", t)
}

// DestroySession this kills the Tmux session with the given name
func DestroySession(sessionName string, t *testing.T) {
	err := exec.Command(Tmux, TmuxKillSession, TmuxFlagT, sessionName).Run()
	if err != nil {
		t.Fatalf("Failed to kill session: %v", err)
	}
}

// DestroyDefaultSession kills the default session with the name ttt-session
func DestroyDefaultSession(t *testing.T) {
	DestroySession(SessionName, t)
}

func RunTestSpec(filePath string, t *testing.T) {
	testSpec, err := ReadTestSpec(filePath)
	if err != nil {
		t.Fatalf("Failed to read test spec: %v", err)
	}
	t.Run(testSpec.Name, func(t *testing.T) {

		CreateSession(testSpec.RootProgramm, t)

		// Send inputs
		sendInputs(t, testSpec.Inputs)

		// Create temporary file
		tmpFile := tempFile(t)
		if !testing.Verbose() {
			defer os.Remove(tmpFile.Name())
		} else {
			defer func() {
				t.Logf("Output saved in %s", tmpFile.Name())
				exec.Command("open", tmpFile.Name()).Run()

			}()
		}
		// Capture and save output
		captureOutput(tmpFile.Name(), t)

		compareOutput(tmpFile, testSpec.ExpectedOutput, t)

		DestroyDefaultSession(t)
	})
}

func compareOutput(file *os.File, output string, t *testing.T) {
	actualOutput, err := io.ReadAll(file)
	if err != nil {
		return
	}

	if string(actualOutput) != output {
		t.Fatalf("Expected output: %s\nActual output: %s", output, actualOutput)
	}
}
