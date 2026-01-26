// File : internal/app/profile/helpers/importer/row_processor.go
// Deskripsi : Row processing helpers untuk import workflow
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package importer

import (
	"fmt"
	"sfdbtools/internal/app/profile/helpers/common"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	"strings"
)

// MarkRowsAsSkipped marks multiple rows sebagai skipped dengan reason
func MarkRowsAsSkipped(rows []profilemodel.ImportRow, skipRowNums map[int]bool, reason string) []profilemodel.ImportRow {
	result := make([]profilemodel.ImportRow, 0, len(rows))
	for _, r := range rows {
		if skipRowNums[r.RowNum] {
			r.Skip = true
			if strings.TrimSpace(r.SkipReason) == "" {
				r.SkipReason = reason
			}
		}
		result = append(result, r)
	}
	return result
}

// BuildProfileInfoFromRow converts ImportRow ke domain.ProfileInfo
func BuildProfileInfoFromRow(r profilemodel.ImportRow) domain.ProfileInfo {
	name := common.Trim(r.PlannedName)
	if common.IsEmpty(name) {
		name = common.NormalizeName(r.Name)
	}

	info := domain.ProfileInfo{
		Name: name,
		DBInfo: domain.DBInfo{
			Host:     strings.TrimSpace(r.Host),
			Port:     r.Port,
			User:     strings.TrimSpace(r.User),
			Password: r.Password,
		},
		SSHTunnel: domain.SSHTunnelConfig{
			Enabled:      r.SSHEnabled,
			Host:         strings.TrimSpace(r.SSHHost),
			Port:         r.SSHPort,
			User:         strings.TrimSpace(r.SSHUser),
			Password:     r.SSHPassword,
			IdentityFile: strings.TrimSpace(r.SSHIdentity),
			LocalPort:    r.SSHLocal,
		},
		EncryptionKey:    r.ProfileKey,
		EncryptionSource: "import",
	}

	return info
}

// CountPlanned counts rows yang tidak di-skip
func CountPlanned(rows []profilemodel.ImportRow) int {
	n := 0
	for _, r := range rows {
		if !r.Skip {
			n++
		}
	}
	return n
}

// CountNonEmptyRows counts rows yang tidak kosong
func CountNonEmptyRows(rows [][]string) int {
	count := 0
	for _, r := range rows {
		isEmpty := true
		for _, c := range r {
			if strings.TrimSpace(c) != "" {
				isEmpty = false
				break
			}
		}
		if !isEmpty {
			count++
		}
	}
	return count
}

// ValidateDuplicateNames memeriksa duplikasi nama profile dalam rows
func ValidateDuplicateNames(rows []profilemodel.ImportRow) []profilemodel.ImportCellError {
	seen := map[string]int{}
	errList := []profilemodel.ImportCellError{}
	for _, r := range rows {
		if r.Skip {
			continue
		}
		name := common.NormalizeName(r.Name)
		if common.IsEmpty(name) {
			continue
		}
		key := common.TrimLower(name)
		if firstRow, ok := seen[key]; ok {
			errList = append(errList, profilemodel.ImportCellError{
				Row:     r.RowNum,
				Column:  "name",
				Message: fmt.Sprintf("duplikasi nama dengan row %d", firstRow),
			})
		} else {
			seen[key] = r.RowNum
		}
	}
	return errList
}

// SafeName returns nama profile yang aman untuk ditampilkan
func SafeName(name string) string {
	return common.SafeString(name, "(unknown)")
}
