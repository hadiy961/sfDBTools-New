// File : pkg/profilehelper/profilehelper_connection.go
// Deskripsi : Helper functions untuk database connection menggunakan ProfileInfo
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025

package profilehelper

import (
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
)

// ConnectWithProfile membuat koneksi database menggunakan ProfileInfo
// Ini adalah wrapper untuk ConnectToSourceDatabase yang lebih sederhana
func ConnectWithProfile(profile *types.ProfileInfo, initialDB string) (*database.Client, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile tidak boleh nil")
	}

	if initialDB == "" {
		initialDB = "mysql" // default ke system database
	}

	creds := types.SourceDBConnection{
		DBInfo:   profile.DBInfo,
		Database: initialDB,
	}

	client, err := database.ConnectToSourceDatabase(creds)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke %s@%s:%d: %w",
			profile.DBInfo.User, profile.DBInfo.Host, profile.DBInfo.Port, err)
	}

	return client, nil
}
