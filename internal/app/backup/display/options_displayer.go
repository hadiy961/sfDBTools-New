package display

import (
	"sfDBTools/internal/app/backup/model/types_backup"
	"sfDBTools/internal/ui/print"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/internal/ui/table"
	"sfDBTools/pkg/validation"
)

// OptionsDisplayer handles display of backup options.
type OptionsDisplayer struct {
	options *types_backup.BackupDBOptions
}

// NewOptionsDisplayer creates new options displayer.
func NewOptionsDisplayer(options *types_backup.BackupDBOptions) *OptionsDisplayer {
	return &OptionsDisplayer{options: options}
}

func (d *OptionsDisplayer) renderTable() {
	print.PrintSubHeader("Opsi Backup")

	data := [][]string{}
	data = append(data, d.buildGeneralSection()...)
	data = append(data, d.buildModeSpecificSection()...)
	data = append(data, d.buildProfileSection()...)
	data = append(data, d.buildFilterSection()...)
	data = append(data, d.buildCompressionSection()...)
	data = append(data, d.buildEncryptionSection()...)

	table.Render([]string{"Parameter", "Value"}, data)
}

// Render menampilkan backup options tanpa meminta konfirmasi.
func (d *OptionsDisplayer) Render() {
	d.renderTable()
}

// Display menampilkan backup options dan meminta konfirmasi.
func (d *OptionsDisplayer) Display() (bool, error) {
	d.renderTable()

	confirm, err := prompt.Confirm("Apakah Anda ingin melanjutkan?", true)
	if err != nil {
		return false, err
	}
	if !confirm {
		return false, validation.ErrUserCancelled
	}
	return true, nil
}
