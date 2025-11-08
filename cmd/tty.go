package cmd

import (
	"os"

	"golang.org/x/term"
)

// IsTTY checks if stdin, stdout, or stderr is a TTY
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) ||
		term.IsTerminal(int(os.Stdin.Fd())) ||
		term.IsTerminal(int(os.Stderr.Fd()))
}
