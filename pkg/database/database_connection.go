package database

import (
	"context"
	"fmt"
	"sfDBTools/internal/services/log"
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

// connectWithSpinner adalah helper untuk membuat koneksi dengan spinner UI.
func connectWithSpinner(info types.DBInfo, database, label string, timeout time.Duration) (*Client, error) {
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

func ConnectToSourceDatabase(creds types.SourceDBConnection) (*Client, error) {
	return connectWithSpinner(creds.DBInfo, creds.Database, "sumber", 10*time.Second)
}

func ConnectToDestinationDatabase(creds types.DestinationDBConnection) (*Client, error) {
	return connectWithSpinner(creds.DBInfo, creds.Database, "tujuan", 5*time.Second)
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
