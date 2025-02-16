package examples

import (
	ttt "github.com/SirMoM/tmux-tui-testing"
	"testing"
)

func TestRunSpecFile(t *testing.T) {
	ttt.RunTestSpec("minimal.ttt", nil, t)
}

func TestRunDir(t *testing.T) {
	ttt.RunTestSpecDir("testfiles", nil, t)
}

func TestRunDev(t *testing.T) {
	ttt.RunTestSpec("testfiles/capture.ttt", nil, t)
}
