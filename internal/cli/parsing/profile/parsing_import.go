package profile

import (
	"fmt"
	profilemodel "sfdbtools/internal/app/profile/model"
	parsingcommon "sfdbtools/internal/cli/parsing/common"
	resolver "sfdbtools/internal/cli/resolver"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingImportProfile parses flags untuk profile import command
func ParsingImportProfile(cmd *cobra.Command) (*profilemodel.ProfileImportOptions, error) {
	input := resolver.GetStringFlagOrEnv(cmd, "input", "")
	sheet := resolver.GetStringFlagOrEnv(cmd, "sheet", "")
	gsheet := resolver.GetStringFlagOrEnv(cmd, "gsheet", "")
	gid := resolver.GetIntFlagOrEnv(cmd, "gid", "")
	onConflict := resolver.GetStringFlagOrEnv(cmd, "on-conflict", "")
	skipConfirm := resolver.GetBoolFlagOrEnv(cmd, "skip-confirm", "")
	skipInvalidRows := resolver.GetBoolFlagOrEnv(cmd, "skip-invalid-rows", "")
	continueOnError := resolver.GetBoolFlagOrEnv(cmd, "continue-on-error", "")
	skipConnTest := resolver.GetBoolFlagOrEnv(cmd, "skip-conn-test", "")

	interactive := parsingcommon.IsInteractiveMode()
	// Jika mode non-interaktif, wajib skip-confirm agar tidak hang.
	if !interactive && !skipConfirm {
		return nil, fmt.Errorf("mode non-interaktif: flag --skip-confirm wajib disertakan untuk automation")
	}

	if strings.TrimSpace(input) != "" && strings.TrimSpace(gsheet) != "" {
		return nil, fmt.Errorf("gunakan salah satu: --input (XLSX lokal) atau --gsheet (Google Spreadsheet), bukan keduanya")
	}
	// Full-interactive mode: bila belum ada sumber, akan dipilih via prompt di executor.
	// Non-interaktif/automation tetap wajib menentukan sumber via flag.
	if strings.TrimSpace(input) == "" && strings.TrimSpace(gsheet) == "" {
		if !interactive || skipConfirm {
			return nil, fmt.Errorf("sumber import tidak tersedia: gunakan --input <file.xlsx> atau --gsheet <url>")
		}
	}

	// Untuk import, interactive runtime ditentukan oleh TTY + bukan quiet + bukan skip-confirm.
	interactive = interactive && !skipConfirm

	// Jika user memilih Google Sheet tapi tidak set --gid secara eksplisit,
	// tandai gid sebagai "unspecified" (-1) untuk memicu prompt selection.
	// (di mode automation / --skip-confirm, kita tetap pakai default 0)
	if interactive && strings.TrimSpace(gsheet) != "" {
		if !cmd.Flags().Changed("gid") {
			gid = -1
		}
	}

	return &profilemodel.ProfileImportOptions{
		Input:           strings.TrimSpace(input),
		Sheet:           strings.TrimSpace(sheet),
		GSheetURL:       strings.TrimSpace(gsheet),
		GID:             gid,
		OnConflict:      strings.TrimSpace(onConflict),
		SkipConfirm:     skipConfirm,
		SkipInvalidRows: skipInvalidRows,
		ContinueOnError: continueOnError,
		SkipConnTest:    skipConnTest,
		Interactive:     interactive,
	}, nil
}
