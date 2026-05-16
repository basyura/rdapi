package cli

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenURL(rawURL string) error {
	var command string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		command = "xdg-open"
	}
	args = append(args, rawURL)

	if err := exec.Command(command, args...).Start(); err != nil {
		return fmt.Errorf("open URL: %w", err)
	}
	return nil
}
