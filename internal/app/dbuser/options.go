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
	IncludeGrants      bool
	SplitOut           bool
}

type ApplyOptions struct {
	Profile       domain.ProfileInfo
	Files         []string
	Force         bool
	SkipUserCheck bool
}
