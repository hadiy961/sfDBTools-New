// File : internal/backup/setup/edit_menu_items_common.go
// Deskripsi : Kumpulan helper untuk menyusun item menu edit yang berulang
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-31
// Last Modified : 2025-12-31

package setup

import (
	"context"

	"sfdbtools/pkg/database"
)

func (s *Setup) editItemsBackupProfileTicket(ctx context.Context, clientPtr **database.Client) []editMenuItem {
	return []editMenuItem{
		{Label: "Profile", Action: func() error { return s.changeBackupProfileAndReconnect(ctx, clientPtr) }},
		{Label: "Ticket number", Action: func() error { return s.changeBackupTicketInteractive() }},
	}
}

type backupToggleMenuOptions struct {
	CaptureGTID      bool
	ExportUserGrants bool
	ExcludeSystem    bool
	ExcludeEmpty     bool
	ExcludeData      bool
}

func (s *Setup) editItemsBackupToggles(opts backupToggleMenuOptions) []editMenuItem {
	items := make([]editMenuItem, 0, 5)
	if opts.CaptureGTID {
		items = append(items, editMenuItem{Label: "Capture GTID", Action: func() error { return s.changeBackupCaptureGTIDInteractive() }})
	}
	if opts.ExportUserGrants {
		items = append(items, editMenuItem{Label: "Export user grants", Action: func() error { return s.changeBackupExportUserGrantsInteractive() }})
	}
	if opts.ExcludeSystem {
		items = append(items, editMenuItem{Label: "Exclude system databases", Action: func() error { return s.changeBackupExcludeSystemInteractive() }})
	}
	if opts.ExcludeEmpty {
		items = append(items, editMenuItem{Label: "Exclude empty databases", Action: func() error { return s.changeBackupExcludeEmptyInteractive() }})
	}
	if opts.ExcludeData {
		items = append(items, editMenuItem{Label: "Exclude data (schema only)", Action: func() error { return s.changeBackupExcludeDataInteractive() }})
	}
	return items
}

func (s *Setup) editItemsBackupOutputSecurity(customOutputDir *string, includeFilename bool) []editMenuItem {
	items := []editMenuItem{
		{Label: "Backup directory", Action: func() error { return s.changeBackupOutputDirInteractive(customOutputDir) }},
	}
	if includeFilename {
		items = append(items, editMenuItem{Label: "Filename", Action: func() error { return s.changeBackupFilenameInteractive() }})
	}
	items = append(items,
		editMenuItem{Label: "Encryption", Action: func() error { return s.changeBackupEncryptionInteractive() }},
		editMenuItem{Label: "Compression", Action: func() error { return s.changeBackupCompressionInteractive() }},
	)
	return items
}
