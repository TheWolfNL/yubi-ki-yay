// Package ipc carries requests from the CLI to the running app over a unix socket.
package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	dialTimeout = 300 * time.Millisecond
	maxLineLen  = 1024
)

// Commands understood by the app.
const (
	CmdNotify   = "notify"
	CmdSettings = "settings"
)

// SocketPath returns ~/Library/Application Support/yubi-ki-yay/notify.sock.
func SocketPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(dir, "yubi-ki-yay", "notify.sock"), nil
}

// ErrAlreadyRunning is returned by Listen when another app instance owns the socket.
var ErrAlreadyRunning = errors.New("another Yubi Ki-Yay instance is running")

// Send delivers a command and its argument to the running app.
func Send(cmd, arg string) error {
	p, err := SocketPath()
	if err != nil {
		return err
	}
	conn, err := net.DialTimeout("unix", p, dialTimeout)
	if err != nil {
		return fmt.Errorf("dial app: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(dialTimeout)); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}
	if _, err := fmt.Fprintln(conn, sanitize(cmd+" "+arg)); err != nil {
		return fmt.Errorf("send %s: %w", cmd, err)
	}
	return nil
}

// Listen serves the socket until the returned listener is closed, calling handle per request.
func Listen(handle func(cmd, arg string)) (net.Listener, error) {
	p, err := SocketPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, fmt.Errorf("create socket dir: %w", err)
	}
	if conn, err := net.DialTimeout("unix", p, dialTimeout); err == nil {
		conn.Close()
		return nil, ErrAlreadyRunning
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove stale socket: %w", err)
	}
	ln, err := net.Listen("unix", p)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	go serve(ln, handle)
	return ln, nil
}

func serve(ln net.Listener, handle func(cmd, arg string)) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(time.Second))
			line, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil && line == "" {
				return
			}
			cmd, arg, _ := strings.Cut(sanitize(line), " ")
			handle(cmd, arg)
		}()
	}
}

func sanitize(line string) string {
	line = strings.TrimSpace(strings.ReplaceAll(line, "\n", " "))
	if len(line) > maxLineLen {
		line = line[:maxLineLen]
	}
	return line
}

// Touch is the payload of a notify request.
type Touch struct {
	Label   string `json:"label"`
	Context string `json:"context,omitempty"`
}

// SendTouch asks the running app to alert for t.
func SendTouch(t Touch) error {
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("encode touch: %w", err)
	}
	return Send(CmdNotify, string(data))
}

// ParseTouch decodes a notify argument, treating non-JSON input as a bare label.
func ParseTouch(arg string) Touch {
	var t Touch
	if err := json.Unmarshal([]byte(arg), &t); err != nil {
		return Touch{Label: arg}
	}
	return t
}
