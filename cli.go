package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tjeerdritsma/yubi-ki-yay/internal/alert"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/config"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/ipc"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/trigger"
)

const (
	bundleID     = "io.ritsma.yubikiyay"
	settingsFlag = "--settings"
	cliUsage     = "usage: yubikiyay notify [label] | yubikiyay settings"
)

// runCLI handles subcommands; ok is false when args are not a CLI invocation.
func runCLI(args []string) (code int, ok bool) {
	if len(args) == 0 {
		return 0, false
	}
	switch args[0] {
	case "notify":
		return runNotify(args[1:]), true
	case "settings":
		return runSettings(), true
	case "help", "-h", "--help":
		fmt.Println(cliUsage)
		return 0, true
	}
	return 0, false
}

// runNotify prints the terminal line itself because only the CLI owns the user's tty.
func runNotify(args []string) int {
	label := strings.Join(args, " ")
	if label == "" {
		label = "ssh"
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "yubi-ki-yay:", err)
	}
	touch := ipc.Touch{Label: label, Context: trigger.Describe(os.Getppid())}
	text := label
	if touch.Context != "" {
		text = touch.Context + " (" + label + ")"
	}
	if cfg.TerminalLine {
		alert.TerminalLine(text)
	}
	if ipc.SendTouch(touch) == nil {
		return 0
	}
	if cfg.Banner {
		if err := alert.FallbackNotification(text); err != nil {
			fmt.Fprintln(os.Stderr, "yubi-ki-yay:", err)
		}
	}
	if cfg.Sound {
		if err := alert.PlaySound(cfg.SoundName); err != nil {
			fmt.Fprintln(os.Stderr, "yubi-ki-yay:", err)
		}
	}
	return 0
}

func runSettings() int {
	if ipc.Send(ipc.CmdSettings, "") == nil {
		return 0
	}
	if err := exec.Command("open", "-b", bundleID, "--args", settingsFlag).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "yubi-ki-yay: launch app:", err)
		return 1
	}
	return 0
}
