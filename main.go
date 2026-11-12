package main

import (
	"embed"
	"errors"
	"log"
	"os"
	"slices"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/TheWolfNL/yubi-ki-yay/internal/ipc"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/tray-template.png
var trayIcon []byte

func main() {
	if code, ok := runCLI(os.Args[1:]); ok {
		os.Exit(code)
	}
	if err := runApp(); err != nil {
		log.Fatal(err)
	}
}

func runApp() error {
	alerter, err := newAlerter()
	if err != nil {
		return err
	}

	app := application.New(application.Options{
		Name:        "Yubi Ki-Yay",
		Description: "Tells you when your YubiKey wants a touch",
		Services: []application.Service{
			application.NewService(&SettingsService{alerter: alerter}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	settings := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:          "settings",
		Title:         "Yubi Ki-Yay",
		Width:         420,
		Height:        560,
		Hidden:        true,
		DisableResize: true,
		URL:           "/",
	})
	settings.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		settings.Hide()
		e.Cancel()
	})

	tray := app.SystemTray.New()
	tray.SetTemplateIcon(trayIcon)
	tray.SetTooltip("Yubi Ki-Yay")
	alerter.tray = tray
	alerter.banner = &Banner{app: app}

	menu := app.NewMenu()
	menu.AddSectionHeader("Yubi Ki-Yay")
	menu.Add("Test alert").OnClick(func(*application.Context) { alerter.Trigger(ipc.Touch{Label: "test alert"}) })
	showSettings := func() {
		settings.Show()
		settings.Focus()
	}
	menu.Add("Settings…").OnClick(func(*application.Context) { showSettings() })
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) { app.Quit() })
	tray.SetMenu(menu)

	ln, err := ipc.Listen(func(cmd, arg string) {
		switch cmd {
		case ipc.CmdNotify:
			alerter.Trigger(ipc.ParseTouch(arg))
		case ipc.CmdSettings:
			showSettings()
		}
	})
	if errors.Is(err, ipc.ErrAlreadyRunning) {
		log.Print(err)
		return nil
	}
	if err != nil {
		return err
	}
	defer ln.Close()

	app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(*application.ApplicationEvent) { showSettings() })

	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		alerter.RevealOnStartup()
		if slices.Contains(os.Args[1:], settingsFlag) {
			showSettings()
		}
	})

	return app.Run()
}
