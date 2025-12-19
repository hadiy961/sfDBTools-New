package database

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// shouldRetry determines whether an error is transient and worth retrying.
// We keep it intentionally simple per project guidelines.
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset by peer") ||
		strings.Contains(s, "invalid connection") ||
		strings.Contains(s, "server has gone away")
}

// backoff returns a small exponential backoff duration for attempt i.
func backoff(i int) time.Duration {
	switch i {
	case 0:
		return 200 * time.Millisecond
	case 1:
		return 600 * time.Millisecond
	default:
		return 1200 * time.Millisecond
	}
}

// ExecContextWithRetry executes a statement with basic retry on transient network errors.
func (c *Client) ExecContextWithRetry(ctx context.Context, query string, args ...any) (sql.Result, error) {
	var res sql.Result
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		res, err = c.db.ExecContext(ctx, query, args...)
		if err == nil {
			return res, nil
		}
		if !shouldRetry(err) {
			return nil, err
		}
		// Try to re-establish health and back off
		_ = c.db.PingContext(ctx)
		time.Sleep(backoff(attempt))
	}
	return nil, err
}

// QueryContextWithRetry performs a query with basic retry on transient network errors.
func (c *Client) QueryContextWithRetry(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		rows, err = c.db.QueryContext(ctx, query, args...)
		if err == nil {
			return rows, nil
		}
		if !shouldRetry(err) {
			return nil, err
		}
		_ = c.db.PingContext(ctx)
		time.Sleep(backoff(attempt))
	}
	return nil, err
}
