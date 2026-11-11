package main

import (
	"net/url"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/tjeerdritsma/yubi-ki-yay/internal/config"
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

// Show replaces any visible banner with one for t at pos.
func (b *Banner) Show(t ipc.Touch, pos string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closeLocked()

	q := url.Values{"label": {t.Label}, "context": {t.Context}, "pos": {pos}}
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
		win.SetRelativePosition(bannerOrigin(screen, pos))
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

// bannerOrigin returns the banner's top-left corner relative to screen, kept inside the area not covered by the menu bar and Dock.
func bannerOrigin(screen *application.Screen, pos string) (x, y int) {
	wa := screen.WorkArea
	left := wa.X - screen.Bounds.X + bannerMargin
	right := wa.X - screen.Bounds.X + wa.Width - bannerWidth - bannerMargin
	top := wa.Y - screen.Bounds.Y + bannerMargin
	bottom := wa.Y - screen.Bounds.Y + wa.Height - bannerHeight - bannerMargin

	switch pos {
	case config.BannerTopLeft:
		return left, top
	case config.BannerTopCenter:
		return (screen.Bounds.Width - bannerWidth) / 2, top
	case config.BannerBottomLeft:
		return left, bottom
	case config.BannerBottomRight:
		return right, bottom
	default:
		return right, top
	}
}
