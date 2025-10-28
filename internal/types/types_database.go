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
