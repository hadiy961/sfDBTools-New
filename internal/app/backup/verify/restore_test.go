package verify

// DryRunEngine adalah antarmuka untuk fitur Dry-Run Restore di Phase 2.
// Nantinya, engine ini akan membaca stream SQL dan melakukan validasi sintaks,
// atau me-restore file ke temporary database (sandbox) untuk memverifikasi integrity skema.
type DryRunEngine interface {
	// ParseAndValidate membaca stream SQL dan memvalidasi struktur tanpa side effect ke DB utama.
	ParseAndValidate(filePath string, opts CheckOptions) error
	
	// TestRestore mencoba mengembalikan backup ke sandbox DB.
	TestRestore(filePath string, opts CheckOptions) error
}

// SQLParser mendefinisikan interface parser untuk SQL statements.
type SQLParser interface {
	// Next mendapatkan SQL statement selanjutnya dari stream.
	Next() (string, error)
}
