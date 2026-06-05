package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"sfdbtools/internal/shared/database"
)

// RemoteProvider defines the interface for communicating with the Remote Hub.
type RemoteProvider interface {
	Migrate(ctx context.Context) error
	PushSettings(ctx context.Context, clientCode string, settings []SyncSetting) error
	PullSettings(ctx context.Context, clientCode string) ([]SyncSetting, error)
	PushProfiles(ctx context.Context, clientCode string, profiles []SyncProfile) error
	PullProfiles(ctx context.Context, clientCode string) ([]SyncProfile, error)
	PushJobs(ctx context.Context, clientCode string, jobs []SyncJob) error
	PullJobs(ctx context.Context, clientCode string) ([]SyncJob, error)
	PushLogs(ctx context.Context, clientCode string, logs []AuditLog) error
	SendHeartbeat(ctx context.Context, hb Heartbeat) error
	Close() error
}

// SQLRemoteProvider is a RemoteProvider implementation using SQL (MySQL/Postgres).
type SQLRemoteProvider struct {
	client *database.Client
}

func NewSQLRemoteProvider(client *database.Client) *SQLRemoteProvider {
	return &SQLRemoteProvider{client: client}
}

func (p *SQLRemoteProvider) Close() error {
	return p.client.Close()
}

func (p *SQLRemoteProvider) Migrate(ctx context.Context) error {
	var queries []string
	if p.client.Driver() == "postgres" {
		queries = []string{
			`CREATE TABLE IF NOT EXISTS sf_client_settings (client_code TEXT, s_key TEXT, s_value TEXT, category TEXT, is_locked INTEGER DEFAULT 0, updated_at TIMESTAMP, PRIMARY KEY (client_code, s_key))`,
			`CREATE TABLE IF NOT EXISTS sf_client_profiles (client_code TEXT, name TEXT, encrypted_data TEXT, account_code TEXT, updated_at TIMESTAMP, PRIMARY KEY (client_code, name))`,
			`CREATE TABLE IF NOT EXISTS sf_client_jobs (client_code TEXT, name TEXT, enabled INTEGER, schedule TEXT, mode TEXT, output_mode TEXT, include_file TEXT, profile_name TEXT, ticket TEXT, output_dir TEXT, retention_days INTEGER, updated_at TIMESTAMP, PRIMARY KEY (client_code, name))`,
			`CREATE TABLE IF NOT EXISTS sf_client_heartbeat (client_code TEXT PRIMARY KEY, client_alias TEXT, os_version TEXT, tool_version TEXT, cpu_model TEXT, mem_total BIGINT, mem_used BIGINT, disk_total BIGINT, disk_free BIGINT, tool_versions_json TEXT, updated_at TIMESTAMP)`,
			`CREATE TABLE IF NOT EXISTS sf_client_logs (id SERIAL PRIMARY KEY, client_code TEXT, event_type TEXT, details TEXT, timestamp TIMESTAMP)`,
		}
	} else {
		queries = []string{
			`CREATE TABLE IF NOT EXISTS sf_client_settings (client_code VARCHAR(100), s_key VARCHAR(100), s_value TEXT, category VARCHAR(100), is_locked TINYINT DEFAULT 0, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (client_code, s_key))`,
			`CREATE TABLE IF NOT EXISTS sf_client_profiles (client_code VARCHAR(100), name VARCHAR(100), encrypted_data TEXT, account_code VARCHAR(100), updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (client_code, name))`,
			`CREATE TABLE IF NOT EXISTS sf_client_jobs (client_code VARCHAR(100), name VARCHAR(100), enabled TINYINT, schedule VARCHAR(100), mode VARCHAR(100), output_mode VARCHAR(100), include_file VARCHAR(100), profile_name VARCHAR(100), ticket VARCHAR(100), output_dir TEXT, retention_days INT, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (client_code, name))`,
			`CREATE TABLE IF NOT EXISTS sf_client_heartbeat (client_code VARCHAR(100) PRIMARY KEY, client_alias VARCHAR(100), os_version VARCHAR(100), tool_version VARCHAR(100), cpu_model VARCHAR(255), mem_total BIGINT, mem_used BIGINT, disk_total BIGINT, disk_free BIGINT, tool_versions_json TEXT, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)`,
			`CREATE TABLE IF NOT EXISTS sf_client_logs (id INT AUTO_INCREMENT PRIMARY KEY, client_code VARCHAR(100), event_type VARCHAR(100), details TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		}
	}

	for _, q := range queries {
		if _, err := p.client.DB().ExecContext(ctx, q); err != nil {
			return fmt.Errorf("failed to migrate remote table: %w", err)
		}
	}
	return nil
}

func (p *SQLRemoteProvider) PushSettings(ctx context.Context, clientCode string, settings []SyncSetting) error {
	for _, s := range settings {
		var query string
		if p.client.Driver() == "postgres" {
			query = `INSERT INTO sf_client_settings (client_code, s_key, s_value, category, is_locked, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6)
					 ON CONFLICT(client_code, s_key) DO UPDATE SET s_value=EXCLUDED.s_value, category=EXCLUDED.category, is_locked=EXCLUDED.is_locked, updated_at=EXCLUDED.updated_at`
			_, err := p.client.DB().ExecContext(ctx, query, clientCode, s.Key, s.Value, s.Category, s.IsLocked, s.UpdatedAt)
			if err != nil {
				return err
			}
		} else {
			query = `INSERT INTO sf_client_settings (client_code, s_key, s_value, category, is_locked, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?)
					 ON DUPLICATE KEY UPDATE s_value=VALUES(s_value), category=VALUES(category), is_locked=VALUES(is_locked), updated_at=VALUES(updated_at)`
			_, err := p.client.DB().ExecContext(ctx, query, clientCode, s.Key, s.Value, s.Category, s.IsLocked, s.UpdatedAt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *SQLRemoteProvider) PullSettings(ctx context.Context, clientCode string) ([]SyncSetting, error) {
	query := `SELECT s_key, s_value, category, is_locked, updated_at FROM sf_client_settings WHERE client_code = ?`
	if p.client.Driver() == "postgres" {
		query = `SELECT s_key, s_value, category, is_locked, updated_at FROM sf_client_settings WHERE client_code = $1`
	}

	rows, err := p.client.DB().QueryContext(ctx, query, clientCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []SyncSetting
	for rows.Next() {
		var s SyncSetting
		var isLocked int
		if err := rows.Scan(&s.Key, &s.Value, &s.Category, &isLocked, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.IsLocked = isLocked == 1
		settings = append(settings, s)
	}
	return settings, nil
}

func (p *SQLRemoteProvider) PushProfiles(ctx context.Context, clientCode string, profiles []SyncProfile) error {
	for _, pr := range profiles {
		var query string
		if p.client.Driver() == "postgres" {
			query = `INSERT INTO sf_client_profiles (client_code, name, encrypted_data, account_code, updated_at)
					 VALUES ($1, $2, $3, $4, $5)
					 ON CONFLICT(client_code, name) DO UPDATE SET encrypted_data=EXCLUDED.encrypted_data, account_code=EXCLUDED.account_code, updated_at=EXCLUDED.updated_at`
			_, err := p.client.DB().ExecContext(ctx, query, clientCode, pr.Name, pr.EncryptedData, pr.AccountCode, pr.UpdatedAt)
			if err != nil {
				return err
			}
		} else {
			query = `INSERT INTO sf_client_profiles (client_code, name, encrypted_data, account_code, updated_at)
					 VALUES (?, ?, ?, ?, ?)
					 ON DUPLICATE KEY UPDATE encrypted_data=VALUES(encrypted_data), account_code=VALUES(account_code), updated_at=VALUES(updated_at)`
			_, err := p.client.DB().ExecContext(ctx, query, clientCode, pr.Name, pr.EncryptedData, pr.AccountCode, pr.UpdatedAt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *SQLRemoteProvider) PullProfiles(ctx context.Context, clientCode string) ([]SyncProfile, error) {
	query := `SELECT name, encrypted_data, account_code, updated_at FROM sf_client_profiles WHERE client_code = ?`
	if p.client.Driver() == "postgres" {
		query = `SELECT name, encrypted_data, account_code, updated_at FROM sf_client_profiles WHERE client_code = $1`
	}

	rows, err := p.client.DB().QueryContext(ctx, query, clientCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []SyncProfile
	for rows.Next() {
		var pr SyncProfile
		if err := rows.Scan(&pr.Name, &pr.EncryptedData, &pr.AccountCode, &pr.UpdatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, pr)
	}
	return profiles, nil
}

func (p *SQLRemoteProvider) PushJobs(ctx context.Context, clientCode string, jobs []SyncJob) error {
	for _, j := range jobs {
		var query string
		enabled := 0
		if j.Enabled {
			enabled = 1
		}
		if p.client.Driver() == "postgres" {
			query = `INSERT INTO sf_client_jobs (client_code, name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
					 ON CONFLICT(client_code, name) DO UPDATE SET enabled=EXCLUDED.enabled, schedule=EXCLUDED.schedule, mode=EXCLUDED.mode, output_mode=EXCLUDED.output_mode, 
					                                             include_file=EXCLUDED.include_file, profile_name=EXCLUDED.profile_name, ticket=EXCLUDED.ticket, 
					                                             output_dir=EXCLUDED.output_dir, retention_days=EXCLUDED.retention_days, updated_at=EXCLUDED.updated_at`
			_, err := p.client.DB().ExecContext(ctx, query, clientCode, j.Name, enabled, j.Schedule, j.Mode, j.OutputMode, j.IncludeFile, j.ProfileName, j.Ticket, j.OutputDir, j.RetentionDays, j.UpdatedAt)
			if err != nil {
				return err
			}
		} else {
			query = `INSERT INTO sf_client_jobs (client_code, name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, updated_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
					 ON DUPLICATE KEY UPDATE enabled=VALUES(enabled), schedule=VALUES(schedule), mode=VALUES(mode), output_mode=VALUES(output_mode), 
					                         include_file=VALUES(include_file), profile_name=VALUES(profile_name), ticket=VALUES(ticket), 
					                         output_dir=VALUES(output_dir), retention_days=VALUES(retention_days), updated_at=VALUES(updated_at)`
			_, err := p.client.DB().ExecContext(ctx, query, clientCode, j.Name, enabled, j.Schedule, j.Mode, j.OutputMode, j.IncludeFile, j.ProfileName, j.Ticket, j.OutputDir, j.RetentionDays, j.UpdatedAt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *SQLRemoteProvider) PullJobs(ctx context.Context, clientCode string) ([]SyncJob, error) {
	query := `SELECT name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, updated_at FROM sf_client_jobs WHERE client_code = ?`
	if p.client.Driver() == "postgres" {
		query = `SELECT name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, updated_at FROM sf_client_jobs WHERE client_code = $1`
	}

	rows, err := p.client.DB().QueryContext(ctx, query, clientCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []SyncJob
	for rows.Next() {
		var j SyncJob
		var enabled int
		if err := rows.Scan(&j.Name, &enabled, &j.Schedule, &j.Mode, &j.OutputMode, &j.IncludeFile, &j.ProfileName, &j.Ticket, &j.OutputDir, &j.RetentionDays, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.Enabled = enabled == 1
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (p *SQLRemoteProvider) PushLogs(ctx context.Context, clientCode string, logs []AuditLog) error {
	for _, l := range logs {
		var query string
		if p.client.Driver() == "postgres" {
			query = `INSERT INTO sf_client_logs (client_code, event_type, details, timestamp) VALUES ($1, $2, $3, $4)`
		} else {
			query = `INSERT INTO sf_client_logs (client_code, event_type, details, timestamp) VALUES (?, ?, ?, ?)`
		}
		_, err := p.client.DB().ExecContext(ctx, query, clientCode, l.EventType, l.Details, l.Timestamp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *SQLRemoteProvider) SendHeartbeat(ctx context.Context, hb Heartbeat) error {
	versionsJSON, _ := json.Marshal(hb.ToolVersions)
	var query string
	if p.client.Driver() == "postgres" {
		query = `INSERT INTO sf_client_heartbeat (client_code, client_alias, os_version, tool_version, cpu_model, mem_total, mem_used, disk_total, disk_free, tool_versions_json, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				 ON CONFLICT(client_code) DO UPDATE SET client_alias=EXCLUDED.client_alias, os_version=EXCLUDED.os_version, tool_version=EXCLUDED.tool_version, 
				                                         cpu_model=EXCLUDED.cpu_model, mem_total=EXCLUDED.mem_total, mem_used=EXCLUDED.mem_used, 
				                                         disk_total=EXCLUDED.disk_total, disk_free=EXCLUDED.disk_free, tool_versions_json=EXCLUDED.tool_versions_json, updated_at=EXCLUDED.updated_at`
		_, err := p.client.DB().ExecContext(ctx, query, hb.ClientCode, hb.ClientAlias, hb.OS, hb.Version, hb.CPUModel, hb.MemTotal, hb.MemUsed, hb.DiskTotal, hb.DiskFree, string(versionsJSON), hb.Timestamp)
		return err
	} else {
		query = `INSERT INTO sf_client_heartbeat (client_code, client_alias, os_version, tool_version, cpu_model, mem_total, mem_used, disk_total, disk_free, tool_versions_json, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				 ON DUPLICATE KEY UPDATE client_alias=VALUES(client_alias), os_version=VALUES(os_version), tool_version=VALUES(tool_version), 
				                         cpu_model=VALUES(cpu_model), mem_total=VALUES(mem_total), mem_used=VALUES(mem_used), 
				                         disk_total=VALUES(disk_total), disk_free=VALUES(disk_free), tool_versions_json=VALUES(tool_versions_json), updated_at=VALUES(updated_at)`
		_, err := p.client.DB().ExecContext(ctx, query, hb.ClientCode, hb.ClientAlias, hb.OS, hb.Version, hb.CPUModel, hb.MemTotal, hb.MemUsed, hb.DiskTotal, hb.DiskFree, string(versionsJSON), hb.Timestamp)
		return err
	}
}
