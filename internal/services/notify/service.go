// File: internal/services/notify/service.go
package notify

import (
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// Service adalah shared notification service.
// Thread-safe: boleh dipakai dari goroutine manapun.
type Service struct {
	cfg    *appconfig.NotifyConfig
	logger applog.Logger
}

// NewService membuat instance Service baru.
// cfg boleh nil — dalam kondisi ini semua Send() akan no-op dengan log warning.
func NewService(cfg *appconfig.NotifyConfig, logger applog.Logger) *Service {
	return &Service{cfg: cfg, logger: logger}
}

// Send mengirimkan Message ke semua channel yang ditentukan.
// Jika msg.Channels kosong, fallback ke cfg.DefaultChannels.
// Error pada satu channel tidak menghentikan channel lain (best-effort).
func (s *Service) Send(msg Message) *SendReport {
	report := &SendReport{Message: msg}

	if s.cfg == nil {
		s.logger.Warn("[notify] config is nil, notification skipped")
		return report
	}

	channels := msg.Channels
	if len(channels) == 0 {
		// Konversi []string ke []Channel
		for _, ch := range s.cfg.DefaultChannels {
			channels = append(channels, Channel(ch))
		}
	}
	if len(channels) == 0 {
		s.logger.Debug("[notify] no channels configured, skipping")
		return report
	}

	for _, ch := range channels {
		var err error
		switch ch {
		case ChannelTelegram:
			err = s.sendTelegram(msg)
		case ChannelEmail:
			err = s.sendEmail(msg)
		default:
			s.logger.Warnf("[notify] unknown channel: %s", ch)
			continue
		}

		result := Result{Channel: ch, Success: err == nil, Err: err}
		if err != nil {
			s.logger.Errorf("[notify] failed to send via %s: %v", ch, err)
		} else {
			s.logger.Debugf("[notify] sent via %s", ch)
		}
		report.Results = append(report.Results, result)
	}

	return report
}

// SendTest mengirimkan pesan test ke semua channel yang enabled.
// Berguna untuk validasi konfigurasi via CLI.
func (s *Service) SendTest() *SendReport {
	return s.Send(Message{
		Title:   "[sfDBTools] Test Notification",
		Body:    "Ini adalah pesan test dari sfDBTools. Jika Anda menerima pesan ini, konfigurasi notifikasi sudah benar.",
		Level:   LevelInfo,
		Feature: "system",
	})
}
