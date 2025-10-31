package types

// DBInfo - Struct to hold database connection details
type DBInfo struct {
	Host     string
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
}

// FilterStats berisi statistik hasil filtering
type FilterStats struct {
	TotalFound     int // Total database yang ditemukan
	TotalIncluded  int // Total database yang included (hasil akhir)
	TotalExcluded  int // Total database yang excluded
	ExcludedSystem int // Excluded karena system database
	ExcludedByList int // Excluded karena ada di blacklist
	ExcludedByFile int // Excluded karena tidak ada di whitelist file
	ExcludedEmpty  int // Excluded karena nama kosong
}
