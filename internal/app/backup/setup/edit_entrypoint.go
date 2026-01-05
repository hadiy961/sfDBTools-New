package setup

import (
	"context"
	"fmt"

	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
)

// editBackupOptionsInteractive adalah satu-satunya entrypoint untuk aksi "Ubah opsi"
// dari session loop.
func (s *Setup) editBackupOptionsInteractive(ctx context.Context, client **database.Client, customOutputDir *string, mode string) error {
	items := s.editItemsBackupProfileTicket(ctx, client)

	includeSelectionItems := []editMenuItem{
		{Label: "Pilih database (multi-select)", Action: func() error { return s.changeBackupIncludeSelectionSelectDatabasesInteractive(ctx, client) }},
		{Label: "Reset ke mode interaktif", Action: func() error { return s.changeBackupIncludeSelectionResetInteractive() }},
		{Label: "Include list (manual)", Action: func() error { return s.changeBackupIncludeSelectionIncludeListManualInteractive() }},
		{Label: "Include file", Action: func() error { return s.changeBackupIncludeSelectionIncludeFileInteractive() }},
	}

	switch mode {
	case consts.ModeAll:
		items = append(items, s.editItemsBackupToggles(backupToggleMenuOptions{
			CaptureGTID:      true,
			ExportUserGrants: true,
			ExcludeSystem:    true,
			ExcludeEmpty:     true,
			ExcludeData:      true,
		})...)
		items = append(items, s.editItemsBackupOutputSecurity(customOutputDir, true)...)
	case consts.ModeSingle:
		items = append(items, editMenuItem{Label: "Database selection", Action: func() error { return s.changeBackupDatabaseSelectionResetInteractive(false) }})
		items = append(items, s.editItemsBackupToggles(backupToggleMenuOptions{
			ExportUserGrants: true,
			ExcludeData:      true,
		})...)
		items = append(items, s.editItemsBackupOutputSecurity(customOutputDir, true)...)
	case consts.ModePrimary:
		items = append(items,
			editMenuItem{Label: "Client code", Action: func() error { return s.changeBackupClientCodeInteractive() }},
			editMenuItem{Label: "Database selection", Action: func() error { return s.changeBackupDatabaseSelectionResetInteractive(true) }},
			editMenuItem{Label: "Include DMart", Action: func() error { return s.changeBackupIncludeDmartInteractive() }},
		)
		items = append(items, s.editItemsBackupToggles(backupToggleMenuOptions{
			ExportUserGrants: true,
			ExcludeData:      true,
		})...)
		items = append(items, s.editItemsBackupOutputSecurity(customOutputDir, true)...)
	case consts.ModeSecondary:
		items = append(items,
			editMenuItem{Label: "Client code", Action: func() error { return s.changeBackupClientCodeInteractive() }},
			editMenuItem{Label: "Instance", Action: func() error { return s.changeBackupSecondaryInstance() }},
			editMenuItem{Label: "Database selection", Action: func() error { return s.changeBackupDatabaseSelectionResetInteractive(true) }},
			editMenuItem{Label: "Include DMart", Action: func() error { return s.changeBackupIncludeDmartInteractive() }},
		)
		items = append(items, s.editItemsBackupToggles(backupToggleMenuOptions{
			ExportUserGrants: true,
			ExcludeData:      true,
		})...)
		items = append(items, s.editItemsBackupOutputSecurity(customOutputDir, true)...)
	case consts.ModeCombined:
		items = append(items, includeSelectionItems...)
		items = append(items, s.editItemsBackupToggles(backupToggleMenuOptions{
			CaptureGTID:      true,
			ExportUserGrants: true,
			ExcludeSystem:    true,
			ExcludeEmpty:     true,
			ExcludeData:      true,
		})...)
		items = append(items, s.editItemsBackupOutputSecurity(customOutputDir, true)...)
	case consts.ModeSeparated:
		items = append(items, includeSelectionItems...)
		items = append(items, s.editItemsBackupToggles(backupToggleMenuOptions{
			ExportUserGrants: true,
			ExcludeSystem:    true,
			ExcludeEmpty:     true,
			ExcludeData:      true,
		})...)
		items = append(items, s.editItemsBackupOutputSecurity(customOutputDir, false)...)
	default:
		return fmt.Errorf("mode tidak mendukung edit interaktif: %s", mode)
	}

	return s.runEditMenuInteractive(items)
}
