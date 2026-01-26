// File : internal/app/profile/validation/import_validation.go
// Deskripsi : Validation helpers untuk import profile (reuse existing validators)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package validation

import (
	"sfdbtools/internal/app/profile/helpers/common"
	profilemodel "sfdbtools/internal/app/profile/model"
	sharedvalidation "sfdbtools/internal/shared/validation"
)

// ValidateRequiredField validates field tidak kosong dan append error jika kosong
func ValidateRequiredField(value, fieldName string, rowNum int, errList *[]profilemodel.ImportCellError) {
	if common.IsEmpty(value) {
		*errList = append(*errList, profilemodel.ImportCellError{
			Row:     rowNum,
			Column:  fieldName,
			Message: fieldName + " kosong",
		})
	}
}

// ValidateProfileNameField validates profile name format (reuse existing)
func ValidateProfileNameField(name string, rowNum int, errList *[]profilemodel.ImportCellError) {
	if err := sharedvalidation.ValidateProfileName(name); err != nil {
		*errList = append(*errList, profilemodel.ImportCellError{
			Row:     rowNum,
			Column:  "name",
			Message: err.Error(),
		})
	}
}
