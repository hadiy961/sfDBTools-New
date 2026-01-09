package database

import (
	"context"
	"fmt"
	"strings"
)

func escapeMySQLStringLiteral(s string) string {
	// Minimal escaping for single-quoted MySQL string literal.
	// MySQL (default) supports backslash escapes.
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func quoteMySQLStringLiteral(s string) string {
	return "'" + escapeMySQLStringLiteral(s) + "'"
}

func quoteMySQLAccount(user, host string) string {
	return quoteMySQLStringLiteral(user) + "@" + quoteMySQLStringLiteral(host)
}

// CheckUserExists mengecek apakah user account sudah ada.
func (c *Client) CheckUserExists(ctx context.Context, user, host string) (bool, error) {
	query := "SELECT 1 FROM mysql.user WHERE User = ? AND Host = ? LIMIT 1"
	rows, err := c.QueryContextWithRetry(ctx, query, user, host)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

// DropUserIfExists menghapus user account jika ada.
func (c *Client) DropUserIfExists(ctx context.Context, user, host string) error {
	q := fmt.Sprintf("DROP USER IF EXISTS %s", quoteMySQLAccount(user, host))
	_, err := c.ExecContextWithRetry(ctx, q)
	return err
}

// CreateUser membuat user (jika belum ada) dengan password plaintext.
func (c *Client) CreateUser(ctx context.Context, user, host, password string) error {
	q := fmt.Sprintf("CREATE USER IF NOT EXISTS %s IDENTIFIED BY %s", quoteMySQLAccount(user, host), quoteMySQLStringLiteral(password))
	_, err := c.ExecContextWithRetry(ctx, q)
	return err
}

// GrantAllPrivilegesOnDatabase memberikan ALL PRIVILEGES ke db.* untuk user.
func (c *Client) GrantAllPrivilegesOnDatabase(ctx context.Context, user, host, dbName string, withGrantOption bool) error {
	grantOpt := ""
	if withGrantOption {
		grantOpt = " WITH GRANT OPTION"
	}
	q := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO %s%s", strings.ReplaceAll(dbName, "`", "``"), quoteMySQLAccount(user, host), grantOpt)
	_, err := c.ExecContextWithRetry(ctx, q)
	return err
}
