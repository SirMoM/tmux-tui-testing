package examples

import (
	ttt "github.com/SirMoM/tmux-tui-testing"
	"testing"
)

func TestRunSpecFile(t *testing.T) {
	ttt.RunTestSpec("minimal.ttt", t)
}

func TestRunDir(t *testing.T) {
	ttt.RunTestSpecDir("testfiles", t)
}
