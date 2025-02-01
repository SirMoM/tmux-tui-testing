package examples

import (
	ttt "github.com/SirMoM/tmux-tui-testing"
	"os/exec"
	"testing"
)

func TestIt(t *testing.T) {
	t.Log("This is a test")
	ttt.CreateShSession(t)

	cmd := exec.Command(ttt.Tmux, "list-sessions")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Failed to list sessions: %v", err)
	}
	t.Log(string(output))

	ttt.CleanUp(t)

	output, err = exec.Command(ttt.Tmux, "list-sessions").CombinedOutput()
	if err != nil {
		t.Log(string(output))
	} else {
		t.Fatal("Tmux session still exists! It should have been killed by `tt.CleanUp()`")
	}
}

func TestRunSpecFile(t *testing.T) {
	ttt.RunTestSpec("testfiles/minimal.ttt", t)
}
