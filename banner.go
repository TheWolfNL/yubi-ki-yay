package main

import (
	"net/url"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/tjeerdritsma/yubi-ki-yay/internal/ipc"
)

const (
	bannerWidth    = 400
	bannerHeight   = 96
	bannerMargin   = 8
	bannerDuration = 10 * time.Second
	bannerHideAnim = 300 * time.Millisecond
)

// Banner slides a touch prompt into the top-right corner of the active screen.
type Banner struct {
	app *application.App

	mu    sync.Mutex
	win   *application.WebviewWindow
	timer *time.Timer
}

// Show replaces any visible banner with one for t.
func (b *Banner) Show(t ipc.Touch) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closeLocked()

	q := url.Values{"label": {t.Label}, "context": {t.Context}}
	screen := activeScreen(b.app)
	if screen == nil {
		screen = b.app.Screen.GetPrimary()
	}
	win := b.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "banner",
		Width:            bannerWidth,
		Height:           bannerHeight,
		Hidden:           true,
		Frameless:        true,
		DisableResize:    true,
		Screen:           screen,
		URL:              "/#banner?" + q.Encode(),
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),
		KeyBindings: map[string]func(application.Window){
			"Escape": func(application.Window) { go b.Dismiss() },
		},
		Mac: application.MacWindow{
			Backdrop:      application.MacBackdropTransparent,
			DisableShadow: true,
			WindowClass:   application.MacWindowClassPanel,
			PanelPreferences: application.MacPanelPreferences{
				NonActivating: true,
				FloatingPanel: true,
			},
			WindowLevel: application.MacWindowLevelPopUpMenu,
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces |
				application.MacWindowCollectionBehaviorFullScreenAuxiliary |
				application.MacWindowCollectionBehaviorStationary |
				application.MacWindowCollectionBehaviorIgnoresCycle,
		},
	})
	if screen != nil {
		menuBar := screen.WorkArea.Y - screen.Bounds.Y
		win.SetRelativePosition(screen.Bounds.Width-bannerWidth-bannerMargin, menuBar+bannerMargin)
	}
	win.Show()
	b.win = win
	b.timer = time.AfterFunc(bannerDuration, b.Dismiss)
}

// Dismiss lets the banner play its exit animation, then closes it.
func (b *Banner) Dismiss() {
	b.mu.Lock()
	win := b.win
	b.win = nil
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.mu.Unlock()
	if win == nil {
		return
	}
	win.ExecJS(`document.body.classList.add("leaving")`)
	time.AfterFunc(bannerHideAnim, win.Close)
}

func (b *Banner) closeLocked() {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	if b.win != nil {
		b.win.Close()
		b.win = nil
	}
}
