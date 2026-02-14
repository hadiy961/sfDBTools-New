package dbuser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/app/usersgrants"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"strings"

	"github.com/mattn/go-isatty"
)

func interactiveAllowed() bool {
	if runtimecfg.IsQuiet() {
		return false
	}
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

func resolveSourceProfile(deps *appdeps.Dependencies, opts *ExportOptions) error {
	if opts == nil {
		return fmt.Errorf("opts nil")
	}
	if deps == nil || deps.Config == nil {
		return fmt.Errorf("config tidak tersedia")
	}

	prof, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:         deps.Config.ConfigDir.DatabaseProfile,
		ProfilePath:       strings.TrimSpace(opts.Profile.Path),
		ProfileKey:        strings.TrimSpace(opts.Profile.EncryptionKey),
		EnvProfilePath:    consts.ENV_SOURCE_PROFILE,
		EnvProfileKey:     consts.ENV_SOURCE_PROFILE_KEY,
		RequireProfile:    true,
		ProfilePurpose:    "source",
		AllowInteractive:  true,
		InteractivePrompt: "Pilih file konfigurasi database source:",
	})
	if err != nil {
		return err
	}
	if prof == nil {
		return fmt.Errorf("%sprofile source tidak tersedia (hasil load nil)", consts.ProfileMsgNonInteractivePrefix)
	}
	opts.Profile = *prof
	return nil
}

func resolveTargetProfile(deps *appdeps.Dependencies, opts *ApplyOptions) error {
	if opts == nil {
		return fmt.Errorf("opts nil")
	}
	if deps == nil || deps.Config == nil {
		return fmt.Errorf("config tidak tersedia")
	}

	prof, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:         deps.Config.ConfigDir.DatabaseProfile,
		ProfilePath:       strings.TrimSpace(opts.Profile.Path),
		ProfileKey:        strings.TrimSpace(opts.Profile.EncryptionKey),
		EnvProfilePath:    consts.ENV_TARGET_PROFILE,
		EnvProfileKey:     consts.ENV_TARGET_PROFILE_KEY,
		RequireProfile:    true,
		ProfilePurpose:    "target",
		AllowInteractive:  true,
		InteractivePrompt: "Pilih file konfigurasi database target:",
	})
	if err != nil {
		return err
	}
	if prof == nil {
		return fmt.Errorf("%sprofile target tidak tersedia (hasil load nil)", consts.ProfileMsgNonInteractivePrefix)
	}
	opts.Profile = *prof
	return nil
}

func completeExportOptionsInteractive(ctx context.Context, deps *appdeps.Dependencies, client *database.Client, opts *ExportOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts == nil {
		return fmt.Errorf("opts nil")
	}

	// Default values from config
	defaultCreateUser := true
	defaultGrants := true
	defaultSplit := false
	if deps != nil && deps.Config != nil {
		defaultCreateUser = deps.Config.DBUser.Export.IncludeCreateUser
		defaultGrants = deps.Config.DBUser.Export.IncludeGrants
		defaultSplit = deps.Config.DBUser.Export.SplitOutput
	}

	// Jika user belum menentukan scope apapun, prompt.
	noScope := len(opts.Users) == 0 && len(opts.Databases) == 0 && strings.TrimSpace(opts.DBFile) == "" && strings.TrimSpace(opts.ClientCode) == ""
	if noScope {
		choice, _, err := prompt.SelectOne(
			"Pilih scope export user/grants",
			[]string{
				"Semua user + semua grants",
				"Pilih user tertentu",
				"Filter grants berdasarkan database",
				"Filter grants berdasarkan client-code",
			},
			0,
		)
		if err != nil {
			return validation.HandleInputError(err)
		}

		switch choice {
		case "Semua user + semua grants":
			// no-op
		case "Pilih user tertentu":
			accounts, rerr := usersgrants.ResolveUserAccounts(ctx, client, usersgrants.ExportOptions{})
			if rerr != nil {
				return rerr
			}
			items := make([]string, 0, len(accounts))
			for _, a := range accounts {
				items = append(items, fmt.Sprintf("%s@%s", a.User, a.Host))
			}
			selected, _, selErr := prompt.SelectMany("Pilih user accounts (space untuk toggle)", items, nil)
			if selErr != nil {
				return validation.HandleInputError(selErr)
			}
			opts.Users = selected
		case "Filter grants berdasarkan database":
			dbs, derr := client.GetDatabaseList(ctx)
			if derr != nil {
				return derr
			}
			selected, _, selErr := prompt.SelectMany("Pilih database (space untuk toggle)", dbs, nil)
			if selErr != nil {
				return validation.HandleInputError(selErr)
			}
			opts.Databases = selected
		case "Filter grants berdasarkan client-code":
			cc, aerr := prompt.AskText("Masukkan client-code")
			if aerr != nil {
				return validation.HandleInputError(aerr)
			}
			opts.ClientCode = strings.TrimSpace(cc)
		}
	}

	// Konfirmasi output path jika belum diisi.
	if strings.TrimSpace(opts.OutPath) == "" && deps != nil && deps.Config != nil {
		baseDir := deps.Config.Backup.Output.BaseDirectory
		if strings.TrimSpace(baseDir) == "" {
			baseDir = "."
		}
		// gunakan default naming (di ExecuteExport) tapi tampilkan preview.
		suggested := filepath.Join(baseDir, "user_grants_<timestamp>.sql")
		ok, err := prompt.Confirm(fmt.Sprintf("Simpan output ke default folder: %s ?", baseDir), true)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if !ok {
			out, terr := prompt.AskText("Masukkan output path (.sql)", prompt.WithDefault(suggested))
			if terr != nil {
				return validation.HandleInputError(terr)
			}
			opts.OutPath = strings.TrimSpace(out)
		}
	}

	// Pengaturan lanjutan (opsional)
	changeAdvanced, err := prompt.Confirm("Ubah pengaturan lanjutan? (exclude system users / include create user / include grants / split output)", false)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if changeAdvanced {
		excl, eerr := prompt.Confirm("Exclude system users?", opts.ExcludeSystemUsers)
		if eerr != nil {
			return validation.HandleInputError(eerr)
		}
		opts.ExcludeSystemUsers = excl

		incCreate, cerr := prompt.Confirm("Include CREATE USER (best-effort)?", defaultCreateUser)
		if cerr != nil {
			return validation.HandleInputError(cerr)
		}
		opts.IncludeCreateUser = incCreate

		incGrants, gerr := prompt.Confirm("Include GRANT statements?", defaultGrants)
		if gerr != nil {
			return validation.HandleInputError(gerr)
		}
		opts.IncludeGrants = incGrants

		// Split output hanya relevan jika CREATE USER dan GRANT keduanya di-include.
		if opts.IncludeCreateUser && opts.IncludeGrants {
			split, serr := prompt.Confirm("Split output jadi 2 file terpisah? (*.users.sql dan *.grants.sql)", defaultSplit)
			if serr != nil {
				return validation.HandleInputError(serr)
			}
			opts.SplitOut = split
		} else {
			opts.SplitOut = false
		}
	} else {
		// Respect defaults if not changed interactively
		opts.IncludeCreateUser = defaultCreateUser
		opts.IncludeGrants = defaultGrants
		if opts.IncludeCreateUser && opts.IncludeGrants {
			opts.SplitOut = defaultSplit
		} else {
			opts.SplitOut = false
		}
	}

	return nil
}

func completeApplyOptionsInteractive(deps *appdeps.Dependencies, opts *ApplyOptions) error {
	if opts == nil {
		return fmt.Errorf("opts nil")
	}

	// Pilih file jika belum ada (support multi-file)
	if len(opts.Files) == 0 {
		startDir := "."
		if deps != nil && deps.Config != nil && strings.TrimSpace(deps.Config.Backup.Output.BaseDirectory) != "" {
			startDir = deps.Config.Backup.Output.BaseDirectory
		}

		for {
			selected, err := prompt.SelectFile(startDir, "Pilih file SQL users/grants", []string{".sql"})
			if err != nil {
				return validation.HandleInputError(err)
			}
			if strings.TrimSpace(selected) == "" {
				manual, terr := prompt.AskText("Masukkan path file SQL")
				if terr != nil {
					return validation.HandleInputError(terr)
				}
				selected = manual
			}
			selected = strings.TrimSpace(selected)
			if selected != "" {
				opts.Files = append(opts.Files, selected)
			}

			addMore, aerr := prompt.Confirm("Tambah file SQL lain?", false)
			if aerr != nil {
				return validation.HandleInputError(aerr)
			}
			if !addMore {
				break
			}
		}
	}

	// Konfirmasi force
	force, err := prompt.Confirm("Gunakan mode --force (-f) agar best-effort?", opts.Force)
	if err != nil {
		return validation.HandleInputError(err)
	}
	opts.Force = force

	// Precheck grants-only: default fail-fast jika user belum ada.
	skip, err := prompt.Confirm("Skip precheck user existence untuk grants-only?", opts.SkipUserCheck)
	if err != nil {
		return validation.HandleInputError(err)
	}
	opts.SkipUserCheck = skip

	return nil
}

func showExportPreview(opts *ExportOptions, dbList []string) (bool, error) {
	print.PrintInfo("=== Export Preview ===")

	fmt.Printf("Source Profile : %s (%s)\n", opts.Profile.Name, opts.Profile.DBInfo.Host)

	outPath := opts.OutPath
	if opts.SplitOut && opts.IncludeCreateUser && opts.IncludeGrants {
		dir := filepath.Dir(outPath)
		base := filepath.Base(outPath)
		ext := filepath.Ext(base)
		n := strings.TrimSuffix(base, ext)
		fmt.Printf("Output Path    : %s (SPLIT OUTPUT)\n", dir)
		fmt.Printf("               - %s.users.sql\n", n+".users.sql")
		fmt.Printf("               - %s.grants.sql\n", n+".grants.sql")
	} else {
		fmt.Printf("Output Path    : %s\n", outPath)
	}

	scope := "All Users & Grants"
	if len(opts.Users) > 0 {
		scope = fmt.Sprintf("Specific Users (%d accounts)", len(opts.Users))
	} else if len(opts.Databases) > 0 || strings.TrimSpace(opts.DBFile) != "" || strings.TrimSpace(opts.ClientCode) != "" { 
		scope = fmt.Sprintf("Filtered by Database (%d DBs resolved)", len(dbList))
	}
	fmt.Printf("Scope          : %s\n", scope)

	optsList := []string{}
	if opts.IncludeCreateUser {
		optsList = append(optsList, "Create User")
	}
	if opts.IncludeGrants {
		optsList = append(optsList, "Grants")
	}
	if opts.ExcludeSystemUsers {
		optsList = append(optsList, "Exclude System Users")
	}
	fmt.Printf("Options        : %s\n", strings.Join(optsList, ", "))

	fmt.Println("----------------------")
	return prompt.Confirm("Lanjutkan proses export?", true)
}
