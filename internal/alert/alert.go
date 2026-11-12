// Package alert holds the alert channels that need no running app: terminal line, sound and the osascript fallback.
package alert

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/TheWolfNL/yubi-ki-yay/internal/config"
)

// Title is the headline used by every alert channel.
const Title = "👆 Touch your YubiKey"

// TerminalLine writes the prompt to the controlling terminal unless another program owns it.
func TerminalLine(text string) {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer tty.Close()
	if !ownsTerminal(tty) {
		return
	}
	fmt.Fprintf(tty, "\a\033[1;33m%s — %s\033[0m\n", Title, text)
}

// PlaySound starts afplay without waiting for it to finish.
func PlaySound(name string) error {
	cmd := exec.Command("afplay", config.SoundPath(name))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start afplay: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

// FallbackNotification posts a notification via osascript, used when the app is not running.
func FallbackNotification(text string) error {
	script := fmt.Sprintf("display notification %s with title %s",
		strconv.Quote(text), strconv.Quote(Title))
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start osascript: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
