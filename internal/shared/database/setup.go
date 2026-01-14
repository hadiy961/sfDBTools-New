// File : internal/shared/database/setup.go
// Deskripsi : Type definitions untuk profile resolution & connection setup (dependency injection)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package database

import (
	"context"

	"sfdbtools/internal/domain"
)

// ProfileResolverOptions konfigurasi untuk resolve profile.
//
// Catatan: Ini adalah struct konfigurasi untuk dependency injection.
// Implementasi actual ada di internal/app/profile/helpers.ResolveAndLoadProfile.
//
// Perilaku yang diharapkan:
// - Jika ProfilePath kosong dan AllowInteractive=true, prompt user untuk pilih file.
// - Jika ProfilePath kosong dan AllowInteractive=false serta RequireProfile=true, return error.
// - Encryption key di-resolve dari EnvProfileKey jika tidak ada di EncryptionKey.
//
// Example (di app layer):
//
//	import profilehelper "sfdbtools/internal/app/profile/helpers"
//
//	profile, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
//	    ConfigDir:         opts.ConfigDir,
//	    ProfilePath:       opts.ProfilePath,
//	    ProfileKey:        opts.EncryptionKey,
//	    AllowInteractive:  opts.AllowInteractive,
//	    RequireProfile:    true,
//	    ProfilePurpose:    "source",
//	    InteractivePrompt: "Pilih profil database sumber:",
//	})
type ProfileResolverOptions struct {
	ProfilePath       string // Path ke file profile (optional)
	EncryptionKey     string // Encryption key (optional)
	AllowInteractive  bool   // Allow user prompt
	ConfigDir         string // Config directory untuk DB profiles
	EnvProfilePath    string // Environment variable untuk profile path (optional)
	EnvProfileKey     string // Environment variable untuk encryption key (optional)
	RequireProfile    bool   // Fail jika profile tidak ditemukan/tidak dipilih
	ProfilePurpose    string // Deskripsi tujuan (contoh: "source", "target")
	InteractivePrompt string // Custom prompt message
}

// ConnectionSetupOptions konfigurasi untuk setup database connection.
//
// Catatan: Ini adalah struct konfigurasi untuk dependency injection.
// Implementasi actual ada di internal/app/profile/helpers.ConnectWithProfile.
//
// Example (di app layer):
//
//	import profilehelper "sfdbtools/internal/app/profile/helpers"
//
//	client, err := profilehelper.ConnectWithProfile(opts.Profile, opts.InitialDB)
//	if err != nil { /* handle */ }
//	defer client.Close()
//
//	// Optional: fetch server hostname
//	if opts.Context != nil {
//	    hostname, _ := client.GetServerHostname(opts.Context)
//	    if hostname != "" {
//	        opts.Profile.DBInfo.HostName = hostname
//	    }
//	}
type ConnectionSetupOptions struct {
	Profile   *domain.ProfileInfo
	Context   context.Context
	InitialDB string // Initial database untuk connection (default: "mysql")
	Logger    interface {
		Infof(string, ...interface{})
		Warnf(string, ...interface{})
		Info(string)
	}
}
