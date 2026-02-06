package dbuser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/mysqlcli"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/usersgrants"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/listx"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func defaultSystemUsers() []string {
	// Conservatively exclude MySQL internal accounts (common across MySQL/MariaDB).
	return []string{
		"mysql.sys",
		"mysql.session",
		"mysql.infoschema",
	}
}

func parseAccounts(userSpecs []string) []usersgrants.UserAccount {
	accs := make([]usersgrants.UserAccount, 0, len(userSpecs))
	for _, s := range userSpecs {
		parts := strings.SplitN(strings.TrimSpace(s), "@", 2)
		if len(parts) != 2 {
			continue
		}
		accs = append(accs, usersgrants.UserAccount{User: strings.TrimSpace(parts[0]), Host: strings.TrimSpace(parts[1])})
	}
	return accs
}

func parseFileMode(permStr string) (os.FileMode, error) {
	if strings.TrimSpace(permStr) == "" {
		return 0o600, nil
	}
	v, err := strconv.ParseUint(permStr, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(v), nil
}

func resolveDatabaseFilters(ctx context.Context, client *database.Client, opts ExportOptions) ([]string, error) {
	dbs := make([]string, 0)
	if len(opts.Databases) > 0 {
		dbs = append(dbs, opts.Databases...)
	}
	if strings.TrimSpace(opts.DBFile) != "" {
		lines, err := fsops.ReadLinesFromFile(opts.DBFile)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca --db-file: %w", err)
		}
		dbs = append(dbs, lines...)
	}
	dbs = listx.ListTrimNonEmpty(dbs)
	dbs = listx.ListUnique(dbs)

	// client-code: derive database list dari server (best-effort) dan gunakan sebagai filter grants
	if strings.TrimSpace(opts.ClientCode) != "" {
		all, err := client.GetDatabaseList(ctx)
		if err != nil {
			return nil, fmt.Errorf("gagal mengambil daftar database untuk client-code: %w", err)
		}
		// pakai selection filter existing (pattern primary NBC)
		filtered := make([]string, 0, len(all))
		ccLower := strings.ToLower(strings.TrimSpace(opts.ClientCode))
		prefix := consts.PrimaryPrefixNBC + ccLower
		for _, name := range all {
			n := strings.ToLower(strings.TrimSpace(name))
			if n == "" {
				continue
			}
			if n == prefix || strings.HasPrefix(n, prefix+"_") {
				// exclude secondary/temp/archive/dmart by convention (best-effort)
				if !strings.Contains(n, consts.SecondarySuffix) &&
					!strings.HasSuffix(n, consts.SuffixDmart) &&
					!strings.HasSuffix(n, consts.SuffixTemp) &&
					!strings.HasSuffix(n, consts.SuffixArchive) {
					filtered = append(filtered, name)
				}
			}
		}
		filtered = listx.ListUnique(listx.ListTrimNonEmpty(filtered))
		if len(filtered) == 0 {
			return nil, fmt.Errorf("tidak ada database yang match client-code %q", opts.ClientCode)
		}
		dbs = append(dbs, filtered...)
		dbs = listx.ListUnique(listx.ListTrimNonEmpty(dbs))
	}

	return dbs, nil
}

func ExecuteExport(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	if deps == nil || deps.Logger == nil || deps.Config == nil {
		return fmt.Errorf("dependencies/config/logger tidak tersedia")
	}
	log := deps.Logger

	parsed, err := ParseExportOptions(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Interactive completion: profile selection + scope selection, only if TTY and not quiet.
	if interactiveAllowed() {
		if err := resolveSourceProfile(deps, &parsed); err != nil {
			return err
		}
	}

	client, err := profileconn.ConnectWithProfile(nil, &parsed.Profile, consts.DefaultInitialDatabase)
	if err != nil {
		return err
	}
	defer client.Close()

	if interactiveAllowed() {
		if err := completeExportOptionsInteractive(ctx, deps, client, &parsed); err != nil {
			if err == validation.ErrUserCancelled {
				log.Warn("Proses dibatalkan oleh pengguna.")
				return nil
			}
			return err
		}
	}

	dbFilters, err := resolveDatabaseFilters(ctx, client, parsed)
	if err != nil {
		return err
	}

	sysUsers := defaultSystemUsers()
	if deps.Config != nil && len(deps.Config.SystemUsers.Users) > 0 {
		sysUsers = append(sysUsers, deps.Config.SystemUsers.Users...)
	}

	sqlText, stats, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
		Users:              parseAccounts(parsed.Users),
		Databases:          dbFilters,
		ExcludeSystemUsers: parsed.ExcludeSystemUsers,
		SystemUsers:        sysUsers,
		IncludeCreateUser:  parsed.IncludeCreateUser,
		FlushPrivileges:    true,
	})
	if err != nil {
		return err
	}

	outPath := strings.TrimSpace(parsed.OutPath)
	if outPath == "" {
		baseDir := deps.Config.Backup.Output.BaseDirectory
		if strings.TrimSpace(baseDir) == "" {
			baseDir = "."
		}
		ts := time.Now().Format("20060102_150405")
		outPath = filepath.Join(baseDir, fmt.Sprintf("user_grants_%s.sql", ts))
	}

	perm := deps.Config.Backup.Output.MetadataPermissions
	if strings.TrimSpace(perm) == "" {
		perm = parsed.OutPerm
	}
	mode, perr := parseFileMode(perm)
	if perr != nil {
		return fmt.Errorf("out-perm tidak valid (%q): %w", perm, perr)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}
	if err := os.WriteFile(outPath, []byte(sqlText), mode); err != nil {
		return fmt.Errorf("gagal menulis output file: %w", err)
	}

	if !runtimecfg.IsQuiet() {
		print.PrintSuccess(fmt.Sprintf("Export user+grants selesai: %s", outPath))
	}
	log.Infof("Export user+grants selesai: file=%s total_users=%d written=%d skipped=%d grant_lines=%d warnings=%d",
		outPath, stats.TotalUsersInput, stats.TotalUsersWritten, stats.TotalUsersSkipped, stats.TotalGrantLines, stats.Warnings)
	return nil
}

func ExecuteApply(cmd *cobra.Command, deps *appdeps.Dependencies) error {
	if deps == nil || deps.Logger == nil {
		return fmt.Errorf("dependencies/logger tidak tersedia")
	}
	log := deps.Logger

	parsed, err := ParseApplyOptions(cmd)
	if err != nil {
		return err
	}

	if interactiveAllowed() {
		if err := resolveTargetProfile(deps, &parsed); err != nil {
			return err
		}
		if err := completeApplyOptionsInteractive(deps, &parsed); err != nil {
			if err == validation.ErrUserCancelled {
				log.Warn("Proses dibatalkan oleh pengguna.")
				return nil
			}
			return err
		}
	}
	if strings.TrimSpace(parsed.File) == "" {
		return fmt.Errorf("--file wajib diisi")
	}

	sqlBytes, err := os.ReadFile(parsed.File)
	if err != nil {
		return fmt.Errorf("gagal membaca file: %w", err)
	}
	sqlText := string(sqlBytes)

	extra := make([]string, 0, 1)
	if parsed.Force {
		extra = append(extra, "-f")
	}

	ctx := context.Background()
	args := mysqlcli.BuildArgs(&parsed.Profile, "", extra...)
	sum, err := mysqlcli.Execute(ctx, args, strings.NewReader(sqlText), log)
	if err != nil && mysqlcli.IsSSLMismatchServerNotSupport(err) && !mysqlcli.HasSkipSSLArg(args) {
		retryArgs := mysqlcli.BuildArgs(&parsed.Profile, "", append([]string{"--skip-ssl"}, extra...)...)
		sum2, err2 := mysqlcli.Execute(ctx, retryArgs, strings.NewReader(sqlText), log)
		sum = sum2
		if err2 != nil {
			return fmt.Errorf("gagal apply user+grants (retry --skip-ssl): %w", err2)
		}
		err = nil
	}
	if sum != nil {
		log.Infof("Apply user+grants summary: sql_errors=%d sql_warnings=%d other_outputs=%d", sum.SQLErrors, sum.SQLWarnings, sum.OtherOutputs)
	}
	if err != nil {
		return fmt.Errorf("gagal apply user+grants: %w", err)
	}

	if !runtimecfg.IsQuiet() {
		print.PrintSuccess("Apply user+grants selesai")
	}
	return nil
}
