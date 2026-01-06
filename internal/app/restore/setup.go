// File : internal/restore/setup.go
// Deskripsi : Shared helper functions untuk setup restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 30 Desember 2025

package restore

import (
	"path/filepath"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/helper"
	"strings"
)

// inferPrimaryPrefixFromTargetOrFile menentukan prefix primary database
// dari target DB atau dari nama file backup
func inferPrimaryPrefixFromTargetOrFile(targetDB string, filePath string) string {
	targetLower := strings.ToLower(strings.TrimSpace(targetDB))
	if strings.HasPrefix(targetLower, consts.PrimaryPrefixBiznet) {
		return consts.PrimaryPrefixBiznet
	}
	if strings.HasPrefix(targetLower, consts.PrimaryPrefixNBC) {
		return consts.PrimaryPrefixNBC
	}

	inferred := helper.ExtractDatabaseNameFromFile(filepath.Base(filePath))
	inferredLower := strings.ToLower(strings.TrimSpace(inferred))
	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
		return consts.PrimaryPrefixBiznet
	}
	if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
		return consts.PrimaryPrefixNBC
	}

	return consts.PrimaryPrefixNBC
}

// buildPrimaryTargetDBFromClientCode membentuk nama database primary
// dari prefix dan client code
func buildPrimaryTargetDBFromClientCode(prefix string, clientCode string) string {
	cc := strings.TrimSpace(clientCode)
	ccLower := strings.ToLower(cc)
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		return cc
	}
	return prefix + cc
}
