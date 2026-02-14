package dbuser

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func splitOutPaths(outPath string) (usersPath string, grantsPath string) {
	p := strings.TrimSpace(outPath)
	if p == "" {
		return "", ""
	}
	base := p
	if strings.EqualFold(filepath.Ext(base), ".sql") {
		base = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return base + ".users.sql", base + ".grants.sql"
}

func looksLikeGrantsOnlySQL(sqlText string) bool {
	up := strings.ToUpper(sqlText)
	if strings.Contains(up, "CREATE USER") {
		return false
	}
	return strings.Contains(up, "GRANT ")
}

func extractGrantAccounts(sqlText string) []usersgrants.UserAccount {
	// Best-effort: parse pola umum dari SHOW GRANTS:
	//   GRANT ... TO 'user'@'host';
	// Catatan: kita sengaja tidak menangani semua variasi syntax (role/proxy/dll).
	re := regexp.MustCompile(`(?i)\bTO\s+'([^']+)'\s*@\s*'([^']+)'`)
	ms := re.FindAllStringSubmatch(sqlText, -1)
	if len(ms) == 0 {
		return nil
	}
	out := make([]usersgrants.UserAccount, 0, len(ms))
	seen := make(map[string]struct{}, len(ms))
	for _, m := range ms {
		if len(m) < 3 {
			continue
		}
		u := strings.TrimSpace(m[1])
		h := strings.TrimSpace(m[2])
		if u == "" || h == "" {
			continue
		}
		key := u + "@" + h
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, usersgrants.UserAccount{User: u, Host: h})
	}
	return out
}

func missingAccounts(ctx context.Context, client *database.Client, accounts []usersgrants.UserAccount) ([]usersgrants.UserAccount, error) {
	if client == nil {
		return nil, fmt.Errorf("client nil")
	}
	if len(accounts) == 0 {
		return nil, nil
	}

	// Best-effort query; butuh privilege SELECT ke mysql.user.
	const q = `SELECT 1 FROM mysql.user WHERE user = ? AND host = ? LIMIT 1`
	missing := make([]usersgrants.UserAccount, 0)
	for _, a := range accounts {
		var one int
		err := client.DB().QueryRowContext(ctx, q, a.User, a.Host).Scan(&one)
		if err == nil {
			continue
		}
		// Jika no rows, account belum ada.
		if err == sql.ErrNoRows {
			missing = append(missing, a)
			continue
		}
		return nil, err
	}
	return missing, nil
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
	if !parsed.IncludeCreateUser && !parsed.IncludeGrants {
		return fmt.Errorf("tidak ada yang diexport: --include-create-user dan --include-grants keduanya false")
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

	writeOut := func(path string, content string) error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("gagal membuat direktori output: %w", err)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			return fmt.Errorf("gagal menulis output file: %w", err)
		}
		return nil
	}

	// Split output (users + grants) jika diminta dan keduanya aktif.
	if parsed.SplitOut && parsed.IncludeCreateUser && parsed.IncludeGrants {
		usersPath, grantsPath := splitOutPaths(outPath)
		usersSQL, usersStats, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
			Users:              parseAccounts(parsed.Users),
			Databases:          dbFilters,
			ExcludeSystemUsers: parsed.ExcludeSystemUsers,
			SystemUsers:        sysUsers,
			IncludeCreateUser:  true,
			IncludeGrants:      false,
			FlushPrivileges:    false,
		})
		if err != nil {
			return err
		}
		grantsSQL, grantsStats, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
			Users:              parseAccounts(parsed.Users),
			Databases:          dbFilters,
			ExcludeSystemUsers: parsed.ExcludeSystemUsers,
			SystemUsers:        sysUsers,
			IncludeCreateUser:  false,
			IncludeGrants:      true,
			FlushPrivileges:    true,
		})
		if err != nil {
			return err
		}
		if err := writeOut(usersPath, usersSQL); err != nil {
			return err
		}
		if err := writeOut(grantsPath, grantsSQL); err != nil {
			return err
		}

		if !runtimecfg.IsQuiet() {
			print.PrintSuccess(fmt.Sprintf("Export users selesai: %s", usersPath))
			print.PrintSuccess(fmt.Sprintf("Export grants selesai: %s", grantsPath))
		}
		log.Infof("Export users selesai: file=%s total_users=%d written=%d skipped=%d grant_lines=%d warnings=%d",
			usersPath, usersStats.TotalUsersInput, usersStats.TotalUsersWritten, usersStats.TotalUsersSkipped, usersStats.TotalGrantLines, usersStats.Warnings)
		log.Infof("Export grants selesai: file=%s total_users=%d written=%d skipped=%d grant_lines=%d warnings=%d",
			grantsPath, grantsStats.TotalUsersInput, grantsStats.TotalUsersWritten, grantsStats.TotalUsersSkipped, grantsStats.TotalGrantLines, grantsStats.Warnings)
		return nil
	}

	sqlText, stats, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
		Users:              parseAccounts(parsed.Users),
		Databases:          dbFilters,
		ExcludeSystemUsers: parsed.ExcludeSystemUsers,
		SystemUsers:        sysUsers,
		IncludeCreateUser:  parsed.IncludeCreateUser,
		IncludeGrants:      parsed.IncludeGrants,
		FlushPrivileges:    parsed.IncludeGrants,
	})
	if err != nil {
		return err
	}
	if err := writeOut(outPath, sqlText); err != nil {
		return err
	}

	if !runtimecfg.IsQuiet() {
		print.PrintSuccess(fmt.Sprintf("Export selesai: %s", outPath))
	}
	log.Infof("Export selesai: file=%s total_users=%d written=%d skipped=%d grant_lines=%d warnings=%d",
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
	if len(parsed.Files) == 0 {
		return fmt.Errorf("--file wajib diisi")
	}

	extra := make([]string, 0, 1)
	if parsed.Force {
		extra = append(extra, "-f")
	}

	ctx := context.Background()

	// Optional precheck untuk grants-only files: fail-fast jika user target belum ada.
	// Catatan: ini best-effort dan butuh privilege baca mysql.user.
	var precheckClient *database.Client
	getPrecheckClient := func() (*database.Client, error) {
		if precheckClient != nil {
			return precheckClient, nil
		}
		c, err := profileconn.ConnectWithProfile(nil, &parsed.Profile, consts.DefaultInitialDatabase)
		if err != nil {
			return nil, err
		}
		precheckClient = c
		return precheckClient, nil
	}
	defer func() {
		if precheckClient != nil {
			precheckClient.Close()
		}
	}()

	for _, f := range parsed.Files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("gagal membaca file: %w", err)
		}
		sqlText := string(sqlBytes)
		if looksLikeGrantsOnlySQL(sqlText) {
			if !runtimecfg.IsQuiet() {
				print.PrintWarning("⚠️  File terlihat berisi GRANT tanpa CREATE USER.")
			}
			if !parsed.SkipUserCheck {
				accs := extractGrantAccounts(sqlText)
				if len(accs) > 0 {
					c, cerr := getPrecheckClient()
					if cerr != nil {
						// Jika tidak bisa connect untuk precheck, fallback ke behavior lama (warning + lanjut).
						log.Warnf("Precheck user grants dilewati (gagal connect): %v", cerr)
					} else {
						miss, merr := missingAccounts(ctx, c, accs)
						if merr != nil {
							log.Warnf("Precheck user grants dilewati (gagal cek mysql.user): %v", merr)
						} else if len(miss) > 0 {
							// Fail-fast: user target belum ada.
							specs := make([]string, 0, len(miss))
							for _, m := range miss {
								specs = append(specs, fmt.Sprintf("%s@%s", m.User, m.Host))
							}
							return fmt.Errorf("apply dibatalkan: file grants-only membutuhkan user target yang belum ada: %s (apply users dulu atau pakai --skip-user-check)", strings.Join(specs, ", "))
						}
					}
				}
			}
		}

		args := mysqlcli.BuildArgs(&parsed.Profile, "", extra...)
		sum, err := mysqlcli.Execute(ctx, args, strings.NewReader(sqlText), log)
		if err != nil && mysqlcli.IsSSLMismatchServerNotSupport(err) && !mysqlcli.HasSkipSSLArg(args) {
			retryArgs := mysqlcli.BuildArgs(&parsed.Profile, "", append([]string{"--skip-ssl"}, extra...)...)
			sum2, err2 := mysqlcli.Execute(ctx, retryArgs, strings.NewReader(sqlText), log)
			sum = sum2
			if err2 != nil {
				return fmt.Errorf("gagal apply file %s (retry --skip-ssl): %w", f, err2)
			}
			err = nil
		}
		if sum != nil {
			log.Infof("Apply file summary: file=%s sql_errors=%d sql_warnings=%d other_outputs=%d", f, sum.SQLErrors, sum.SQLWarnings, sum.OtherOutputs)
		}
		if err != nil {
			return fmt.Errorf("gagal apply file %s: %w", f, err)
		}
	}
	if !runtimecfg.IsQuiet() {
		print.PrintSuccess("Apply SQL selesai")
	}
	return nil
}
