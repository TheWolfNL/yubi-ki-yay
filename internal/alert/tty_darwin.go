package alert

import (
	"os"

	"golang.org/x/sys/unix"
)

// ownsTerminal reports whether tty is a line-mode terminal whose foreground job includes this process.
func ownsTerminal(tty *os.File) bool {
	fd := int(tty.Fd())
	fg, err := unix.IoctlGetInt(fd, unix.TIOCGPGRP)
	if err != nil || fg != unix.Getpgrp() {
		return false
	}
	t, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return false
	}
	// Full-screen programs such as Claude Code switch the terminal out of canonical mode.
	return t.Lflag&unix.ICANON != 0
}
