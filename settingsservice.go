package main

import (
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

func (s *SettingsService) Sounds() ([]string, error) {
	return config.SystemSounds()
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
