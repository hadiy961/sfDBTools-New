package types

// CleanupOptions menyimpan opsi cleanup untuk backup.
type CleanupOptions struct {
	Enabled         bool
	Days            int
	CleanupSchedule string
	Pattern         string
	Background      bool
	DryRun          bool
}

// CleanupEntryConfig menyimpan konfigurasi untuk entry point cleanup.
type CleanupEntryConfig struct {
	HeaderTitle string
	Mode        string
	ShowOptions bool
	SuccessMsg  string
	LogPrefix   string
}
