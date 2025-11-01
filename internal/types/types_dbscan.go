package types

// ScanAllDBOptions berisi opsi untuk scan semua database
type ScanAllDBOptions struct {
	ProfileInfo   ProfileInfo
	ExcludeSystem bool
	SaveToDB      bool
	Background    bool // Jalankan scanning di background
}

// ScanEntryConfig untuk konfigurasi scan entry point
type ScanEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	SuccessMsg  string
	LogPrefix   string
	Mode        string // "all" atau "database"
}

// ScanResult berisi hasil scanning
type ScanResult struct {
	TotalDatabases int
	SuccessCount   int
	FailedCount    int
	Duration       string
	Errors         []string
}

// DatabaseFilterStats menyimpan statistik hasil filtering database.
type DatabaseFilterStats struct {
	TotalFound     int
	ToScan         int
	ExcludedSystem int
	ExcludedByList int
	ExcludedByFile int // Merepresentasikan database yang tidak ada di include list
	ExcludedEmpty  int
}

// ScanOptions berisi opsi untuk database scan
type ScanOptions struct {
	// Database Configuration
	ProfileInfo ProfileInfo
	LocalScan   bool
	// Encryption
	Encryption struct {
		Key string
	}

	// Database Selection
	DatabaseList struct {
		File      string
		Databases []string
		UseFile   bool
	}

	// Source Database untuk mode single
	SourceDatabase string

	// Filter Options
	ExcludeSystem bool
	IncludeList   []string
	ExcludeList   []string

	// Target Database untuk menyimpan hasil scan
	TargetDB struct {
		Host     string
		Port     int
		User     string
		Password string
		Database string
	}

	// Output Options
	DisplayResults bool
	SaveToDB       bool
	Background     bool // Jalankan scanning di background

	// Internal use only
	Mode        string // "all" atau "selection" atau "single" atau "rescan"
	ShowOptions bool   // Tampilkan opsi scanning sebelum eksekusi
}

// SystemDatabases adalah canonical list dari database sistem MySQL/MariaDB
// Menggunakan map untuk O(1) lookup performance
var SystemDatabases = map[string]struct{}{
	"information_schema": {},
	"mysql":              {},
	"performance_schema": {},
	"sys":                {},
	"innodb":             {},
}
