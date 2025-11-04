package backup

import (
	"fmt"
	"strings"
)

// buildMysqldumpArgs membangun argumen mysqldump dengan kredensial database
// Parameter singleDB: jika tidak kosong, akan backup database tunggal tersebut
// Parameter dbFiltered: list database untuk backup multiple (diabaikan jika singleDB diisi)
// Parameter totalDBFound: total database yang ditemukan di server
func (s *Service) buildMysqldumpArgs(baseDumpArgs string, dbFiltered []string, singleDB string, totalDBFound int) []string {
	var args []string

	// Tambahkan kredensial database
	dbConn := s.BackupDBOptions.Profile.DBInfo

	// Host
	if dbConn.Host != "" {
		args = append(args, "--host="+dbConn.Host)
	}

	// Port
	if dbConn.Port != 0 {
		args = append(args, fmt.Sprintf("--port=%d", dbConn.Port))
	}

	// User
	if dbConn.User != "" {
		args = append(args, "--user="+dbConn.User)
	}

	// Password
	if dbConn.Password != "" {
		args = append(args, "--password="+dbConn.Password)
	}

	// Tambahkan argumen mysqldump dari konfigurasi
	if baseDumpArgs != "" {
		baseArgs := strings.Fields(baseDumpArgs)
		args = append(args, baseArgs...)
	}

	// jika exclude-data diaktifkan, tambahkan --no-data
	if s.BackupDBOptions.Filter.ExcludeData {
		args = append(args, "--no-data")
	}

	// Mode single database
	if singleDB != "" {
		args = append(args, "--databases")
		args = append(args, singleDB)
		return args
	}

	// Mode multiple databases
	// Jika tidak ada filter (exclude/include), gunakan --all-databases
	hasFilter := len(s.BackupDBOptions.Filter.ExcludeDatabases) > 0 ||
		len(s.BackupDBOptions.Filter.IncludeDatabases) > 0 ||
		s.BackupDBOptions.Filter.ExcludeSystem ||
		s.BackupDBOptions.Filter.ExcludeDBFile != "" ||
		s.BackupDBOptions.Filter.IncludeFile != ""

	// Jika jumlah database yang akan di-backup sama dengan total dan tidak ada filter, gunakan --all-databases
	if !hasFilter && len(dbFiltered) == totalDBFound {
		args = append(args, "--all-databases")
	} else {
		args = append(args, "--databases")
		args = append(args, dbFiltered...)
	}

	return args
}
