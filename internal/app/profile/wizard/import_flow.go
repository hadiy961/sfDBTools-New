// File : internal/app/profile/wizard/import_flow.go
// Deskripsi : Import wizard orchestration (main flow)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026

package wizard

import (

	"sfdbtools/internal/app/profile/helpers/importer"
	"sfdbtools/internal/app/profile/helpers/parser"
	profilemodel "sfdbtools/internal/app/profile/model"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// ImportWizard handles interactive import workflow
type ImportWizard struct {
	Log       applog.Logger
	Config    *appconfig.Config
	ConfigDir string
}

// NewImportWizard creates a new import wizard instance
func NewImportWizard(log applog.Logger, cfg *appconfig.Config, configDir string) *ImportWizard {
	return &ImportWizard{
		Log:       log,
		Config:    cfg,
		ConfigDir: configDir,
	}
}

// Run executes the complete import workflow
// Returns planned rows yang siap untuk disave oleh executor
func (w *ImportWizard) Run(opts *profilemodel.ProfileImportOptions) ([]profilemodel.ImportRow, error) {
	// PHASE 1: Read source (interactive source selection)
	headers, dataRows, warnings, srcLabel, err := w.readSource(opts)
	if err != nil {
		w.Log.Warnf("[profile-import] Gagal membaca sumber %s: %v", srcLabel, err)
		return nil, err
	}
	w.logSourceInfo(srcLabel, headers, dataRows, warnings)

	// PHASE 2: Schema validation + row parsing
	parsedRows, err := w.validateAndParse(headers, dataRows, srcLabel, opts)
	if err != nil {
		return nil, err
	}

	// PHASE 3: Conflict resolution + connection test
	planned, err := w.resolveAndPlan(parsedRows, opts)
	if err != nil {
		w.Log.Warnf("[profile-import] Pre-save checks gagal: %v", err)
		return nil, err
	}
	w.Log.Infof("[profile-import] Pre-save checks selesai | akan_simpan=%d | total_rows=%d",
		importer.CountPlanned(planned), len(planned))

	return planned, nil
}

// logSourceInfo logs informasi sumber data yang dibaca
func (w *ImportWizard) logSourceInfo(srcLabel string, headers []string, dataRows [][]string, warnings []string) {
	nonEmpty := importer.CountNonEmptyRows(dataRows)
	w.Log.Infof("[profile-import] Sukses membaca %s | header=%d | rows_total=%d | rows_non_empty=%d",
		srcLabel, len(headers), len(dataRows), nonEmpty)
	for _, warn := range warnings {
		w.Log.Warn(warn)
	}
}

// validateAndParse melakukan validasi schema dan parsing rows
func (w *ImportWizard) validateAndParse(headers []string, dataRows [][]string, srcLabel string, opts *profilemodel.ProfileImportOptions) ([]profilemodel.ImportRow, error) {
	// Schema validation
	w.Log.Infof("[profile-import] Validasi schema/kolom (1/3)")
	schemaResult, err := parser.BuildImportSchema(headers)
	if err != nil {
		w.Log.Warnf("[profile-import] Validasi schema gagal: %v", err)
		return nil, err
	}
	schema := schemaResult.Schema
	w.Log.Infof("[profile-import] Validasi schema OK | kolom_terdeteksi=%d", len(schema))
	for _, warn := range schemaResult.Warnings {
		w.Log.Warn(warn)
	}

	// Row parsing + validation
	nonEmpty := importer.CountNonEmptyRows(dataRows)
	w.Log.Infof("[profile-import] Validasi per-row (2/3) | rows_non_empty=%d", nonEmpty)
	parsedRows, rowErrs := parser.ParseImportRows(dataRows, schema)

	// Duplicate names validation
	w.Log.Infof("[profile-import] Memeriksa duplikasi nama profile")
	dupErrs := importer.ValidateDuplicateNames(parsedRows)

	// Handle invalid rows via wizard (interactive atau automation)
	parsedRows, err = w.handleInvalidRows(parsedRows, rowErrs, dupErrs, srcLabel, opts)
	if err != nil {
		return nil, err
	}

	return parsedRows, nil
}
