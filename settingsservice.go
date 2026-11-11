package main

import (
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/tjeerdritsma/yubi-ki-yay/internal/alert"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/config"
	"github.com/tjeerdritsma/yubi-ki-yay/internal/ipc"
)

// SettingsService is the frontend's API for reading and changing settings.
type SettingsService struct {
	alerter *Alerter
}

func (s *SettingsService) Get() config.Config {
	return s.alerter.Config()
}

func (s *SettingsService) Save(cfg config.Config) error {
	return s.alerter.SetConfig(cfg)
}

// SoundList groups the selectable alert sounds.
type SoundList struct {
	Custom []string `json:"custom"`
	System []string `json:"system"`
}

func (s *SettingsService) Sounds() (SoundList, error) {
	system, err := config.SystemSounds()
	if err != nil {
		return SoundList{}, err
	}
	custom, err := config.CustomSounds()
	if err != nil {
		return SoundList{}, err
	}
	return SoundList{Custom: custom, System: system}, nil
}

// AddSound lets the user pick an audio file and returns its sound name, or "" when cancelled.
func (s *SettingsService) AddSound() (string, error) {
	dlg := application.Get().Dialog.OpenFile().
		SetTitle("Choose an alert sound").
		CanChooseFiles(true).
		AddFilter("Audio", "*"+strings.Join(config.AudioExtensions, ";*"))
	path, err := dlg.PromptForSingleSelection()
	if err != nil || path == "" {
		return "", err
	}
	return config.ImportSound(path)
}

func (s *SettingsService) RemoveSound(name string) error {
	return config.RemoveSound(name)
}

func (s *SettingsService) PreviewSound(name string) error {
	return alert.PlaySound(name)
}

func (s *SettingsService) Test() {
	s.alerter.Trigger(ipc.Touch{Label: "test alert"})
}

func (s *SettingsService) DismissBanner() {
	if s.alerter.banner != nil {
		s.alerter.banner.Dismiss()
	}
}

func (s *SettingsService) LaunchAtLogin() (bool, error) {
	return application.Get().Autostart.IsEnabled()
}

func (s *SettingsService) SetLaunchAtLogin(enabled bool) error {
	if enabled {
		return application.Get().Autostart.Enable()
	}
	return application.Get().Autostart.Disable()
}
