// File : internal/jobs/interactive_list.go
// Deskripsi : Mode interaktif khusus untuk subcommand jobs list
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-04
// Last Modified :  2026-01-05
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
