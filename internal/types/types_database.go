package types

import "time"

// DBInfo - Struct to hold database connection details
type DBInfo struct {
	Host     string
	HostName string
	Port     int
	User     string
	Password string
	Version  string // mariadb or mysql
}

type SourceDBConnection struct {
	DBInfo   DBInfo
	Database string
}

type DestinationDBConnection struct {
	DBInfo   DBInfo
	Database string
}

type AppDBConnection struct {
	DBInfo   DBInfo
	Database string
}

// ServerDBConnection menyimpan kredensial server database tujuan sederhana
// Digunakan untuk menampilkan/merangkai koneksi target (app database)
type ServerDBConnection struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// FilterOptions berisi opsi untuk filtering database
type FilterOptions struct {
	ExcludeSystem    bool     // Exclude system databases (information_schema, mysql, etc)
	ExcludeDatabases []string // Blacklist - database yang harus di-exclude
	IncludeDatabases []string // Whitelist - hanya database ini yang diizinkan (priority tertinggi)
	IncludeFile      string   // Path ke file berisi whitelist database (satu per baris)
	ExcludeDBFile    string   // Path ke file berisi blacklist database (satu per baris)
	ExcludeData      bool     // Exclude databases with no data
	ExcludeEmpty     bool     // Exclude databases with empty data
	IsFilterCommand  bool     // Flag untuk menandai apakah ini dari command filter (untuk multi-select logic)
}

// FilterStats berisi statistik hasil filtering
type FilterStats struct {
	TotalFound          int      // Total database yang ditemukan
	TotalIncluded       int      // Total database yang included (hasil akhir)
	TotalExcluded       int      // Total database yang excluded
	ExcludedSystem      int      // Excluded karena system database
	ExcludedByList      int      // Excluded karena ada di blacklist
	ExcludedByFile      int      // Excluded karena tidak ada di whitelist file
	ExcludedEmpty       int      // Excluded karena nama kosong
	ExcludedDatabases   []string // List database yang dikecualikan (untuk metadata)
	NotFoundInInclude   []string // Database di include list yang tidak ditemukan di server
	NotFoundInExclude   []string // Database di exclude list yang tidak ditemukan di server
	NotFoundInWhitelist []string // Database di whitelist file yang tidak ditemukan di server
	NotFoundInBlacklist []string // Database di blacklist file (exclude file) yang tidak ditemukan di server
}

// DatabaseDetail menyimpan informasi detail database
type DatabaseDetail struct {
	DatabaseName   string    `db:"database_name"`
	SizeBytes      int64     `db:"size_bytes"`
	SizeHuman      string    `db:"size_human"`
	TableCount     int       `db:"table_count"`
	ProcedureCount int       `db:"procedure_count"`
	FunctionCount  int       `db:"function_count"`
	ViewCount      int       `db:"view_count"`
	UserGrantCount int       `db:"user_grant_count"`
	CollectionTime time.Time `db:"collection_time"`
	ErrorMessage   *string   `db:"error_message"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// CompressionOptions menyimpan opsi kompresi (shared across backup/restore)
type CompressionOptions struct {
	Enabled bool
	Type    string
	Level   int
}

// EncryptionOptions menyimpan opsi enkripsi (shared across backup/restore)
type EncryptionOptions struct {
	Enabled bool
	Key     string
}
