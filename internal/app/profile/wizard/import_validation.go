// File : internal/app/profile/wizard/import_validation.go
// Deskripsi : Import validation + invalid rows handling wizard
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026

package wizard

import (
	importdisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/app/profile/helpers/importer"
	profilemodel "sfdbtools/internal/app/profile/model"
	sharedvalidation "sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

// handleInvalidRows handles invalid rows validation errors
// Prompts user untuk skip atau cancel (interactive mode)
// Automation mode: mengikuti --skip-invalid-rows atau return error
func (w *ImportWizard) handleInvalidRows(
	parsedRows []profilemodel.ImportRow,
	rowErrs []profilemodel.ImportCellError,
	dupErrs []profilemodel.ImportCellError,
	srcLabel string,
	opts *profilemodel.ProfileImportOptions,
) ([]profilemodel.ImportRow, error) {
	// Combine all validation errors
	allErrs := make([]profilemodel.ImportCellError, 0, len(rowErrs)+len(dupErrs))
	allErrs = append(allErrs, rowErrs...)
	allErrs = append(allErrs, dupErrs...)

	// No errors, return early
	if len(allErrs) == 0 {
		w.Log.Infof("[profile-import] Ringkasan validasi: ok=%d invalid=0", len(parsedRows))
		return parsedRows, nil
	}

	// Count invalid rows
	perRow := importer.GroupErrorsByRow(allErrs)
	invalidRowsCount := len(perRow)
	w.Log.Warnf("[profile-import] Ditemukan error validasi: %d (baris terdampak: %d)", len(allErrs), len(perRow))

	// Determine action: skip invalid atau cancel
	skipInvalid := opts.SkipInvalidRows
	if !skipInvalid {
		if opts.SkipConfirm {
			// Automation mode: error jika tidak ada --skip-invalid-rows
			return nil, importer.FormatImportErrors("validasi import gagal", srcLabel, allErrs)
		}

		// Full interaktif: tampilkan ringkasan + prompt global
		importdisplay.PrintImportInvalidRowsSummary(allErrs, srcLabel, 20)
		choice, _, askErr := prompt.SelectOne(
			"Ada baris invalid. Pilih aksi:",
			[]string{"Skip semua invalid rows dan lanjut", "Batalkan (perbaiki sheet dulu)"},
			0,
		)
		if askErr != nil {
			return nil, sharedvalidation.HandleInputError(askErr)
		}
		switch choice {
		case "Skip semua invalid rows dan lanjut":
			skipInvalid = true
			w.Log.Warnf("[profile-import] Mode run ini: skip invalid rows (tanpa perlu flag --skip-invalid-rows)")
		default:
			return nil, sharedvalidation.ErrUserCancelled
		}
	}

	// Mark invalid rows
	if skipInvalid {
		// Mark invalid data rows
		invalidRowNums := make(map[int]bool, len(rowErrs))
		for _, er := range rowErrs {
			invalidRowNums[er.Row] = true
		}
		parsedRows = importer.MarkRowsAsSkipped(parsedRows, invalidRowNums, profilemodel.ImportSkipReasonInvalid)

		// Mark duplicate name rows
		dupRowNums := make(map[int]bool, len(dupErrs))
		for _, er := range dupErrs {
			dupRowNums[er.Row] = true
		}
		parsedRows = importer.MarkRowsAsSkipped(parsedRows, dupRowNums, profilemodel.ImportSkipReasonDuplicate)
	}

	// Log per-row validation results (debug level untuk sheet besar)
	validCount := 0
	for i, r := range parsedRows {
		if r.Skip {
			w.Log.Debugf("[profile-import] %d/%d validasi row=%d name=%s -> invalid (skip)",
				i+1, len(parsedRows), r.RowNum, importer.SafeName(r.Name))
		} else {
			validCount++
			w.Log.Debugf("[profile-import] %d/%d validasi row=%d name=%s -> OK",
				i+1, len(parsedRows), r.RowNum, importer.SafeName(r.Name))
		}
	}
	w.Log.Infof("[profile-import] Ringkasan validasi: ok=%d invalid=%d", validCount, invalidRowsCount)

	return parsedRows, nil
}
