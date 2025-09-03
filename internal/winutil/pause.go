package winutil

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

// ShouldPause determines whether to pause on Windows, based on flag/env or if launched via Explorer.
func ShouldPause(requested bool) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	pauseEnv := strings.ToLower(strings.TrimSpace(os.Getenv("PAUSE_ON_EXIT")))
	pauseOnExit := requested || pauseEnv == "1" || pauseEnv == "true" || pauseEnv == "yes" || pauseEnv == "on" || pauseEnv == "y"
	if pauseOnExit {
		return true
	}
	// Auto-enable if double-clicked from Explorer
	pid := os.Getppid()
	for i := 0; i < 3 && pid > 0; i++ {
		if proc, err := ps.FindProcess(pid); err == nil && proc != nil {
			if strings.ToLower(proc.Executable()) == "explorer.exe" {
				return true
			}
			pid = proc.PPid()
		} else {
			break
		}
	}
	return false
}

// Pause blocks until Enter is pressed. No-op on non-Windows.
func Pause() {
	if runtime.GOOS != "windows" {
		return
	}
	fmt.Fprint(os.Stderr, "\nPress Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
