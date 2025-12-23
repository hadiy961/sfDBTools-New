package execution

import (
	"strconv"
	"strings"
)

// DatabaseConn is a small representation of connection info required to build mysqldump args.
type DatabaseConn struct {
	Host     string
	Port     int
	User     string
	Password string
}

// FilterOptions mirrors the subset of backup filter options used when building mysqldump args.
type FilterOptions struct {
	ExcludeData      bool
	ExcludeDatabases []string
	IncludeDatabases []string
	ExcludeSystem    bool
	ExcludeDBFile    string
	IncludeFile      string
}

// BuildMysqldumpArgs membuat argumen mysqldump dari parameter umum.
func BuildMysqldumpArgs(baseDumpArgs string, conn DatabaseConn, filter FilterOptions, dbFiltered []string, singleDB string, totalDBFound int) []string {
	var args []string

	if conn.Host != "" {
		args = append(args, "--host="+conn.Host)
	}
	if conn.Port != 0 {
		args = append(args, "--port="+strconv.Itoa(conn.Port))
	}
	if conn.User != "" {
		args = append(args, "--user="+conn.User)
	}
	if conn.Password != "" {
		args = append(args, "--password="+conn.Password)
	}

	if baseDumpArgs != "" {
		args = append(args, strings.Fields(baseDumpArgs)...)
	}

	if filter.ExcludeData {
		args = append(args, "--no-data")
	}

	// KASUS 1: Single DB eksplisit
	if singleDB != "" {
		args = append(args, singleDB)
		return args
	}

	hasFilter := len(filter.ExcludeDatabases) > 0 ||
		len(filter.IncludeDatabases) > 0 ||
		filter.ExcludeSystem ||
		filter.ExcludeDBFile != "" ||
		filter.IncludeFile != ""

	// KASUS 2: Full backup
	if !hasFilter && len(dbFiltered) == totalDBFound {
		args = append(args, "--all-databases")
		return args
	}

	// KASUS 3: Filtered
	if len(dbFiltered) == 1 {
		args = append(args, dbFiltered[0])
		return args
	}
	if len(dbFiltered) > 1 {
		args = append(args, "--databases")
		args = append(args, dbFiltered...)
	}

	return args
}

// MaskPasswordInArgs returns a copy of args with password values masked for logging.
func MaskPasswordInArgs(args []string) []string {
	masked := make([]string, len(args))
	copy(masked, args)

	for i := 0; i < len(masked); i++ {
		arg := masked[i]
		if strings.HasPrefix(arg, "-p") && len(arg) > 2 {
			masked[i] = "-p********"
			continue
		}
		if strings.HasPrefix(arg, "--password=") {
			masked[i] = "--password=********"
			continue
		}
		if arg == "-p" || arg == "--password" {
			if i+1 < len(masked) {
				masked[i+1] = "********"
			}
		}
	}

	return masked
}
