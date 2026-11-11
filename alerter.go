package main

import (
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/tjeerdritsma/yubi-ki-yay/internal/alert"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/config"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/ipc"
)

const (
	trayFlashDuration   = 10 * time.Second
	trayStartupDuration = 5 * time.Second
)

// Alerter fans a touch request out to the app-side alert channels and owns tray visibility.
type Alerter struct {
	banner *Banner
	tray   *application.SystemTray

	mu         sync.Mutex
	cfg        config.Config
	revealStop *time.Timer
	revealing  bool
	revealGen  int
}

func newAlerter() (*Alerter, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Alerter{cfg: cfg}, nil
}

// Config returns the settings currently in effect.
func (a *Alerter) Config() config.Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

// SetConfig persists cfg and applies it to the tray and subsequent alerts.
func (a *Alerter) SetConfig(cfg config.Config) error {
	if err := config.Save(cfg); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg = cfg
	revealing := a.revealing
	a.mu.Unlock()
	if !revealing {
		a.settleTray()
	}
	return nil
}

// Trigger alerts the user that a touch is needed.
func (a *Alerter) Trigger(t ipc.Touch) {
	cfg := a.Config()
	if cfg.Banner && a.banner != nil {
		a.banner.Show(t, cfg.BannerPos)
	}
	if cfg.Sound {
		if err := alert.PlaySound(cfg.SoundName); err != nil {
			log.Printf("sound: %v", err)
		}
	}
	if cfg.TrayMode != config.TrayNever && (cfg.TrayFlash || cfg.TrayHidden()) {
		label := ""
		if cfg.TrayFlash {
			label = "Touch!"
		}
		a.revealTray(label, trayFlashDuration)
	}
}

// RevealOnStartup applies the tray mode at launch, briefly showing an on-touch icon so the user sees the app started.
func (a *Alerter) RevealOnStartup() {
	if a.Config().TrayMode == config.TrayOnTouch {
		a.revealTray("", trayStartupDuration)
		return
	}
	a.settleTray()
}

func (a *Alerter) revealTray(label string, d time.Duration) {
	if a.tray == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.revealStop != nil {
		a.revealStop.Stop()
	}
	a.revealing = true
	a.revealGen++
	gen := a.revealGen
	application.InvokeAsync(func() {
		a.tray.SetLabel(label)
		a.tray.Show()
	})
	a.revealStop = time.AfterFunc(d, func() {
		a.mu.Lock()
		stale := gen != a.revealGen
		if !stale {
			a.revealing = false
		}
		a.mu.Unlock()
		if !stale {
			a.settleTray()
		}
	})
}

func (a *Alerter) settleTray() {
	if a.tray == nil {
		return
	}
	hidden := a.Config().TrayHidden()
	application.InvokeAsync(func() {
		a.tray.SetLabel("")
		if hidden {
			a.tray.Hide()
		} else {
			a.tray.Show()
		}
	})
}
