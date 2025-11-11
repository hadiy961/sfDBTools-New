package backup

import (
	"sfDBTools/pkg/backuphelper"
)

// buildMysqldumpArgs membangun argumen mysqldump dengan kredensial database
// Parameter singleDB: jika tidak kosong, akan backup database tunggal tersebut
// Parameter dbFiltered: list database untuk backup multiple (diabaikan jika singleDB diisi)
// Parameter totalDBFound: total database yang ditemukan di server
func (s *Service) buildMysqldumpArgs(baseDumpArgs string, dbFiltered []string, singleDB string, totalDBFound int) []string {
	conn := backuphelper.DatabaseConn{
		Host:     s.BackupDBOptions.Profile.DBInfo.Host,
		Port:     s.BackupDBOptions.Profile.DBInfo.Port,
		User:     s.BackupDBOptions.Profile.DBInfo.User,
		Password: s.BackupDBOptions.Profile.DBInfo.Password,
	}

	filter := backuphelper.FilterOptions{
		ExcludeData:      s.BackupDBOptions.Filter.ExcludeData,
		ExcludeDatabases: s.BackupDBOptions.Filter.ExcludeDatabases,
		IncludeDatabases: s.BackupDBOptions.Filter.IncludeDatabases,
		ExcludeSystem:    s.BackupDBOptions.Filter.ExcludeSystem,
		ExcludeDBFile:    s.BackupDBOptions.Filter.ExcludeDBFile,
		IncludeFile:      s.BackupDBOptions.Filter.IncludeFile,
	}

	return backuphelper.BuildMysqldumpArgs(baseDumpArgs, conn, filter, dbFiltered, singleDB, totalDBFound)
}
