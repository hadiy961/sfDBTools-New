// File : internal/backup/display/options_display.go
// Deskripsi : Display logic untuk backup options dan filter stats dengan modular builders
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package display

import (
	"fmt"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
	"sort"
)

// OptionsDisplayer handles display of backup options
type OptionsDisplayer struct {
	options *types_backup.BackupDBOptions
}

// NewOptionsDisplayer creates new options displayer
func NewOptionsDisplayer(options *types_backup.BackupDBOptions) *OptionsDisplayer {
	return &OptionsDisplayer{options: options}
}

// Display menampilkan backup options dan meminta konfirmasi
func (d *OptionsDisplayer) Display() (bool, error) {
	ui.PrintSubHeader("Opsi Backup")

	data := [][]string{}

	// Build sections
	data = append(data, d.buildGeneralSection()...)
	data = append(data, d.buildModeSpecificSection()...)
	data = append(data, d.buildProfileSection()...)
	data = append(data, d.buildFilterSection()...)
	data = append(data, d.buildCompressionSection()...)
	data = append(data, d.buildEncryptionSection()...)
	data = append(data, d.buildCleanupSection()...)

	ui.FormatTable([]string{"Parameter", "Value"}, data)

	// Konfirmasi
	confirm, err := input.AskYesNo("Apakah Anda ingin melanjutkan?", true)
	if err != nil {
		return false, err
	}
	if !confirm {
		return false, validation.ErrUserCancelled
	}
	return true, nil
}

// buildGeneralSection builds general information section
func (d *OptionsDisplayer) buildGeneralSection() [][]string {
	data := [][]string{
		{"Mode Backup", ui.ColorText(d.options.Mode, consts.UIColorCyan)},
		{"Output Directory", d.options.OutputDir},
	}

	// Filename display logic:
	// - Mode single/primary/secondary/combined/all: tampilkan filename akurat
	// - Mode separated dengan filter: tampilkan contoh filename
	if d.isSeparatedMode() {
		// Mode separated - tampilkan contoh
		data = append(data, []string{"Filename Example", ui.ColorText(d.options.File.Path, consts.UIColorCyan)})
	} else if d.isSingleMode() || d.options.Mode == consts.ModeCombined || d.options.Mode == consts.ModeAll {
		// Mode single/primary/secondary/combined/all - tampilkan filename akurat
		label := d.getFilenameLabel()
		data = append(data, []string{label, ui.ColorText(d.options.File.Path, consts.UIColorCyan)})
	}

	data = append(data, []string{"Dry Run", fmt.Sprintf("%v", d.options.DryRun)})

	return data
}

// buildModeSpecificSection builds mode-specific section (single/primary/secondary)
func (d *OptionsDisplayer) buildModeSpecificSection() [][]string {
	if !d.isSingleMode() {
		// Combined/separated/all mode specific
		data := [][]string{}

		// Tampilkan metode filter untuk combined vs all
		if d.options.Mode == consts.ModeAll {
			data = append(data, []string{"Metode Filter", ui.ColorText("Exclude (semua kecuali yang dikecualikan)", consts.UIColorYellow)})
		} else if d.options.Mode == consts.ModeCombined {
			data = append(data, []string{"Metode Filter", ui.ColorText("Include (hanya yang dipilih)", consts.UIColorYellow)})
		}

		if d.options.Mode == consts.ModeCombined || d.options.Mode == consts.ModeAll {
			data = append(data, []string{"Capture GTID", fmt.Sprintf("%v", d.options.CaptureGTID)})
		}

		data = append(data, []string{"Export User Grants", d.getExportUserStatus()})
		return data
	}

	// Single/primary/secondary mode
	data := [][]string{}

	dbName := d.options.DBName
	if dbName == "" {
		dbName = "<belum dipilih>"
	}
	data = append(data, []string{"Database Utama", ui.ColorText(dbName, consts.UIColorYellow)})

	if d.options.File.Filename != "" {
		data = append(data, []string{"Custom Filename", d.options.File.Filename})
	}

	// Companion options hanya untuk primary/secondary, tidak untuk single
	if d.options.Mode == consts.ModePrimary || d.options.Mode == consts.ModeSecondary {
		data = append(data, []string{"Include DMart", fmt.Sprintf("%v", d.options.IncludeDmart)})
		data = append(data, []string{"Include Temp", fmt.Sprintf("%v", d.options.IncludeTemp)})
		data = append(data, []string{"Include Archive", fmt.Sprintf("%v", d.options.IncludeArchive)})

		// Companion status
		if len(d.options.CompanionStatus) > 0 {
			data = append(data, d.buildCompanionStatus()...)
		}
	}

	data = append(data, []string{"Export User Grants", d.getExportUserStatus()})

	return data
}

// buildCompanionStatus builds companion database status rows
func (d *OptionsDisplayer) buildCompanionStatus() [][]string {
	data := [][]string{{"Companion DB", ""}}

	keys := make([]string, 0, len(d.options.CompanionStatus))
	for name := range d.options.CompanionStatus {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		status := "ditemukan"
		color := consts.UIColorGreen
		if !d.options.CompanionStatus[name] {
			status = "tidak ditemukan"
			color = consts.UIColorRed
		}
		data = append(data, []string{"  - " + name, ui.ColorText(status, color)})
	}

	return data
}

// buildProfileSection builds profile information section
func (d *OptionsDisplayer) buildProfileSection() [][]string {
	if d.options.Profile.Name == "" {
		return [][]string{}
	}

	return [][]string{
		{"Profile", ui.ColorText(d.options.Profile.Name, consts.UIColorYellow)},
		{"HostName", d.options.Profile.DBInfo.HostName},
		{"Host", fmt.Sprintf("%s:%d", d.options.Profile.DBInfo.Host, d.options.Profile.DBInfo.Port)},
		{"User", d.options.Profile.DBInfo.User},
	}
}

// buildFilterSection builds filter options section
func (d *OptionsDisplayer) buildFilterSection() [][]string {
	data := [][]string{
		{"", ""}, // Separator
		{ui.ColorText("Filter Options", consts.UIColorPurple), ""},
	}

	// Exclude system/empty untuk non-single modes
	if !d.isSingleMode() {
		data = append(data, []string{"Exclude System DB", fmt.Sprintf("%v", d.options.Filter.ExcludeSystem)})
		data = append(data, []string{"Exclude Empty DB", fmt.Sprintf("%v", d.options.Filter.ExcludeEmpty)})
	}

	// Exclude data
	dataLabel := "Exclude Data DB"
	if d.isSingleMode() {
		dataLabel = "Exclude Data"
	}
	data = append(data, []string{dataLabel, fmt.Sprintf("%v", d.options.Filter.ExcludeData)})

	// Include/Exclude lists
	data = append(data, d.buildDatabaseList("Include List", d.options.Filter.IncludeDatabases)...)
	data = append(data, d.buildDatabaseList("Exclude List", d.options.Filter.ExcludeDatabases)...)

	// Include/Exclude files
	if d.options.Filter.IncludeFile != "" {
		data = append(data, []string{"Include File", d.options.Filter.IncludeFile})
	}
	if d.options.Filter.ExcludeDBFile != "" {
		data = append(data, []string{"Exclude File", d.options.Filter.ExcludeDBFile})
	}

	return data
}

// buildDatabaseList builds database list rows (include/exclude)
func (d *OptionsDisplayer) buildDatabaseList(label string, databases []string) [][]string {
	if len(databases) == 0 {
		return [][]string{}
	}

	data := [][]string{
		{label, fmt.Sprintf("%d database", len(databases))},
	}

	// Show details if less than 5
	if len(databases) < 5 {
		for _, db := range databases {
			data = append(data, []string{"  - " + db, ""})
		}
	}

	return data
}

// buildCompressionSection builds compression options section
func (d *OptionsDisplayer) buildCompressionSection() [][]string {
	data := [][]string{
		{"", ""}, // Separator
		{ui.ColorText("Compression", consts.UIColorPurple), ""},
		{"Enabled", fmt.Sprintf("%v", d.options.Compression.Enabled)},
	}

	if d.options.Compression.Enabled {
		data = append(data, []string{"Type", d.options.Compression.Type})
		data = append(data, []string{"Level", fmt.Sprintf("%d", d.options.Compression.Level)})
	}

	return data
}

// buildEncryptionSection builds encryption options section
func (d *OptionsDisplayer) buildEncryptionSection() [][]string {
	data := [][]string{
		{"", ""}, // Separator
		{ui.ColorText("Encryption", consts.UIColorPurple), ""},
		{"Enabled", fmt.Sprintf("%v", d.options.Encryption.Enabled)},
	}

	if d.options.Encryption.Enabled {
		data = append(data, []string{"Key Status", ui.ColorText("Configured", consts.UIColorGreen)})
	}

	return data
}

// buildCleanupSection builds cleanup options section
func (d *OptionsDisplayer) buildCleanupSection() [][]string {
	data := [][]string{
		{"", ""}, // Separator
		{ui.ColorText("Cleanup", consts.UIColorPurple), ""},
		{"Enabled", fmt.Sprintf("%v", d.options.Cleanup.Enabled)},
	}

	if d.options.Cleanup.Enabled {
		data = append(data, []string{"Days to Keep", fmt.Sprintf("%d", d.options.Cleanup.Days)})
		if d.options.Cleanup.Pattern != "" {
			data = append(data, []string{"Pattern", d.options.Cleanup.Pattern})
		}
		if d.options.Cleanup.CleanupSchedule != "" {
			data = append(data, []string{"Schedule", d.options.Cleanup.CleanupSchedule})
		}
		data = append(data, []string{"Background", fmt.Sprintf("%v", d.options.Cleanup.Background)})
	}

	return data
}

// Helper methods
func (d *OptionsDisplayer) isSingleMode() bool {
	return d.options.Mode == consts.ModeSingle || d.options.Mode == consts.ModePrimary || d.options.Mode == consts.ModeSecondary
}

func (d *OptionsDisplayer) isSeparatedMode() bool {
	return d.options.Mode == consts.ModeSeparated || d.options.Mode == consts.ModeSeparate
}

func (d *OptionsDisplayer) getExportUserStatus() string {
	if d.options.ExcludeUser {
		return "No"
	}
	return ui.ColorText("Yes", consts.UIColorGreen)
}

// getFilenameLabel menentukan label filename berdasarkan mode
func (d *OptionsDisplayer) getFilenameLabel() string {
	switch d.options.Mode {
	case consts.ModeAll:
		return "Output File (All)"
	case consts.ModeCombined:
		return "Output File (Filter)"
	default:
		return "Backup Filename"
	}
}

// DisplayFilterStats menampilkan statistik hasil pemfilteran database
func DisplayFilterStats(stats *types.FilterStats, logger applog.Logger) {
	ui.DisplayFilterStats(stats, consts.FeatureBackup, logger)
}
