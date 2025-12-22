// File : internal/backup/helpers/user_grants_base.go
// Deskripsi : Base functions untuk user grants - GetUserList dan GetUserGrants
// Author : Hadiyatna Muflihun
// Tanggal : 22 Desember 2024
// Last Modified : 22 Desember 2024
// Moved from: pkg/database/database_user_base.go & database_user_grants.go

package helpers

import (
	"context"
	"fmt"
	"sfDBTools/pkg/database"
	"strings"
)

// UserInfo menyimpan informasi user dari mysql.user
type UserInfo struct {
	User string
	Host string
}

// GetUserList mengambil daftar user dari mysql.user table
func GetUserList(ctx context.Context, client *database.Client) ([]UserInfo, error) {
	query := `SELECT user, host FROM mysql.user WHERE user != ''`

	rows, err := client.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan daftar user: %w", err)
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var user, host string
		if err := rows.Scan(&user, &host); err != nil {
			return nil, fmt.Errorf("gagal scan user: %w", err)
		}
		users = append(users, UserInfo{User: user, Host: host})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi rows: %w", err)
	}

	return users, nil
}

// GetUserGrants mengambil SHOW GRANTS untuk setiap user
func GetUserGrants(ctx context.Context, client *database.Client, user, host string) ([]string, error) {
	query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", user, host)

	rows, err := client.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan grants untuk %s@%s: %w", user, host, err)
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			return nil, fmt.Errorf("gagal scan grant: %w", err)
		}
		// Tambahkan semicolon jika belum ada
		if !strings.HasSuffix(grant, ";") {
			grant += ";"
		}
		grants = append(grants, grant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi grants: %w", err)
	}

	return grants, nil
}
