// Package trigger names the command that caused a YubiKey touch by walking the process tree.
package trigger

import (
	"bufio"
	"bytes"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const maxDepth = 20

type proc struct {
	ppid int
	args []string
}

// Describe returns e.g. "git push · my-repo · via Claude Code" for the ancestors of pid, or "" if nothing is recognised.
func Describe(pid int) string {
	procs, err := processTable()
	if err != nil {
		return ""
	}
	var action, repo string
	viaClaude := false
	for depth := 0; depth < maxDepth && pid > 1; depth++ {
		p, ok := procs[pid]
		if !ok || len(p.args) == 0 {
			break
		}
		switch name := filepath.Base(p.args[0]); {
		case name == "git" && repo == "":
			action = gitAction(p.args)
			repo = cwdBase(pid)
		case name == "claude" || strings.Contains(strings.Join(p.args, " "), "claude-code"):
			viaClaude = true
		}
		pid = p.ppid
	}
	parts := make([]string, 0, 3)
	for _, s := range []string{action, repo} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	if viaClaude {
		parts = append(parts, "via Claude Code")
	}
	return strings.Join(parts, " · ")
}

func gitAction(args []string) string {
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-C" || a == "-c" || a == "--git-dir" || a == "--work-tree":
			i++
		case strings.HasPrefix(a, "-"):
		default:
			return "git " + a
		}
	}
	return "git"
}

func processTable() (map[int]proc, error) {
	out, err := exec.Command("ps", "-Ao", "pid=,ppid=,args=").Output()
	if err != nil {
		return nil, err
	}
	procs := make(map[int]proc)
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(f[0])
		ppid, err2 := strconv.Atoi(f[1])
		if err1 != nil || err2 != nil {
			continue
		}
		procs[pid] = proc{ppid: ppid, args: f[2:]}
	}
	return procs, sc.Err()
}

func cwdBase(pid int) string {
	out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if path, ok := strings.CutPrefix(line, "n"); ok {
			return filepath.Base(path)
		}
	}
	return ""
}
