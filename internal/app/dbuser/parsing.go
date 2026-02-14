package dbuser

import (
	"fmt"
	"sfdbtools/internal/cli/parsing"
	resolver "sfdbtools/internal/cli/resolver"
	"strings"

	"github.com/spf13/cobra"
)

func parseUserSpecs(specs []string) ([]string, error) {
	out := make([]string, 0, len(specs))
	for _, s := range specs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "@", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("format --user tidak valid: %q (harus user@host)", s)
		}
		out = append(out, s)
	}
	return out, nil
}

func ParseExportOptions(cmd *cobra.Command) (ExportOptions, error) {
	opts := ExportOptions{
		ExcludeSystemUsers: true,
		IncludeCreateUser:  true,
		IncludeGrants:      true,
		OutPerm:            "0600",
	}

	// Profile source (backup-style)
	if err := parsing.PopulateProfileFlags(cmd, &opts.Profile); err != nil {
		return ExportOptions{}, err
	}

	if v := resolver.GetStringFlagOrEnv(cmd, "out", ""); v != "" {
		opts.OutPath = strings.TrimSpace(v)
	}
	if v := resolver.GetStringFlagOrEnv(cmd, "out-perm", ""); v != "" {
		opts.OutPerm = strings.TrimSpace(v)
	}

	opts.ExcludeSystemUsers = resolver.GetBoolFlagOrEnv(cmd, "exclude-system-users", "")
	opts.IncludeCreateUser = resolver.GetBoolFlagOrEnv(cmd, "include-create-user", "")
	opts.IncludeGrants = resolver.GetBoolFlagOrEnv(cmd, "include-grants", "")
	opts.SplitOut = resolver.GetBoolFlagOrEnv(cmd, "split-out", "")

	users := resolver.GetStringArrayFlagOrEnv(cmd, "user", "")
	parsedUsers, err := parseUserSpecs(users)
	if err != nil {
		return ExportOptions{}, err
	}
	opts.Users = parsedUsers

	opts.Databases = resolver.GetStringArrayFlagOrEnv(cmd, "db", "")
	if v := resolver.GetStringFlagOrEnv(cmd, "db-file", ""); v != "" {
		opts.DBFile = strings.TrimSpace(v)
	}
	if v := resolver.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
		opts.ClientCode = strings.TrimSpace(v)
	}

	return opts, nil
}

func ParseApplyOptions(cmd *cobra.Command) (ApplyOptions, error) {
	opts := ApplyOptions{
		Force: true,
	}
	// Profile target (restore-style)
	if err := parsing.PopulateTargetProfileFlags(cmd, &opts.Profile); err != nil {
		return ApplyOptions{}, err
	}
	files := resolver.GetStringArrayFlagOrEnv(cmd, "file", "")
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		opts.Files = append(opts.Files, f)
	}
	opts.Force = resolver.GetBoolFlagOrEnv(cmd, "force", "")
	opts.SkipUserCheck = resolver.GetBoolFlagOrEnv(cmd, "skip-user-check", "")
	return opts, nil
}
