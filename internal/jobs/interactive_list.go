// File : internal/jobs/interactive_list.go
// Deskripsi : Mode interaktif khusus untuk subcommand jobs list
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package jobs

import (
	"context"

	"sfDBTools/internal/services/scheduler"
)

func RunInteractiveList(ctx context.Context, defaultScope schedulerutil.Scope, isRoot bool) error {
	scope, err := pickScopeInteractive(defaultScope)
	if err != nil {
		return err
	}
	return PrintListBody(ctx, scope, isRoot)
}
