package display

import (
	"fmt"
	"sfDBTools/internal/ui/text"
	"sfDBTools/pkg/consts"
	"sort"
)

func (d *OptionsDisplayer) buildGeneralSection() [][]string {
	data := [][]string{
		{"Mode Backup", text.ColorText(d.options.Mode, consts.UIColorCyan)},
		{"Ticket", text.ColorText(d.options.Ticket, consts.UIColorYellow)},
		{"Output Directory", d.options.OutputDir},
	}

	// Filename display logic:
	// - Mode single/primary/secondary/combined/all: tampilkan filename akurat
	// - Mode separated dengan filter: tampilkan contoh filename
	if d.isSeparatedMode() {
		data = append(data, []string{"Filename Example", text.ColorText(d.options.File.Path, consts.UIColorCyan)})
	} else if d.isSingleMode() || d.options.Mode == consts.ModeCombined || d.options.Mode == consts.ModeAll {
		label := d.getFilenameLabel()
		data = append(data, []string{label, text.ColorText(d.options.File.Path, consts.UIColorCyan)})
	}

	data = append(data, []string{"Dry Run", fmt.Sprintf("%v", d.options.DryRun)})
	return data
}

func (d *OptionsDisplayer) buildModeSpecificSection() [][]string {
	if !d.isSingleMode() {
		data := [][]string{}

		if d.options.Mode == consts.ModeAll {
			data = append(data, []string{"Metode Filter", text.ColorText("Exclude (semua kecuali yang dikecualikan)", consts.UIColorYellow)})
		} else if d.options.Mode == consts.ModeCombined {
			data = append(data, []string{"Metode Filter", text.ColorText("Include (hanya yang dipilih)", consts.UIColorYellow)})
		}

		if d.options.Mode == consts.ModeCombined || d.options.Mode == consts.ModeAll {
			data = append(data, []string{"Capture GTID", fmt.Sprintf("%v", d.options.CaptureGTID)})
		}

		data = append(data, []string{"Export User Grants", d.getExportUserStatus()})
		return data
	}

	data := [][]string{}

	dbName := d.options.DBName
	if dbName == "" {
		dbName = "<belum dipilih>"
	}
	data = append(data, []string{"Database Utama", text.ColorText(dbName, consts.UIColorYellow)})

	if d.options.File.Filename != "" {
		data = append(data, []string{"Custom Filename", d.options.File.Filename})
	}

	if d.options.Mode == consts.ModePrimary || d.options.Mode == consts.ModeSecondary {
		data = append(data, []string{"Include DMart", fmt.Sprintf("%v", d.options.IncludeDmart)})

		if len(d.options.CompanionStatus) > 0 {
			data = append(data, d.buildCompanionStatus()...)
		}
	}

	data = append(data, []string{"Export User Grants", d.getExportUserStatus()})
	return data
}

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
		data = append(data, []string{"  - " + name, text.ColorText(status, color)})
	}

	return data
}

func (d *OptionsDisplayer) buildProfileSection() [][]string {
	if d.options.Profile.Name == "" {
		return [][]string{}
	}

	return [][]string{
		{"Profile", text.ColorText(d.options.Profile.Name, consts.UIColorYellow)},
		{"HostName", d.options.Profile.DBInfo.HostName},
		{"Host", fmt.Sprintf("%s:%d", d.options.Profile.DBInfo.Host, d.options.Profile.DBInfo.Port)},
		{"User", d.options.Profile.DBInfo.User},
	}
}

func (d *OptionsDisplayer) buildFilterSection() [][]string {
	data := [][]string{
		{"", ""},
		{text.ColorText("Filter Options", consts.UIColorPurple), ""},
	}

	if !d.isSingleMode() {
		data = append(data, []string{"Exclude System DB", fmt.Sprintf("%v", d.options.Filter.ExcludeSystem)})
		data = append(data, []string{"Exclude Empty DB", fmt.Sprintf("%v", d.options.Filter.ExcludeEmpty)})
	}

	dataLabel := "Exclude Data DB"
	if d.isSingleMode() {
		dataLabel = "Exclude Data"
	}
	data = append(data, []string{dataLabel, fmt.Sprintf("%v", d.options.Filter.ExcludeData)})

	data = append(data, d.buildDatabaseList("Include List", d.options.Filter.IncludeDatabases)...)
	data = append(data, d.buildDatabaseList("Exclude List", d.options.Filter.ExcludeDatabases)...)

	if d.options.Filter.IncludeFile != "" {
		data = append(data, []string{"Include File", d.options.Filter.IncludeFile})
	}
	if d.options.Filter.ExcludeDBFile != "" {
		data = append(data, []string{"Exclude File", d.options.Filter.ExcludeDBFile})
	}

	return data
}

func (d *OptionsDisplayer) buildDatabaseList(label string, databases []string) [][]string {
	if len(databases) == 0 {
		return [][]string{}
	}

	data := [][]string{{label, fmt.Sprintf("%d database", len(databases))}}

	if len(databases) < 5 {
		for _, db := range databases {
			data = append(data, []string{"  - " + db, ""})
		}
	}

	return data
}

func (d *OptionsDisplayer) buildCompressionSection() [][]string {
	data := [][]string{
		{"", ""},
		{text.ColorText("Compression", consts.UIColorPurple), ""},
		{"Enabled", fmt.Sprintf("%v", d.options.Compression.Enabled)},
	}

	if d.options.Compression.Enabled {
		data = append(data, []string{"Type", d.options.Compression.Type})
		data = append(data, []string{"Level", fmt.Sprintf("%d", d.options.Compression.Level)})
	}

	return data
}

func (d *OptionsDisplayer) buildEncryptionSection() [][]string {
	data := [][]string{
		{"", ""},
		{text.ColorText("Encryption", consts.UIColorPurple), ""},
		{"Enabled", fmt.Sprintf("%v", d.options.Encryption.Enabled)},
	}

	if d.options.Encryption.Enabled {
		statusText := text.ColorText("Missing", consts.UIColorRed)
		if d.options.Encryption.Key != "" {
			statusText = text.ColorText("Configured", consts.UIColorGreen)
		}
		data = append(data, []string{"Key Status", statusText})
	}

	return data
}
