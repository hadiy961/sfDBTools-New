package defaultVal

import (
	"sfdbtools/internal/domain"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/helper"
)

// DefaultDBInfo mengembalikan nilai default untuk koneksi database
func DefaultDBInfo() domain.DBInfo {
	return domain.DBInfo{
		Host:     helper.GetEnvOrDefault(consts.ENV_TARGET_DB_HOST, "test-db-host"),
		Port:     helper.GetEnvOrDefaultInt(consts.ENV_TARGET_DB_PORT, 3306),
		User:     helper.GetEnvOrDefault(consts.ENV_TARGET_DB_USER, "root"),
		Password: helper.GetEnvOrDefault(consts.ENV_TARGET_DB_PASSWORD, ""),
	}
}
