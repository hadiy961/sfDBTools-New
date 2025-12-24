// File : internal/appconfig/appconfig_types.go
// Deskripsi : Struct untuk konfigurasi aplikasi yang di-load dari file YAML
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package appconfig

import "sfDBTools/internal/types"

// Semua struct config dipusatkan di package internal/types.
// Type alias dipertahankan untuk menjaga API package appconfig tetap stabil.
type (
	Config             = types.Config
	BackupConfig       = types.BackupConfig
	IncludeConfig      = types.IncludeConfig
	CompressionConfig  = types.CompressionConfig
	ExcludeConfig      = types.ExcludeConfig
	CleanupConfig      = types.CleanupConfig
	EncryptionConfig   = types.EncryptionConfig
	OutputConfig       = types.OutputConfig
	VerificationConfig = types.VerificationConfig
	ReplicationConfig  = types.ReplicationConfig
	ConfigDirConfig    = types.ConfigDirConfig
	GeneralConfig      = types.GeneralConfig
	LogConfig          = types.LogConfig
	MariadbConfig      = types.MariadbConfig
	ScriptConfig       = types.ScriptConfig
	SystemUsersConfig  = types.SystemUsersConfig
)
