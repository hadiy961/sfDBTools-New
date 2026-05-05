package database

import (
	"context"
	"fmt"
	"sfdbtools/internal/domain"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
)

// connectWithSpinner adalah helper untuk membuat koneksi dengan spinner UI.
func connectWithSpinner(info domain.DBInfo, database, label string, timeout time.Duration) (*Client, error) {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" Menghubungkan ke database %s %s:%d...", label, info.Host, info.Port)
	spin.Start()
	defer spin.Stop()

	cfg := Config{
		Host:                 info.Host,
		Port:                 info.Port,
		User:                 info.User,
		Password:             info.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             database,
		ReadTimeout:          0,
		WriteTimeout:         0,
	}

	client, err := NewClient(context.Background(), cfg, timeout, 10, 5, 0)
	if err != nil {
		return nil, fmt.Errorf("koneksi database %s gagal: %w", label, err)
	}
	return client, nil
}

func ConnectToSourceDatabase(creds domain.SourceDBConnection) (*Client, error) {
	return connectWithSpinner(creds.DBInfo, creds.Database, "sumber", 10*time.Second)
}

func ConnectToDestinationDatabase(creds domain.DestinationDBConnection) (*Client, error) {
	return connectWithSpinner(creds.DBInfo, creds.Database, "tujuan", 5*time.Second)
}

// ConnectionTest - Menguji koneksi database berdasarkan informasi yang diberikan
func ConnectToRemoteHub() (*Client, error) {
	syncType := GetSetting("sync_type")
	host := GetSetting("sync_host")
	portStr := GetSetting("sync_port")
	user := GetSetting("sync_user")
	pass := GetSetting("sync_password")
	dbName := GetSetting("sync_database")
	
	port, _ := strconv.Atoi(portStr)

	if syncType == "postgres" || syncType == "supabase" {
		cfg := PostgresConfig{
			Host:     host,
			Port:     port,
			User:     user,
			Password: pass,
			Database: dbName,
		}
		return NewPostgresClient(context.Background(), cfg, 5*time.Second)
	}

	// Default to MySQL/MariaDB
	cfg := Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		Database: dbName,
	}
	return NewClient(context.Background(), cfg, 5*time.Second, 5, 2, 0)
}
