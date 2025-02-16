package internal

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// confirmKey represents the possible confirmation keys.
type confirmKey int

const (
	enter confirmKey = iota
	tab
	none
)

type metaCommand int

func (m *metaCommand) String() string {
	switch *m {
	case snapshot:
		return "Snapshot"
	default:
		return "NOT IMPLEMENTED"
	}
}

const (
	None metaCommand = iota
	snapshot
)

// String returns the string representation of the confirmKey.
func (c *confirmKey) String() string {
	switch *c {
	case enter:
		return "Enter"
	case tab:
		return "Tab"
	case none:
		return "None"
	}
	return "Unknown"
}

// tmuxInput represents a single input that can be passed to tmux.
type tmuxInput struct {
	text            string
	metaCommand     metaCommand
	confirmationKey confirmKey
	sleepMs         time.Duration
}

// tmuxInputs is a slice of tmuxInput
type tmuxInputs []tmuxInput

// String the string representation
func (t tmuxInputs) String() string {
	var s string
	for idx, i := range t {
		s += i.String()
		if idx < len(t)-1 {
			s += ", "
		} else {
			s += " "
		}
	}
	return s

}

// String the string representation
func (t tmuxInput) String() string {
	var inputText string
	if t.metaCommand != None && t.text != "" {
		inputText = t.metaCommand.String()
	} else {
		inputText = t.text
	}
	return fmt.Sprintf("'%s' %s %v", inputText, t.confirmationKey.String(), t.sleepMs)
}

const (
	SessionName     = "ttt-session"
	Tmux            = "tmux"
	TmuxKillServer  = "kill-server"
	TmuxSendKeys    = "send-keys"
	TmuxFlagT       = "-t"
	TmuxKillSession = "kill-session"
	TmuxNewSession  = "new-session"
	TmuxDetached    = "-d"
	TmuxCapturePane = "capture-pane"
	TmuxSessionFlag = "-s"
	TmuxEnterKey    = "C-m"
	TmuxTabKey      = "C-i"
)

// vLog logs a string to the console if the testing framework is in verbose mode.
func vLog(str string) {
	if testing.Verbose() {
		fmt.Println(str)
	}
}

// vLogP logs a string to the console with a leading padding if the testing framework is in verbose mode.
func vLogP(str string) {
	if testing.Verbose() {
		fmt.Println("\t|" + str)
	}
}

// SendInputs sends a series of tmux inputs, executing the appropriate actions for each input's text and confirmation key.
// It logs each input, sends text or confirmation keys (Enter, Tab, None), and waits for a specified sleep duration between inputs.
// If the terminal does not finish processing within five minutes, a timeout error is raised.
// TODO: do i want t expose this?
func SendInputs(inputs []tmuxInput, t *testing.T) []string {
	results := make([]string, 0)
	vLog("Sending Inputs:")
	for idx, input := range inputs {
		vLogP(fmt.Sprintf("%d: %v", idx, input))

		if input.metaCommand != None {
			switch input.metaCommand {
			case snapshot:
				// Take snapshot
				vLogP("Taking snapshot")
				results = append(results, strings.TrimSpace(CaptureOutput(t)))
				continue
			default:
				t.Fatalf("Invalid meta command: %v", input.metaCommand)
			}
		} else if input.text != "" {
			sendCharacters(t, input.text)
		}

		switch input.confirmationKey {
		case enter:
			sendEnter(t)
		case tab:
			sendTab(t)
		case none:
			// Do nothing
			vLogP("No confirmation key")
		default:
			t.Fatalf("Invalid input type: %v", input)
		}

		// todo is this really the best way to wait for the terminal to finish the processing
		done := make(chan bool)
		go func() {
			time.Sleep(input.sleepMs)
			done <- true
		}()

		// Optionally wait for completion
		select {
		case <-done:
			// Continue
		case <-time.After(5 * time.Minute): // Timeout after 1 minute
			t.Fatal("sendKeysWithDelay timed out")
		}
	}
	return results
}

// sendEnter sends Enter as the escaped code to the tmux session.
func sendEnter(t *testing.T) {
	cmd := exec.Command(Tmux, TmuxSendKeys, TmuxFlagT, SessionName, TmuxEnterKey)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to send Enter: %v", err)
	}
}

// sendEnter sends Tab as the escaped code to the tmux session.
func sendTab(t *testing.T) {
	cmd := exec.Command(Tmux, TmuxSendKeys, TmuxFlagT, SessionName, TmuxTabKey)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to send Tab: %v", err)
	}
}

// sendCharacters sends the provided characters to the tmux session.
func sendCharacters(t *testing.T, chars string) {
	cmd := exec.Command(Tmux, TmuxSendKeys, TmuxFlagT, SessionName, chars)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to send chars:%s :: %v", chars, err)
	}
	err := cmd.Wait()
	if err != nil {
		t.Fatalf("Failed to send chars:%s :: %v", chars, err)
	}
}

// CaptureOutput captures the content of the tmux pane and returns it as a string.
// It captures the output from the beginning till the end of the history of the main pane.
func CaptureOutput(t *testing.T) string {
	// Capture pane content
	cmd := exec.Command(Tmux, TmuxCapturePane, "-J", "-p", TmuxFlagT, SessionName, "-S", "-", "-E", "-")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Fatalf("Failed to capture pane: %v", err)
	}

	return string(out)
}

func cleanupSession(t *testing.T) {
	cmd := exec.Command(Tmux, TmuxKillSession, TmuxFlagT, SessionName)
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: failed to kill session: %v", err)
	}
}

// TestSpec represents a complete test specification from file
type TestSpec struct {
	Name           string
	RootProgramm   string
	Inputs         tmuxInputs
	ExpectedOutput string
}

// String returns a string representation of the TestSpec.
func (t *TestSpec) String() string {
	return fmt.Sprintf("Name: %s\nSubcommand: %s\nInputs: %v\nExpectedOutput\n: %s\n", t.Name, t.RootProgramm, t.Inputs, t.ExpectedOutput)
}

// parseConfirmationKey converts string to confirmKey type
func parseConfirmationKey(s string) (confirmKey, error) {
	switch strings.ToLower(s) {
	case "enter":
		return enter, nil
	case "tab":
		return tab, nil
	case "none", "":
		return none, nil
	default:
		return none, fmt.Errorf("invalid confirmation key: %s", s)
	}
}
