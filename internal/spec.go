package internal

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// specSection represents the different sections in a test specification.
type specSection int

const (
	sectionStart specSection = iota
	sectionInputs
	sectionArgs
	sectionExpectedOutput
)

// String returns the string representation of the specSection.
func (s *specSection) String() string {
	switch *s {
	case 0:
		return "Start"
	case 1:
		return "Input"
	case 2:
		return "Args"
	case 3:
		return "ExpectedOutput"
	default:
		fmt.Fprintf(os.Stderr, "Wrong section type: %d", *s)
		return "Unknown"
	}
}

// ReadTestSpec parses a test specification file
func ReadTestSpec(filePath string) (*TestSpec, error) {
	vLogP("Reading TestSpec:")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening test file: %w", err)
	}
	defer file.Close()

	spec := &TestSpec{}
	currentSection := sectionStart

	fullFile, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading test file: %w", err)
	}

	for lineNr, line := range strings.Split(string(fullFile), "\n") {
		// remove windows artifact
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" {
			if currentSection != sectionExpectedOutput {
				currentSection = sectionExpectedOutput
				continue
			}
		}

		switch {
		case strings.HasPrefix(line, "#"):
			err = parseNameSection(&currentSection, spec, line, lineNr)
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(line, "%"):
			err = parseRootProgramSection(&currentSection, spec, line, lineNr)
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
				return nil, fmt.Errorf("unparsable line in line " + strconv.Itoa(lineNr))
			}
		}
	}

	// Clean up final output
	spec.ExpectedOutput = strings.TrimSpace(spec.ExpectedOutput)
	vLogP("Done reading TestSpec")
	vLogP(fmt.Sprintf("%+#v", spec))
	return spec, nil
}

// parseInputsSection parses the input section of the specification and updates the TestSpec.
// It expects the currentSection to be sectionInputs, and processes the line to extract inputs, keys, and sleep durations.
// TODO parse " quotes in input section
func parseInputsSection(currentSection *specSection, spec *TestSpec, line string, lineNr int) error {
	if *currentSection != sectionInputs {
		return fmt.Errorf("input line before command specification " + strconv.Itoa(lineNr))
	}

	// Split line by spaces and quotes
	quoted := false
	parts := strings.FieldsFunc(strings.TrimSpace(line[1:]), func(r rune) bool {
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

// parseNameSection parses the name section of the specification and updates the TestSpec.
// It expects the currentSection to be sectionStart and sets the Name field in the spec.
func parseNameSection(currentSection *specSection, spec *TestSpec, line string, lineNr int) error {
	if *currentSection != sectionStart {
		return fmt.Errorf("unexpected name declaration in line " + strconv.Itoa(lineNr))
	}
	spec.Name = strings.TrimSpace(line[1:])
	*currentSection = sectionArgs
	return nil
}

// parseRootProgramSection parses the root program section of the specification and updates the TestSpec.
// It expects the currentSection to be sectionArgs and sets the RootProgramm field in the spec.
func parseRootProgramSection(currentSection *specSection, spec *TestSpec, line string, lineNr int) error {
	if *currentSection != sectionArgs {
		return fmt.Errorf("unexpected command declaration in line " + strconv.Itoa(lineNr))
	}
	spec.RootProgramm = strings.TrimSpace(line[1:])
	*currentSection = sectionInputs
	return nil
}

// CompareOutput compares the actual output with the expected output and reports differences.
// todo this is lacking implement edge cases
func CompareOutput(actualOutput, expectedOutput string, t *testing.T) {
	vLog("Compare Output")
	expectedOutput = strings.TrimSpace(expectedOutput)
	actualOutput = strings.TrimSpace(actualOutput)

	vLog("Actual Output: \n" + actualOutput + "\n")
	vLog("Expect Output: \n" + expectedOutput)

	expectedOutputLines := strings.Split(expectedOutput, "\n")
	actualOutputLines := strings.Split(actualOutput, "\n")
	if len(expectedOutputLines) != len(actualOutputLines) {
		t.Errorf("Expected %d lines got %d actual lines!", len(expectedOutputLines), len(actualOutputLines))
	}

	maxLen := len(actualOutputLines)
	if len(expectedOutputLines) > maxLen {
		maxLen = len(expectedOutputLines)
	}

	for idx := 0; idx < maxLen; idx++ {
		actualLine, expectedLine := "", ""

		if idx < len(actualOutputLines) {
			actualLine = strings.TrimSpace(actualOutputLines[idx])
		}
		if idx < len(expectedOutputLines) {
			expectedLine = strings.TrimSpace(expectedOutputLines[idx])
		}
		if strings.Compare(actualLine, expectedLine) != 0 {
			t.Errorf("Line %d differ:\n", idx+1)
			t.Errorf("\nA: '%s'\nE: '%s'\n", actualLine, expectedLine)
		}
	}

}
