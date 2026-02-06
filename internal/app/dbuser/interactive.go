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
	changeAdvanced, err := prompt.Confirm("Ubah pengaturan lanjutan? (exclude system users / include create user)", false)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if changeAdvanced {
		excl, eerr := prompt.Confirm("Exclude system users?", opts.ExcludeSystemUsers)
		if eerr != nil {
			return validation.HandleInputError(eerr)
		}
		opts.ExcludeSystemUsers = excl

		incCreate, cerr := prompt.Confirm("Include CREATE USER (best-effort)?", opts.IncludeCreateUser)
		if cerr != nil {
			return validation.HandleInputError(cerr)
		}
		opts.IncludeCreateUser = incCreate
	}

	return nil
}

func completeApplyOptionsInteractive(deps *appdeps.Dependencies, opts *ApplyOptions) error {
	if opts == nil {
		return fmt.Errorf("opts nil")
	}

	// Pilih file jika belum ada
	if strings.TrimSpace(opts.File) == "" {
		startDir := "."
		if deps != nil && deps.Config != nil && strings.TrimSpace(deps.Config.Backup.Output.BaseDirectory) != "" {
			startDir = deps.Config.Backup.Output.BaseDirectory
		}
		selected, err := prompt.SelectFile(startDir, "Pilih file SQL user+grants", []string{".sql"})
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
		opts.File = strings.TrimSpace(selected)
	}

	// Konfirmasi force
	force, err := prompt.Confirm("Gunakan mode --force (-f) agar best-effort?", opts.Force)
	if err != nil {
		return validation.HandleInputError(err)
	}
	opts.Force = force

	return nil
}
