// File : internal/app/backup/model/errors.go
// Deskripsi : Custom error types untuk backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 20 Januari 2026
// Last Modified : 20 Januari 2026
package model

import "errors"

// Sentinel errors untuk backup operations
var (
	// ErrInvalidRetryState menandakan retry dipanggil dengan state yang invalid
	ErrInvalidRetryState = errors.New("retry state invalid")

	// ErrNoDatabaseFound menandakan tidak ada database yang ditemukan
	ErrNoDatabaseFound = errors.New("no database found")

	// ErrNoDatabaseSelected menandakan tidak ada database yang dipilih
	ErrNoDatabaseSelected = errors.New("no database selected")

	// ErrProfileRequired menandakan profile wajib diisi
	ErrProfileRequired = errors.New("profile required")

	// ErrTicketRequired menandakan ticket wajib diisi
	ErrTicketRequired = errors.New("ticket required")

	// ErrDatabaseNameRequired menandakan database name wajib diisi
	ErrDatabaseNameRequired = errors.New("database name required")

	// ErrInvalidInput menandakan input tidak valid
	ErrInvalidInput = errors.New("invalid input")

	// ErrOperationCancelled menandakan operasi dibatalkan oleh user
	ErrOperationCancelled = errors.New("operation cancelled")

	// ErrNoSchedulerJobs menandakan tidak ada scheduler jobs yang aktif
	ErrNoSchedulerJobs = errors.New("no scheduler jobs configured")

	// ErrJobNameRequired menandakan job name wajib diisi
	ErrJobNameRequired = errors.New("job name required")

	// ErrRootRequired menandakan operasi membutuhkan root access
	ErrRootRequired = errors.New("root access required")

	// ErrConfigNotAvailable menandakan config belum tersedia
	ErrConfigNotAvailable = errors.New("config not available")

	// ErrConnectionNotAvailable menandakan koneksi database belum tersedia
	ErrConnectionNotAvailable = errors.New("database connection not available")

	// ErrPathGeneratorNotAvailable menandakan path generator tidak tersedia
	ErrPathGeneratorNotAvailable = errors.New("path generator not available")

	// ErrMetadataIsNil menandakan metadata object is nil
	ErrMetadataIsNil = errors.New("metadata is nil")

	// ErrNoUserFound menandakan tidak ada user yang ditemukan
	ErrNoUserFound = errors.New("no user found")

	// ErrEmptyDatabaseList menandakan daftar database kosong
	ErrEmptyDatabaseList = errors.New("database list is empty")

	// ErrBinlogNotEnabled menandakan binary logging tidak diaktifkan
	ErrBinlogNotEnabled = errors.New("binary logging not enabled")

	// ErrGTIDPosNull menandakan BINLOG_GTID_POS mengembalikan NULL
	ErrGTIDPosNull = errors.New("BINLOG_GTID_POS returned NULL")

	// ErrBackupOptionsNotAvailable menandakan backup options tidak tersedia
	ErrBackupOptionsNotAvailable = errors.New("backup options not available")

	// ErrInvalidStateType menandakan invalid state type
	ErrInvalidStateType = errors.New("invalid state type")

	// ErrBackgroundModeRequiresQuiet menandakan mode background membutuhkan --quiet
	ErrBackgroundModeRequiresQuiet = errors.New("background mode requires --quiet (non-interactive)")
)
