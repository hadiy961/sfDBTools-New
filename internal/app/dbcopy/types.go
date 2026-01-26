// File : internal/app/dbcopy/types.go
// Deskripsi : Types/options untuk db-copy
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopy

type CommonOptions struct {
	SourceProfile    string
	SourceProfileKey string
	TargetProfile    string
	TargetProfileKey string

	Ticket string

	SkipConfirm     bool
	ContinueOnError bool
	DryRun          bool
	ExcludeData     bool

	IncludeDmart    bool
	PrebackupTarget bool
	Workdir         string
}

type P2SOptions struct {
	Common CommonOptions

	ClientCode string
	Instance   string

	SourceDB string
	TargetDB string
}

type P2POptions struct {
	Common CommonOptions

	ClientCode       string
	TargetClientCode string

	SourceDB string
	TargetDB string
}

type S2SOptions struct {
	Common CommonOptions

	ClientCode     string
	SourceInstance string
	TargetInstance string
	SourceDB       string
	TargetDB       string
}
