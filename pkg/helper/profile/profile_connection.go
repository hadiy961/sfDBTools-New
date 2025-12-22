package profile

import (
	"fmt"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
)

// ConnectWithProfile membuat koneksi database menggunakan ProfileInfo.
func ConnectWithProfile(profile *types.ProfileInfo, initialDB string) (*database.Client, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile tidak boleh nil")
	}

	if initialDB == "" {
		initialDB = consts.DefaultInitialDatabase
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
