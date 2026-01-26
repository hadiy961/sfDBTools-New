// File : internal/app/profile/helpers/parser/import_parser.go
// Deskripsi : Parser untuk import data (schema validation, row parsing) - REFACTORED with common utils
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package parser

import (
	"fmt"
	"strings"

	"sfdbtools/internal/app/profile/helpers/common"
	profilemodel "sfdbtools/internal/app/profile/model"
	profilevalidation "sfdbtools/internal/app/profile/validation"
)

// SchemaResult represents hasil validasi schema kolom
type SchemaResult struct {
	Schema   map[string]int // column name -> index mapping
	Warnings []string
}

// BuildImportSchema memvalidasi dan membuild mapping kolom dari headers
// Returns: schema map (normalized_column_name -> index), warnings, error
func BuildImportSchema(headers []string) (*SchemaResult, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("header kosong")
	}

	// Normalize headers
	norm := make([]string, 0, len(headers))
	for _, h := range headers {
		norm = append(norm, normalizeHeader(h))
	}

	// Build schema map
	schema := map[string]int{}
	for i, h := range norm {
		if common.IsEmpty(h) {
			continue
		}
		// Jika duplicate header, ambil yang pertama (warn)
		if _, ok := schema[h]; !ok {
			schema[h] = i
		}
	}

	// Validate required columns
	required := []string{"name", "host", "user", "password", "profile_key"}
	missing := make([]string, 0, len(required))
	for _, k := range required {
		if _, ok := schema[k]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("kolom wajib tidak ditemukan: %s", strings.Join(missing, ", "))
	}

	// Check for unknown columns (warning only)
	known := map[string]bool{
		"name": true, "host": true, "port": true, "user": true, "password": true, "profile_key": true,
		"ssh_enabled": true, "ssh_host": true, "ssh_port": true, "ssh_user": true, "ssh_password": true,
		"ssh_identity_file": true, "ssh_local_port": true,
	}

	warnings := []string{}
	unknownCols := []string{}
	for _, h := range norm {
		if common.IsEmpty(h) {
			continue
		}
		if !known[h] {
			unknownCols = append(unknownCols, h)
		}
	}
	if len(unknownCols) > 0 {
		warnings = append(warnings, fmt.Sprintf("Kolom tidak dikenal akan diabaikan: %s", strings.Join(common.UniqueStrings(unknownCols), ", ")))
	}

	return &SchemaResult{
		Schema:   schema,
		Warnings: warnings,
	}, nil
}

// ParseImportRows mem-parse data rows menjadi ImportRow structs dengan validasi
func ParseImportRows(dataRows [][]string, schema map[string]int) ([]profilemodel.ImportRow, []profilemodel.ImportCellError) {
	out := make([]profilemodel.ImportRow, 0, len(dataRows))
	errList := make([]profilemodel.ImportCellError, 0)

	for i, row := range dataRows {
		rowNum := i + 2 // header di row 1

		// Skip empty rows
		if common.IsEmptyRow(row) {
			continue
		}

		// Helper untuk get column value
		get := func(col string) string {
			idx, ok := schema[col]
			if !ok {
				return ""
			}
			if idx < 0 || idx >= len(row) {
				return ""
			}
			return common.Trim(row[idx])
		}

		// Extract basic fields
		name := get("name")
		host := get("host")
		user := get("user")
		password := get("password")
		profileKey := get("profile_key")

		// Validate required fields (pakai existing validator)
		profilevalidation.ValidateRequiredField(name, "name", rowNum, &errList)
		profilevalidation.ValidateRequiredField(host, "host", rowNum, &errList)
		profilevalidation.ValidateRequiredField(user, "user", rowNum, &errList)
		profilevalidation.ValidateRequiredField(password, "password", rowNum, &errList)
		profilevalidation.ValidateRequiredField(profileKey, "profile_key", rowNum, &errList)

		// Parse port (pakai common util)
		port, perr := common.ParsePort(get("port"), 3306, "port")
		if perr != nil {
			errList = append(errList, profilemodel.ImportCellError{
				Row: rowNum, Column: "port", Message: perr.Error(),
			})
			port = 3306 // fallback
		}

		// SSH fields
		sshEnabled := common.ParseBool(get("ssh_enabled"))
		sshHost := get("ssh_host")
		sshUser := get("ssh_user")
		sshPassword := get("ssh_password")
		sshIdentity := get("ssh_identity_file")

		// SSH port (pakai common util)
		sshPort, sperr := common.ParsePort(get("ssh_port"), 22, "ssh_port")
		if sperr != nil {
			errList = append(errList, profilemodel.ImportCellError{
				Row: rowNum, Column: "ssh_port", Message: sperr.Error(),
			})
			sshPort = 22
		}

		// SSH local port (allow 0 = auto-assign)
		sshLocalPort, lperr := common.ParsePortAllowZero(get("ssh_local_port"), 0, "ssh_local_port")
		if lperr != nil {
			errList = append(errList, profilemodel.ImportCellError{
				Row: rowNum, Column: "ssh_local_port", Message: lperr.Error(),
			})
			sshLocalPort = 0
		}

		// Validate SSH conditional fields (pakai existing validator)
		if sshEnabled {
			profilevalidation.ValidateRequiredField(sshHost, "ssh_host (ssh_enabled=true)", rowNum, &errList)
			profilevalidation.ValidateRequiredField(sshUser, "ssh_user (ssh_enabled=true)", rowNum, &errList)

			// Untuk SSH tunnel, minimal harus ada 1 metode autentikasi:
			// - password, atau
			// - identity file (private key)
			if common.IsEmpty(sshPassword) && common.IsEmpty(sshIdentity) {
				errList = append(errList, profilemodel.ImportCellError{
					Row:     rowNum,
					Column:  "ssh_password",
					Message: "wajib isi ssh_password atau ssh_identity_file jika ssh_enabled=true",
				})
			}
		}

		// Build ImportRow
		out = append(out, profilemodel.ImportRow{
			RowNum:      rowNum,
			Name:        name,
			Host:        host,
			Port:        port,
			User:        user,
			Password:    password,
			ProfileKey:  profileKey,
			SSHEnabled:  sshEnabled,
			SSHHost:     sshHost,
			SSHPort:     sshPort,
			SSHUser:     sshUser,
			SSHPassword: sshPassword,
			SSHIdentity: sshIdentity,
			SSHLocal:    sshLocalPort,
		})
	}

	return out, errList
}

// normalizeHeader normalizes header string untuk konsistensi
// Converts: "Profile Name" -> "profile_name"
func normalizeHeader(s string) string {
	v := common.TrimLower(s)
	v = strings.ReplaceAll(v, " ", "_")
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, ".", "_")
	return v
}
