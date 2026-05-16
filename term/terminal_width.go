package term

import (
	"os/exec"
	"strconv"
	"strings"
)

func GetTerminalWidth() int {
	output, err := exec.Command("stty", "size").Output()
	if err != nil {
		return 80
	}

	fields := strings.Fields(string(output))
	if len(fields) != 2 {
		return 80
	}

	width, err := strconv.Atoi(fields[1])
	if err != nil || width <= 0 {
		return 80
	}
	return width
}
