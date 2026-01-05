package domain

import "time"

// SSHTunnelConfig menyimpan konfigurasi SSH tunnel untuk akses database melalui bastion.
// Catatan: ResolvedLocalPort hanya runtime (tidak perlu disimpan ke file).
type SSHTunnelConfig struct {
	Enabled           bool
	Host              string
	Port              int
	User              string
	Password          string
	IdentityFile      string
	LocalPort         int
	ResolvedLocalPort int
}

// ProfileInfo menyimpan informasi profile koneksi database.
type ProfileInfo struct {
	Name             string
	DBInfo           DBInfo
	SSHTunnel        SSHTunnelConfig
	EncryptionKey    string
	EncryptionSource string
	Size             string
	LastModified     time.Time
	Path             string
}
