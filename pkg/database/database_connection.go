package database

import (
	"context"
	"fmt"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"time"

	"github.com/briandowns/spinner"
)

// ConnectToAppDatabase membuat koneksi ke database aplikasi berdasarkan environment variables.
func ConnectToAppDatabase() (*Client, error) {
	host := helper.GetEnvOrDefault(consts.ENV_DB_HOST, "localhost")
	port := helper.GetEnvOrDefaultInt(consts.ENV_DB_PORT, 3306)
	user := helper.GetEnvOrDefault(consts.ENV_DB_USER, "root")
	password := helper.GetEnvOrDefault(consts.ENV_DB_PASSWORD, "DataOn24!!")
	database := helper.GetEnvOrDefault(consts.ENV_DB_NAME, "sfDBTools")

	cfg := Config{
		Host:                 host,
		Port:                 port,
		User:                 user,
		Password:             password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             database,
	}

	if cfg.Host == "" || cfg.Port == 0 {
		return nil, fmt.Errorf("invalid database configuration, please check your environment variables") // Tidak ada koneksi database yang diatur
	}

	ctx := context.Background()
	client, err := NewClient(ctx, cfg, 5*time.Second, 10, 5, time.Minute*5)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ConnectToSourceDatabase(creds types.SourceDBConnection) (*Client, error) {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" Menghubungkan ke database sumber %s:%d...", creds.DBInfo.Host, creds.DBInfo.Port)
	spin.Start()

	cfg := Config{
		Host:                 creds.DBInfo.Host,
		Port:                 creds.DBInfo.Port,
		User:                 creds.DBInfo.User,
		Password:             creds.DBInfo.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             creds.Database,
	}

	ctx := context.Background()
	client, err := NewClient(ctx, cfg, 10*time.Second, 10, 5, 0) // Timeout lebih lama
	spin.Stop()

	if err != nil {
		return nil, fmt.Errorf("koneksi database sumber gagal: %w", err)
	}

	return client, nil
}

func ConnectToDestinationDatabase(creds types.DestinationDBConnection) (*Client, error) {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" Menghubungkan ke database tujuan %s:%d...", creds.DBInfo.Host, creds.DBInfo.Port)
	spin.Start()

	cfg := Config{
		Host:                 creds.DBInfo.Host,
		Port:                 creds.DBInfo.Port,
		User:                 creds.DBInfo.User,
		Password:             creds.DBInfo.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             creds.Database,
	}

	ctx := context.Background()
	// Untuk destination database (backup/restore), gunakan ConnMaxLifetime 0 (unlimited)
	// karena operasi bisa memakan waktu lama
	client, err := NewClient(ctx, cfg, 5*time.Second, 10, 5, 0)
	spin.Stop()

	if err != nil {
		return nil, fmt.Errorf("koneksi database tujuan gagal: %w", err)
	}

	return client, nil
}

// ConnectionTest - Menguji koneksi database berdasarkan informasi yang diberikan
func ConnectionTest(dbInfo *types.DBInfo, applog applog.Logger) error {
	applog.Info("Memeriksa koneksi database ke " + dbInfo.Host + ":" + fmt.Sprintf("%d", dbInfo.Port) + "...")
	connectionInfo := types.DestinationDBConnection{
		DBInfo:   *dbInfo,
		Database: "mysql", // Tidak perlu database spesifik untuk tes koneksi
	}
	client, err := ConnectToDestinationDatabase(connectionInfo)
	if err != nil {
		applog.Error(err.Error())
		return err
	}
	defer client.db.Close()
	applog.Info("Koneksi database ke " + dbInfo.Host + ":" + fmt.Sprintf("%d", dbInfo.Port) + " berhasil.")
	return nil
}
