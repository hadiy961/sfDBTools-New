package consts

const (
	// App Config
	ENV_APPS_CONFIG = "SFDB_APPS_CONFIG"

	// Database Connection
	ENV_DB_HOST        = "SFDB_DB_HOST"
	ENV_DB_PORT        = "SFDB_DB_PORT"
	ENV_DB_USER        = "SFDB_DB_USER"
	ENV_DB_PASSWORD    = "SFDB_DB_PASSWORD"
	ENV_DB_NAME        = "SFDB_DB_NAME"
	ENV_SOURCE_DB_NAME = "SFDB_SOURCE_DB_NAME"
	ENV_DEST_DB_NAME   = "SFDB_DEST_DB_NAME"
	ENV_APP_DB_NAME    = "SFDB_APP_DB_NAME"

	// Encryption Key
	ENV_SOURCE_PROFILE_KEY = "SFDB_SOURCE_PROFILE_KEY"
	ENV_TARGET_PROFILE_KEY = "SFDB_TARGET_PROFILE_KEY"
	// Quiet mode flag to suppress banners/logs on stdout for piping
	ENV_QUIET = "SFDB_QUIET"
	// Generic encryption key for utility commands
	ENV_ENCRYPTION_KEY = "SFDB_ENCRYPTION_KEY"
	// Script bundle encryption key
	ENV_SCRIPT_KEY = "SFDB_SCRIPT_KEY"
	// Backup encryption key
	ENV_BACKUP_ENCRYPTION_KEY = "SFDB_BACKUP_ENCRYPTION_KEY"

	// Other constants can be added here as needed

	ENV_TARGET_DB_HOST     = "SFDB_TARGET_DB_HOST"
	ENV_TARGET_DB_PORT     = "SFDB_TARGET_DB_PORT"
	ENV_TARGET_DB_USER     = "SFDB_TARGET_DB_USER"
	ENV_TARGET_DB_PASSWORD = "SFDB_TARGET_DB_PASSWORD"

	ENV_SOURCE_DB_HOST     = "SFDB_SOURCE_DB_HOST"
	ENV_SOURCE_DB_PORT     = "SFDB_SOURCE_DB_PORT"
	ENV_SOURCE_DB_USER     = "SFDB_SOURCE_DB_USER"
	ENV_SOURCE_DB_PASSWORD = "SFDB_SOURCE_DB_PASSWORD"

	ENV_SOURCE_PROFILE = "SFDB_SOURCE_PROFILE"
	ENV_TARGET_PROFILE = "SFDB_TARGET_PROFILE"

	// SSH
	// Jika diset 1, sfdbtools akan mengabaikan verifikasi host key SSH (TIDAK AMAN).
	// Default: verifikasi host key wajib (secure-by-default).
	ENV_SSH_INSECURE_IGNORE_HOSTKEY = "SFDB_SSH_INSECURE_IGNORE_HOSTKEY"

	// Profile Connection Timeout
	// Override timeout untuk koneksi database saat create/edit profile (format: "15s", "1m", etc.)
	// Default: 15s
	ENV_PROFILE_CONNECT_TIMEOUT = "SFDB_PROFILE_CONNECT_TIMEOUT"

	// Auto Update (GitHub Releases)
	ENV_AUTO_UPDATE       = "SFDB_AUTO_UPDATE"       // set 1 untuk enable auto-update saat start
	ENV_NO_AUTO_UPDATE    = "SFDB_NO_AUTO_UPDATE"    // set 1 untuk disable paksa (override)
	ENV_UPDATE_REPO_OWNER = "SFDB_UPDATE_REPO_OWNER" // default: hadiy961
	ENV_UPDATE_REPO_NAME  = "SFDB_UPDATE_REPO_NAME"  // default: sfdbtools-New
	ENV_GITHUB_TOKEN      = "SFDB_GITHUB_TOKEN"      // optional, untuk menghindari rate limit GitHub API

	ENV_PASSWORD_APP = "sfdbtools" // hardcoded password for app database
)
