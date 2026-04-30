package sync

import (
	"context"
)

// PullSettings retrieves remote settings from the hub.
func (s *Service) PullSettings(ctx context.Context) error {
	query := "SELECT s_key, s_value, category FROM sf_client_settings WHERE client_code = ?"
	if s.remoteDB.Driver() == "postgres" {
		query = "SELECT s_key, s_value, category FROM sf_client_settings WHERE client_code = $1"
	}

	rows, err := s.remoteDB.DB().QueryContext(ctx, query, s.clientID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value, category string
		if err := rows.Scan(&key, &value, &category); err != nil {
			return err
		}

		_, err = s.localDB.ExecContext(ctx, `INSERT OR REPLACE INTO app_settings (key, value, category, updated_at) 
											VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, key, value, category)
		if err != nil {
			return err
		}
	}
	return nil
}

// PullJobs retrieves backup jobs from the remote hub.
func (s *Service) PullJobs(ctx context.Context) error {
	query := "SELECT name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days FROM sf_client_jobs WHERE client_code = ?"
	if s.remoteDB.Driver() == "postgres" {
		query = "SELECT name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days FROM sf_client_jobs WHERE client_code = $1"
	}

	rows, err := s.remoteDB.DB().QueryContext(ctx, query, s.clientID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, sched, mode, outMode, incFile, prof, ticket, outDir string
		var enabled, retDays int
		if err := rows.Scan(&name, &enabled, &sched, &mode, &outMode, &incFile, &prof, &ticket, &outDir, &retDays); err != nil {
			return err
		}

		_, err = s.localDB.ExecContext(ctx, `INSERT OR REPLACE INTO backup_jobs (name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days) 
											VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, name, enabled, sched, mode, outMode, incFile, prof, ticket, outDir, retDays)
		if err != nil {
			return err
		}
	}
	return nil
}

// PullProfiles retrieves database profiles from the remote hub.
func (s *Service) PullProfiles(ctx context.Context) error {
	query := "SELECT name, encrypted_data, account_code FROM sf_client_profiles WHERE client_code = ?"
	if s.remoteDB.Driver() == "postgres" {
		query = "SELECT name, encrypted_data, account_code FROM sf_client_profiles WHERE client_code = $1"
	}

	rows, err := s.remoteDB.DB().QueryContext(ctx, query, s.clientID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, data, code string
		if err := rows.Scan(&name, &data, &code); err != nil {
			return err
		}

		_, err = s.localDB.ExecContext(ctx, `INSERT OR REPLACE INTO profiles (name, encrypted_data, account_code, created_at) 
											VALUES (?, ?, ?, CURRENT_TIMESTAMP)`, name, data, code)
		if err != nil {
			return err
		}
	}
	return nil
}
