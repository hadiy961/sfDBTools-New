package dbuser

import "sfdbtools/internal/domain"

type ExportOptions struct {
	Profile domain.ProfileInfo

	OutPath string
	OutPerm string

	Users      []string // user@host
	Databases  []string
	DBFile     string
	ClientCode string

	ExcludeSystemUsers bool
	IncludeCreateUser  bool
}

type ApplyOptions struct {
	Profile domain.ProfileInfo
	File    string
	Force   bool
}
