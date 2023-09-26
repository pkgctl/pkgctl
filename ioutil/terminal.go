package ioutil

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var spins = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	index int
}

func (s *Spinner) Current() string {
	return spins[s.index]
}

func (s *Spinner) Next() string {
	s.index = (s.index + 1) % len(spins)
	return s.Current()
}

func TerminalSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), " ")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid output from stty: %v", string(output))
	}

	lines, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	cols, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return lines, cols, nil
}
