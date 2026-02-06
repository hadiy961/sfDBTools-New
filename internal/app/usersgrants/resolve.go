package usersgrants

import (
	"context"
	"fmt"
	"sfdbtools/internal/shared/database"
)

// ResolveUserAccounts mengambil list account dari mysql.user.
// Jika opts.Users diisi, akan mengembalikan list tersebut (setelah normalisasi/unique).
func ResolveUserAccounts(ctx context.Context, client *database.Client, opts ExportOptions) ([]UserAccount, error) {
	if len(opts.Users) > 0 {
		return uniqSortedAccounts(opts.Users), nil
	}

	query := `SELECT user, host FROM mysql.user WHERE user != ''`
	rows, err := client.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan daftar user: %w", err)
	}
	defer rows.Close()

	users := make([]UserAccount, 0, 64)
	for rows.Next() {
		var user, host string
		if err := rows.Scan(&user, &host); err != nil {
			return nil, fmt.Errorf("gagal scan user: %w", err)
		}
		users = append(users, UserAccount{User: user, Host: host})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi rows: %w", err)
	}

	return uniqSortedAccounts(users), nil
}
