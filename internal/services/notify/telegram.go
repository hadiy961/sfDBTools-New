// File: internal/services/notify/telegram.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sfdbtools/internal/shared/database"
	"time"
)

// telegramAPIURL adalah base URL Telegram Bot API
const telegramAPIURL = "https://api.telegram.org/bot%s/sendMessage"
const telegramUpdatesURL = "https://api.telegram.org/bot%s/getUpdates"
const telegramDeleteWebhookURL = "https://api.telegram.org/bot%s/deleteWebhook"

// TelegramUpdate adalah struktur untuk getUpdates
type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		From struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
	} `json:"message"`
}

type telegramGetUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

// GetUpdates mengambil pesan terbaru yang masuk ke bot
func (s *Service) GetUpdates() ([]TelegramUpdate, error) {
	botToken := s.cfg.Telegram.BotToken
	if botToken == "" {
		botToken = database.GetSetting("telegram_bot_token")
	}

	if botToken == "" {
		return nil, fmt.Errorf("telegram bot_token harus diisi di config atau settings")
	}

	url := fmt.Sprintf(telegramUpdatesURL, botToken)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 409 {
		return nil, fmt.Errorf("conflict (409): getUpdates tidak bisa jalan jika Webhook aktif. Silakan jalankan 'sfdbtools notify delete-webhook' terlebih dahulu")
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("telegram API returned HTTP %d", resp.StatusCode)
	}

	var result telegramGetUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Result, nil
}

// DeleteWebhook menghapus webhook aktif agar bisa menggunakan polling (getUpdates)
func (s *Service) DeleteWebhook() error {
	botToken := s.cfg.Telegram.BotToken
	if botToken == "" {
		botToken = database.GetSetting("telegram_bot_token")
	}

	if botToken == "" {
		return fmt.Errorf("telegram bot_token harus diisi di config atau settings")
	}

	url := fmt.Sprintf(telegramDeleteWebhookURL, botToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to delete webhook: HTTP %d", resp.StatusCode)
	}
	return nil
}

// telegramPayload adalah struktur JSON untuk Telegram sendMessage
type telegramPayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// sendTelegram mengirimkan Message ke Telegram
func (s *Service) sendTelegram(msg Message) error {
	enabled := s.cfg.Telegram.Enabled
	if !enabled {
		enabled = database.GetSettingBool("telegram_enabled")
	}

	if !enabled {
		return nil // silently skip jika tidak diaktifkan
	}

	botToken := s.cfg.Telegram.BotToken
	if botToken == "" {
		botToken = database.GetSetting("telegram_bot_token")
	}

	chatID := s.cfg.Telegram.ChatID
	if chatID == "" {
		chatID = database.GetSetting("telegram_chat_id")
	}

	if botToken == "" || chatID == "" {
		return fmt.Errorf("telegram bot_token dan chat_id harus diisi di config atau settings")
	}

	cfg := s.cfg.Telegram
	parseMode := cfg.ParseMode
	if parseMode == "" {
		parseMode = "HTML"
	}

	text := FormatTelegramMessage(msg)

	payload := telegramPayload{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	url := fmt.Sprintf(telegramAPIURL, botToken)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram API returned HTTP %d", resp.StatusCode)
	}
	return nil
}
