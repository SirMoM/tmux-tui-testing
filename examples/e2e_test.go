package examples

//
//import (
//	"github.com/SirMoM/tmux-tui-testing"
//	"os"
//	"os/exec"
//	"testing"
//)
//
//// TestMain handles setup and teardown
//func TestMain(m *testing.M) {
//	// Ensure tmux server is running
//	exec.Command(ttt.Tmux, ttt.TmuxStartServer).Run()
//
//	// Run tests
//	code := m.Run()
//
//	// Cleanup
//	exec.Command(ttt.Tmux, ttt.TmuxKillSession, ttt.TmuxFlagT, ttt.SessionName).Run()
//	os.Exit(code)
//}
//
//func TestAllFiles(t *testing.T) {
//	files, err := os.ReadDir("testfiles")
//	if err != nil {
//		t.Fatalf("Failed to read directory: %v", err)
//	}
//	for _, file := range files {
//		if file.IsDir() {
//			continue
//		}
//		t.Logf("Running test for file: %s", file.Name())
//		testSpec, err := ttt.ReadTestSpec("testfiles/" + file.Name())
//		if err != nil {
//			t.Fatalf("Failed to read test spec: %v", err)
//		}
//		t.Run(testSpec.Name, func(t *testing.T) {
//			ttt.cleanupSession(t)
//			ttt.createSession(t, testSpec.RootProgramm)
//			// defer cleanupSession(t)
//
//			// Send inputs
//			ttt.sendInputs(t, testSpec.Inputs)
//
//			// Create temporary file
//			tmpFile := ttt.tempFile(t)
//			// Capture and save output
//			ttt.captureOutput(t, tmpFile.Name())
//		})
//	}
//}
//
//func TestReadTestSpec(t *testing.T) {
//	testSpec, err := ttt.ReadTestSpec("testfiles/Case.txt")
//	if err != nil {
//		t.Fatalf("Failed to read test spec: %v", err)
//	}
//	ttt.cleanupSession(t)
//	ttt.createSession(t, testSpec.RootProgramm)
//	// defer cleanupSession(t)
//
//	// Send inputs
//	ttt.sendInputs(t, testSpec.Inputs)
//
//	// Create temporary file
//	tmpFile := ttt.tempFile(t)
//	// Capture and save output
//	ttt.captureOutput(t, tmpFile.Name())
//}
