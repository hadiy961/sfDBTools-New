package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime"
	"sfdbtools/internal/app/version"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/systeminfo"
	"time"
)

type SyncManager struct {
	localDB    *sql.DB
	remote     RemoteProvider
	clientCode string
	masterKey  string
}

func NewSyncManager(localDB *sql.DB, remote RemoteProvider, clientCode, masterKey string) *SyncManager {
	return &SyncManager{
		localDB:    localDB,
		remote:     remote,
		clientCode: clientCode,
		masterKey:  masterKey,
	}
}

func (m *SyncManager) RunSync(ctx context.Context, mode string) error {
	// 1. Migrate Remote
	if err := m.remote.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate remote: %w", err)
	}

	// 2. Sync Logic based on mode
	// Mode: push-only, pull-only, two-way
	if mode == "two-way" || mode == "pull-only" {
		if err := m.PullAll(ctx); err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}
	}

	if mode == "two-way" || mode == "push-only" {
		if err := m.PushAll(ctx); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}
	}

	// 3. Heartbeat
	if err := m.SendHeartbeat(ctx); err != nil {
		m.logSyncHistory("heartbeat", "failed", 0, err.Error())
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	m.logSyncHistory(mode, "success", 0, "")
	return nil
}

func (m *SyncManager) PullAll(ctx context.Context) error {
	// 1. Pull Settings
	settings, err := m.remote.PullSettings(ctx, m.clientCode)
	if err != nil {
		return err
	}
	for _, s := range settings {
		var localUpdatedAt time.Time
		err := m.localDB.QueryRow("SELECT updated_at FROM app_settings WHERE key = ?", s.Key).Scan(&localUpdatedAt)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if s.IsLocked || s.UpdatedAt.After(localUpdatedAt) {
			isLockedVal := 0
			if s.IsLocked {
				isLockedVal = 1
			}
			_, err = m.localDB.Exec("INSERT OR REPLACE INTO app_settings (key, value, category, is_locked, updated_at) VALUES (?, ?, ?, ?, ?)",
				s.Key, s.Value, s.Category, isLockedVal, s.UpdatedAt)
			if err != nil {
				return err
			}
		}
	}

	// 2. Pull Profiles
	profiles, err := m.remote.PullProfiles(ctx, m.clientCode)
	if err != nil {
		return err
	}
	for _, p := range profiles {
		_, err = m.localDB.Exec("INSERT OR REPLACE INTO profiles (name, encrypted_data, account_code, created_at) VALUES (?, ?, ?, ?)",
			p.Name, p.EncryptedData, p.AccountCode, p.UpdatedAt)
		if err != nil {
			return err
		}
	}

	// 3. Pull Jobs
	jobs, err := m.remote.PullJobs(ctx, m.clientCode)
	if err != nil {
		return err
	}
	for _, j := range jobs {
		enabled := 0
		if j.Enabled {
			enabled = 1
		}
		_, err = m.localDB.Exec(`INSERT OR REPLACE INTO backup_jobs (name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, last_run) 
								VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			j.Name, enabled, j.Schedule, j.Mode, j.OutputMode, j.IncludeFile, j.ProfileName, j.Ticket, j.OutputDir, j.RetentionDays, j.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *SyncManager) PushAll(ctx context.Context) error {
	// 1. Push Settings
	rows, err := m.localDB.Query("SELECT key, value, category, is_locked, updated_at FROM app_settings")
	if err != nil {
		return err
	}
	defer rows.Close()

	var settings []SyncSetting
	for rows.Next() {
		var s SyncSetting
		var isLocked int
		if err := rows.Scan(&s.Key, &s.Value, &s.Category, &isLocked, &s.UpdatedAt); err != nil {
			return err
		}
		s.IsLocked = isLocked == 1
		settings = append(settings, s)
	}
	if err := m.remote.PushSettings(ctx, m.clientCode, settings); err != nil {
		return err
	}

	// 2. Push Profiles
	pRows, err := m.localDB.Query("SELECT name, encrypted_data, account_code, created_at FROM profiles")
	if err != nil {
		return err
	}
	defer pRows.Close()

	var profiles []SyncProfile
	for pRows.Next() {
		var p SyncProfile
		if err := pRows.Scan(&p.Name, &p.EncryptedData, &p.AccountCode, &p.UpdatedAt); err != nil {
			return err
		}
		profiles = append(profiles, p)
	}
	if err := m.remote.PushProfiles(ctx, m.clientCode, profiles); err != nil {
		return err
	}

	// 3. Push Jobs
	jRows, err := m.localDB.Query("SELECT name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, last_run FROM backup_jobs")
	if err != nil {
		return err
	}
	defer jRows.Close()

	var jobs []SyncJob
	for jRows.Next() {
		var j SyncJob
		var enabled int
		var lastRun sql.NullTime
		if err := jRows.Scan(&j.Name, &enabled, &j.Schedule, &j.Mode, &j.OutputMode, &j.IncludeFile, &j.ProfileName, &j.Ticket, &j.OutputDir, &j.RetentionDays, &lastRun); err != nil {
			return err
		}
		j.Enabled = enabled == 1
		if lastRun.Valid {
			j.UpdatedAt = lastRun.Time
		} else {
			j.UpdatedAt = time.Now()
		}
		jobs = append(jobs, j)
	}
	if err := m.remote.PushJobs(ctx, m.clientCode, jobs); err != nil {
		return err
	}

	return nil
}


func (m *SyncManager) SendHeartbeat(ctx context.Context) error {
	backupDir := database.GetSetting("backup_output_base_directory")
	info, err := systeminfo.GetSystemInfo(backupDir)
	if err != nil {
		return err
	}

	hb := Heartbeat{
		ClientCode:   m.clientCode,
		ClientAlias:  database.GetSetting("client_alias"),
		OS:           runtime.GOOS + " " + runtime.GOARCH,
		Version:      version.Version,
		CPUModel:     info.CPUModel,
		MemTotal:     info.MemoryTotal,
		MemUsed:      info.MemoryUsed,
		DiskTotal:    info.DiskTotal,
		DiskFree:     info.DiskFree,
		ToolVersions: systeminfo.GetToolVersions(),
		Timestamp:    time.Now(),
	}

	return m.remote.SendHeartbeat(ctx, hb)
}

func (m *SyncManager) logSyncHistory(direction, status string, changes int, errMsg string) {
	_, _ = m.localDB.Exec("INSERT INTO sync_history (direction, status, changes_count, error_msg) VALUES (?, ?, ?, ?)",
		direction, status, changes, errMsg)
}
