package ttt

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Create a temporary file
func tempFile(t *testing.T) *os.File {
	tmp, err := os.CreateTemp("", "tempfile-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Logf("Temp file created: %s", tmp.Name())
	return tmp
}

type confirmKey int

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

const (
	enter confirmKey = iota
	tab
	none
)

type tmuxInput struct {
	text            string
	confirmationKey confirmKey
	sleepMs         time.Duration
}
type tmuxInputs []tmuxInput

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

func (t tmuxInput) String() string {
	return fmt.Sprintf("'%s' %s %v", t.text, t.confirmationKey.String(), t.sleepMs)
}

const (
	SessionName     = "ttt-session"
	Tmux            = "tmux"
	TmuxInfo        = "info"
	TmuxKillServer  = "kill-server"
	TmuxSendKeys    = "send-keys"
	TmuxFlagT       = "-t"
	TmuxKillSession = "kill-session"
	TmuxStartServer = "start-server"
	TmuxNewSession  = "new-session"
	TmuxDetached    = "-d"
	TmuxCapturePane = "capture-pane"
	TmuxSaveBuffer  = "save-buffer"
	TmuxSessionFlag = "-s"
	TmuxEnterKey    = "C-m"
	TmuxTabKey      = "C-i"
)

func createSession(t *testing.T, vpdArgs string) {
	cmd := exec.Command(Tmux, TmuxNewSession, TmuxDetached, TmuxSessionFlag, SessionName, "sh")

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create tmux session: %v", err)
	}

	//time.Sleep(50 * time.Millisecond)
	//sendText(t, vpdArgs)
	//sendEnter(t)
	time.Sleep(500 * time.Millisecond)
}

// TODO: do i want t expose this?
func sendInputs(t *testing.T, inputs []tmuxInput) {

	for idx, input := range inputs {
		fmt.Println("Sending input:")

		if input.text != "" {
			sendText(t, input.text)
		}

		switch input.confirmationKey {
		case enter:
			sendEnter(t)
		case tab:
			sendTab(t)
		case none:
			// Do nothing
			t.Log("No confirmation key")
		default:
			t.Fatalf("Invalid input type: %v", input)
		}

		//time.Sleep(input.sleepMs)
		fmt.Printf("%d: %v\n", idx, input)
		done := make(chan bool)

		go func() {
			time.Sleep(input.sleepMs)
			done <- true
		}()

		// Optionally wait for completion
		select {
		case <-done:
			// Continue
			fmt.Println("Continuing")
		case <-time.After(time.Minute): // Timeout after 1 minute
			t.Fatal("sendKeysWithDelay timed out")
		}
	}
}

func sendEnter(t *testing.T) {
	cmd := exec.Command(Tmux, TmuxSendKeys, TmuxFlagT, SessionName, TmuxEnterKey)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to send Enter: %v", err)
	}
}

func sendTab(t *testing.T) {
	cmd := exec.Command(Tmux, TmuxSendKeys, TmuxFlagT, SessionName, TmuxTabKey)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to send Tab: %v", err)
	}
}
func sendText(t *testing.T, text string) {
	cmd := exec.Command(Tmux, TmuxSendKeys, TmuxFlagT, SessionName, text)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to send text:%s :: %v", text, err)
	}
	err := cmd.Wait()
	if err != nil {
		t.Fatalf("Failed to send text:%s :: %v", text, err)
	}
}

func captureOutput(fileName string, t *testing.T) {
	// Capture pane content
	cmd := exec.Command(Tmux, TmuxCapturePane, TmuxFlagT, SessionName, "-S", "-", "-E", "-")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to capture pane: %v", err)
	}

	// Save buffer to file
	cmd = exec.Command(Tmux, TmuxSaveBuffer, fileName)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to save buffer: %v", err)
	}
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

type specSection int

const (
	sectionStart specSection = iota
	sectionInputs
	sectionArgs
	sectionExpectedOutput
)

// ReadTestSpec parses a test specification file
func ReadTestSpec(filePath string) (*TestSpec, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening test file: %w", err)
	}
	defer file.Close()

	spec := &TestSpec{}
	currentSection := sectionStart
	lineNr := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			currentSection = sectionExpectedOutput
			lineNr++
			continue
		}

		switch {
		case strings.HasPrefix(line, "#"):
			err = parseNameSection(&currentSection, spec, line, lineNr)
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(line, "%"):
			err = parseArgsSection(&currentSection, spec, line, lineNr)
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(line, ">"):
			err = parseInputsSection(&currentSection, spec, line, lineNr)
			if err != nil {
				return nil, err
			}

		default:
			if currentSection == sectionExpectedOutput {
				spec.ExpectedOutput += line + "\n"
			} else {
				return nil, fmt.Errorf("unparsable line in line" + strconv.Itoa(lineNr))
			}
		}
		lineNr++
	}

	// Clean up final output
	spec.ExpectedOutput = strings.TrimSpace(spec.ExpectedOutput)

	return spec, nil
}

// TODO parse " quotes in input section
func parseInputsSection(currentSection *specSection, spec *TestSpec, line string, lineNr int) error {
	if *currentSection != sectionInputs {
		return fmt.Errorf("input line before command specification " + strconv.Itoa(lineNr))
	}

	// Split line by spaces and quotes
	quoted := false
	parts := strings.FieldsFunc(line[1:], func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return r == ' ' && !quoted
	})

	if len(parts) < 3 {
		return fmt.Errorf(`invalid input format, expected: > "text" key sleepInMs`)
	}

	unquotedText := strings.Trim(parts[0], "\"")

	key, err := parseConfirmationKey(parts[1])
	if err != nil {
		return err
	}

	var sleep time.Duration
	if parts[2] != "0" {
		i, err := strconv.Atoi(parts[2])
		if err != nil {
			return err
		}
		sleep = time.Duration(i) * time.Millisecond
	}

	spec.Inputs = append(spec.Inputs, tmuxInput{
		text:            unquotedText,
		confirmationKey: key,
		sleepMs:         sleep,
	})

	return nil
}

func parseNameSection(currentSection *specSection, spec *TestSpec, line string, lineNr int) error {
	if *currentSection != sectionStart {
		return fmt.Errorf("unexpected name declaration in line " + strconv.Itoa(lineNr))
	}
	spec.Name = strings.TrimSpace(line[1:])
	*currentSection = sectionArgs
	return nil
}
func parseArgsSection(currentSection *specSection, spec *TestSpec, line string, lineNr int) error {
	if *currentSection != sectionArgs {
		return fmt.Errorf("unexpected command declaration in line " + strconv.Itoa(lineNr))
	}
	spec.RootProgramm = strings.TrimSpace(line[1:])
	*currentSection = sectionInputs
	return nil
}
