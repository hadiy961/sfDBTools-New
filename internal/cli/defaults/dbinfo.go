package defaultVal

import (
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/envx"
	"sfdbtools/pkg/consts"
)

// DefaultDBInfo mengembalikan nilai default untuk koneksi database
func DefaultDBInfo() domain.DBInfo {
	return domain.DBInfo{
		Host:     envx.GetEnvOrDefault(consts.ENV_TARGET_DB_HOST, "test-db-host"),
		Port:     envx.GetEnvOrDefaultInt(consts.ENV_TARGET_DB_PORT, 3306),
		User:     envx.GetEnvOrDefault(consts.ENV_TARGET_DB_USER, "root"),
		Password: envx.GetEnvOrDefault(consts.ENV_TARGET_DB_PASSWORD, ""),
	}
}
