// File : internal/restore/setup_confirmation.go
// Deskripsi : Confirmation display utilities untuk setup restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"fmt"
	"path/filepath"
	"sfDBTools/internal/app/restore/display"
	"strings"
)

// ConfirmationBuilder membangun data konfirmasi untuk ditampilkan
type ConfirmationBuilder struct {
	options map[string]string
}

// NewConfirmationBuilder membuat instance builder baru
func NewConfirmationBuilder() *ConfirmationBuilder {
	return &ConfirmationBuilder{
		options: make(map[string]string),
	}
}

// Add menambahkan pasangan key-value
func (b *ConfirmationBuilder) Add(key, value string) *ConfirmationBuilder {
	b.options[key] = value
	return b
}

// AddBool menambahkan boolean value
func (b *ConfirmationBuilder) AddBool(key string, value bool) *ConfirmationBuilder {
	b.options[key] = fmt.Sprintf("%v", value)
	return b
}

// AddFile menambahkan file path (hanya basename)
func (b *ConfirmationBuilder) AddFile(key, filePath string) *ConfirmationBuilder {
	if filePath != "" {
		b.options[key] = filepath.Base(filePath)
	}
	return b
}

// AddConditional menambahkan value berdasarkan kondisi
func (b *ConfirmationBuilder) AddConditional(key string, condition bool, trueVal, falseVal string) *ConfirmationBuilder {
	if condition {
		b.options[key] = trueVal
	} else {
		b.options[key] = falseVal
	}
	return b
}

// AddHostPort menambahkan format host:port
func (b *ConfirmationBuilder) AddHostPort(key, host string, port int) *ConfirmationBuilder {
	b.options[key] = fmt.Sprintf("%s:%d", host, port)
	return b
}

// AddCompanion menambahkan status companion (dmart)
func (b *ConfirmationBuilder) AddCompanion(includeDmart bool, companionFile string) *ConfirmationBuilder {
	if includeDmart {
		status := "Auto-detect"
		if strings.TrimSpace(companionFile) != "" {
			status = filepath.Base(companionFile)
		}
		b.options["Companion (dmart)"] = status
	}
	return b
}

// AddGrants menambahkan status grants file
func (b *ConfirmationBuilder) AddGrants(skipGrants bool, grantsFile string) *ConfirmationBuilder {
	if skipGrants {
		b.options["Grants File"] = "Skipped"
	} else if grantsFile != "" {
		b.options["Grants File"] = filepath.Base(grantsFile)
	} else {
		b.options["Grants File"] = "Tidak ada"
	}
	return b
}

// AddBackupDir menambahkan direktori backup jika applicable
func (b *ConfirmationBuilder) AddBackupDir(skipBackup bool, backupDir string) *ConfirmationBuilder {
	if !skipBackup && backupDir != "" {
		b.options["Backup Directory"] = backupDir
	}
	return b
}

// Display menampilkan konfirmasi ke user
func (b *ConfirmationBuilder) Display() error {
	return display.DisplayConfirmation(b.options)
}

// Build mengembalikan map hasil building
func (b *ConfirmationBuilder) Build() map[string]string {
	return b.options
}
