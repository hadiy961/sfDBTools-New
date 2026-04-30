package sync

import (
	"context"
	"fmt"
)

// PushSettings sends local settings to the remote hub.
func (s *Service) PushSettings(ctx context.Context) error {
	rows, err := s.localDB.QueryContext(ctx, "SELECT key, value, category FROM app_settings")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value, category string
		if err := rows.Scan(&key, &value, &category); err != nil {
			return err
		}

		query := `INSERT INTO sf_client_settings (client_code, s_key, s_value, category, updated_at) 
				  VALUES (?, ?, ?, ?, NOW())
				  ON DUPLICATE KEY UPDATE s_value = VALUES(s_value), category = VALUES(category), updated_at = NOW()`
		
		if s.remoteDB.Driver() == "postgres" {
			query = `INSERT INTO sf_client_settings (client_code, s_key, s_value, category, updated_at) 
					 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
					 ON CONFLICT(client_code, s_key) DO UPDATE SET s_value = EXCLUDED.s_value, category = EXCLUDED.category, updated_at = CURRENT_TIMESTAMP`
		}
		_, err = s.remoteDB.DB().ExecContext(ctx, query, s.clientID, key, value, category)
		if err != nil {
			return fmt.Errorf("failed to push setting %s: %w", key, err)
		}
	}
	return nil
}

// PushJobs sends local backup jobs to the remote hub.
func (s *Service) PushJobs(ctx context.Context) error {
	rows, err := s.localDB.QueryContext(ctx, "SELECT name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days FROM backup_jobs")
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

		query := `INSERT INTO sf_client_jobs (client_code, name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, updated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
				  ON DUPLICATE KEY UPDATE enabled=VALUES(enabled), schedule=VALUES(schedule), mode=VALUES(mode), output_mode=VALUES(output_mode),
										  include_file=VALUES(include_file), profile_name=VALUES(profile_name), ticket=VALUES(ticket), 
										  output_dir=VALUES(output_dir), retention_days=VALUES(retention_days), updated_at=NOW()`
		
		if s.remoteDB.Driver() == "postgres" {
			query = `INSERT INTO sf_client_jobs (client_code, name, enabled, schedule, mode, output_mode, include_file, profile_name, ticket, output_dir, retention_days, updated_at) 
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
					 ON CONFLICT(client_code, name) DO UPDATE SET enabled=EXCLUDED.enabled, schedule=EXCLUDED.schedule, mode=EXCLUDED.mode, 
																    output_mode=EXCLUDED.output_mode, include_file=EXCLUDED.include_file, 
																    profile_name=EXCLUDED.profile_name, ticket=EXCLUDED.ticket, 
																    output_dir=EXCLUDED.output_dir, retention_days=EXCLUDED.retention_days, 
																    updated_at=CURRENT_TIMESTAMP`
		}
		_, err = s.remoteDB.DB().ExecContext(ctx, query, s.clientID, name, enabled, sched, mode, outMode, incFile, prof, ticket, outDir, retDays)
		if err != nil {
			return err
		}
	}
	return nil
}

// PushProfiles sends database profiles to the remote hub.
func (s *Service) PushProfiles(ctx context.Context) error {
	rows, err := s.localDB.QueryContext(ctx, "SELECT name, encrypted_data, account_code FROM profiles")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, data, code string
		if err := rows.Scan(&name, &data, &code); err != nil {
			return err
		}

		query := `INSERT INTO sf_client_profiles (client_code, name, encrypted_data, account_code, updated_at) 
				  VALUES (?, ?, ?, ?, NOW())
				  ON DUPLICATE KEY UPDATE encrypted_data=VALUES(encrypted_data), account_code=VALUES(account_code), updated_at=NOW()`
		
		if s.remoteDB.Driver() == "postgres" {
			query = `INSERT INTO sf_client_profiles (client_code, name, encrypted_data, account_code, updated_at) 
					 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
					 ON CONFLICT(client_code, name) DO UPDATE SET encrypted_data=EXCLUDED.encrypted_data, account_code=EXCLUDED.account_code, updated_at=CURRENT_TIMESTAMP`
		}
		_, err = s.remoteDB.DB().ExecContext(ctx, query, s.clientID, name, data, code)
		if err != nil {
			return err
		}
	}
	return nil
}
