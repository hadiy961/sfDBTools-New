// File : internal/shared/naming/database.go
// Deskripsi : Helper penamaan database (primary/secondary/companion) untuk menghindari duplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package naming

import (
	"path/filepath"
	"strings"

	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/shared/consts"
)

// BuildCompanionDBName membentuk nama database companion (_dmart).
//
// Example:
//
//	BuildCompanionDBName("dbsf_nbc_client") // "dbsf_nbc_client_dmart"
//	BuildCompanionDBName("dbsf_nbc_client_nodata") // "dbsf_nbc_client_dmart_nodata"
func BuildCompanionDBName(dbName string) string {
	companionDBName := dbName + consts.SuffixDmart
	lowerDBName := strings.ToLower(strings.TrimSpace(dbName))
	if strings.HasSuffix(lowerDBName, "_nodata") {
		base := dbName[:len(dbName)-len("_nodata")]
		companionDBName = base + consts.SuffixDmart + "_nodata"
	}
	return companionDBName
}

// InferPrimaryPrefix menentukan prefix primary database (biznet_/nbc_).
//
// Example:
//
//	InferPrimaryPrefix("dbsf_nbc_client", "") // "dbsf_nbc_"
//	InferPrimaryPrefix("", "/path/dbsf_biznet_client.sql") // "dbsf_biznet_"
func InferPrimaryPrefix(targetDB, filePath string) string {
	targetLower := strings.ToLower(strings.TrimSpace(targetDB))
	if strings.HasPrefix(targetLower, consts.PrimaryPrefixBiznet) {
		return consts.PrimaryPrefixBiznet
	}
	if strings.HasPrefix(targetLower, consts.PrimaryPrefixNBC) {
		return consts.PrimaryPrefixNBC
	}

	// Infer dari filename
	if strings.TrimSpace(filePath) != "" {
		inferred := backupfile.ExtractDatabaseNameFromFile(filepath.Base(filePath))
		inferredLower := strings.ToLower(strings.TrimSpace(inferred))
		if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
			return consts.PrimaryPrefixBiznet
		}
		if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
			return consts.PrimaryPrefixNBC
		}
	}

	return consts.PrimaryPrefixNBC
}

// ExtractClientCode mengekstrak client code dari nama database.
//
// Example:
//
//	ExtractClientCode("dbsf_nbc_client123", consts.PrimaryPrefixNBC) // "client123"
//	ExtractClientCode("dbsf_biznet_test", "") // "test"
func ExtractClientCode(dbName, prefix string) string {
	defaultClientCode := strings.ToLower(strings.TrimSpace(dbName))

	if p := strings.ToLower(strings.TrimSpace(prefix)); p != "" {
		if strings.HasPrefix(defaultClientCode, p) {
			return strings.TrimPrefix(defaultClientCode, p)
		}
	}

	if strings.HasPrefix(defaultClientCode, consts.PrimaryPrefixNBC) {
		return strings.TrimPrefix(defaultClientCode, consts.PrimaryPrefixNBC)
	}
	if strings.HasPrefix(defaultClientCode, consts.PrimaryPrefixBiznet) {
		return strings.TrimPrefix(defaultClientCode, consts.PrimaryPrefixBiznet)
	}

	return defaultClientCode
}

// BuildPrimaryDBName membentuk nama database primary dari prefix dan client code.
//
// Example:
//
//	BuildPrimaryDBName(consts.PrimaryPrefixNBC, "client123") // "dbsf_nbc_client123"
func BuildPrimaryDBName(prefix, clientCode string) string {
	cc := strings.TrimSpace(clientCode)
	ccLower := strings.ToLower(cc)
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		return cc
	}
	return prefix + cc
}

// BuildSecondaryDBName membentuk nama database secondary.
//
// Example:
//
//	BuildSecondaryDBName("dbsf_nbc_client", "001") // "dbsf_nbc_client_secondary_001"
func BuildSecondaryDBName(primaryDB, instance string) string {
	inst := strings.TrimSpace(instance)
	return primaryDB + "_secondary_" + inst
}
