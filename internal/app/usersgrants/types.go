package usersgrants

// UserAccount merepresentasikan MySQL account: 'user'@'host'.
type UserAccount struct {
	User string
	Host string
}

type ExportOptions struct {
	// Users: jika diisi, hanya export untuk user ini (bypass pemilihan dari mysql.user).
	Users []UserAccount

	// Databases: jika diisi, hanya export grants yang relevan dengan database ini.
	// Catatan: kita tetap perlu SHOW GRANTS per user untuk filter.
	Databases []string

	// ExcludeSystemUsers: jika true, user yang ada di SystemUsers akan di-skip.
	ExcludeSystemUsers bool
	SystemUsers        []string // spec format: "user" atau "user@host"

	IncludeCreateUser bool // best-effort: SHOW CREATE USER jika bisa, fallback minimal jika tidak
	IncludeGrants     bool // include GRANT statements; jika false, hanya export CREATE USER (jika IncludeCreateUser true)
	FlushPrivileges   bool // default true disarankan
}

type ExportStats struct {
	TotalUsersInput   int
	TotalUsersSkipped int
	TotalUsersWritten int
	TotalGrantLines   int
	Warnings          int
}
