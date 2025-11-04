package types

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
	ExcludeUser      bool     // Exclude databases user
	ExcludeData      bool     // Exclude databases with no data
	ExcludeEmpty     bool     // Exclude databases with empty data
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
	NotFoundInInclude   []string // Database di include list yang tidak ditemukan di server
	NotFoundInExclude   []string // Database di exclude list yang tidak ditemukan di server
	NotFoundInWhitelist []string // Database di whitelist file yang tidak ditemukan di server
	NotFoundInBlacklist []string // Database di blacklist file (exclude file) yang tidak ditemukan di server
}

// DatabaseBackupInfo berisi informasi database yang berhasil dibackup
type DatabaseBackupInfo struct {
	DatabaseName        string `json:"database_name"`
	OutputFile          string `json:"output_file"`
	FileSize            int64  `json:"file_size_bytes"`        // Ukuran file backup actual (compressed)
	FileSizeHuman       string `json:"file_size_human"`        // Ukuran file backup actual (human-readable)
	OriginalDBSize      int64  `json:"original_db_size_bytes"` // Ukuran database asli (sebelum backup)
	OriginalDBSizeHuman string `json:"original_db_size_human"` // Ukuran database asli (human-readable)
	Duration            string `json:"duration"`
	Status              string `json:"status"`                   // "success", "success_with_warnings", "failed"
	Warnings            string `json:"warnings,omitempty"`       // Warning/error messages dari mysqldump
	ErrorLogFile        string `json:"error_log_file,omitempty"` // Path ke file log error
}
