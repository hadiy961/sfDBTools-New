// File : internal/profile/executor/clone.go
// Deskripsi : Eksekusi clone profile
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package executor

import (
	"fmt"
	profiledisplay "sfdbtools/internal/app/profile/display"
	profilemodel "sfdbtools/internal/app/profile/model"
	profilevalidation "sfdbtools/internal/app/profile/validation"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"strings"
)

func (e *Executor) CloneProfile() error {
	isInteractive := e.isInteractiveMode()
	if !isInteractive {
		e.Log.Info("Clone profil dimulai...")
	}

	cloneOpts, ok := e.State.CloneOptions()
	if !ok || cloneOpts == nil {
		return fmt.Errorf("clone options tidak tersedia")
	}

	// Step 1: Load source profile
	if err := e.loadSourceProfile(cloneOpts, isInteractive); err != nil {
		return err
	}

	// Step 2: Setup target profile with pre-fill from source
	e.setupTargetProfileFromSource(cloneOpts)

	// Step 3: Interactive wizard atau validasi non-interactive
	skipWizard := false
	for {
		if !skipWizard && isInteractive {
			// Mode interaktif: wizard dengan pre-fill dari source
			if err := e.Ops.NewWizard().Run(consts.ProfileModeClone); err != nil {
				return err
			}
		} else if !skipWizard {
			// Non-interactive: validasi params
			e.Log.Info(consts.ProfileLogModeNonInteractiveEnabled)
			e.Log.Info(consts.ProfileLogValidatingParams)

			if err := profilevalidation.ValidateProfileInfo(e.State.ProfileInfo); err != nil {
				e.Log.Errorf(consts.ProfileLogValidationFailedFmt, err)
				return err
			}

			e.Log.Info(consts.ProfileLogValidationSuccess)
			if !runtimecfg.IsQuiet() {
				profiledisplay.DisplayProfileDetails(e.ConfigDir, e.State)
			}
		}
		skipWizard = false

		// Check uniqueness
		if err := e.Ops.CheckConfigurationNameUnique(consts.ProfileModeClone); err != nil {
			return err
		}

		// Save cloned profile
		if err := e.Ops.SaveProfile(consts.ProfileSaveModeCreate); err != nil {
			retry, err2 := e.handleConnectionFailedRetryIfNeeded(err, consts.ProfileMsgRetryClone, consts.ProfileMsgCloneCancelled)
			if err2 != nil {
				return err2
			}
			if retry {
				// UX: setelah retry, tampilkan selector field
				if e.isInteractiveMode() {
					if err := e.Ops.NewWizard().PromptCreateRetrySelectedFields(); err != nil {
						return err
					}
					skipWizard = true
				}
				continue
			}
			return validation.ErrUserCancelled
		}
		break
	}

	return nil
}

func (e *Executor) loadSourceProfile(cloneOpts *profilemodel.ProfileCloneOptions, isInteractive bool) error {
	sourcePath := strings.TrimSpace(cloneOpts.SourceProfile)

	// Jika sourcePath kosong, gunakan interactive selection (baik interactive maupun non-interactive)
	if sourcePath == "" {
		if !isInteractive {
			return fmt.Errorf("source profile tidak boleh kosong (--source)")
		}
		// Interactive: prompt untuk select source profile
		if e.Ops == nil {
			return fmt.Errorf("prompt selector tidak tersedia")
		}
		if err := e.Ops.PromptSelectExistingConfig(); err != nil {
			return err
		}
		selected := e.State.ProfileInfo
		if selected == nil {
			return fmt.Errorf("source profile tidak dipilih")
		}
		sourcePath = selected.Path
		if sourcePath == "" {
			return fmt.Errorf("source profile tidak dipilih")
		}

		// Reuse profile yang sudah berhasil didekripsi (mencegah prompt kunci 2x).
		srcCopy := *selected
		cloneOpts.SourceInfo = &srcCopy

		if !runtimecfg.IsQuiet() {
			print.PrintSuccess(fmt.Sprintf("✓ Source profile dimuat: %s", srcCopy.Name))
		}

		// Reset state untuk target clone (agar wizard clone tidak memakai state source langsung).
		e.State.ProfileInfo = &domain.ProfileInfo{}
		return nil
	}

	// Resolve path
	absPath, _, err := e.resolveProfilePath(sourcePath)
	if err != nil {
		return fmt.Errorf("gagal resolve path source profile: %w", err)
	}

	// Load snapshot
	sourceInfo, err := e.Ops.LoadSnapshotFromPath(absPath)
	if err != nil {
		return fmt.Errorf("gagal memuat source profile: %w", err)
	}

	// Set encryption key if provided
	if cloneOpts.ProfileKey != "" {
		sourceInfo.EncryptionKey = cloneOpts.ProfileKey
	}

	cloneOpts.SourceInfo = sourceInfo

	if !runtimecfg.IsQuiet() {
		print.PrintSuccess(fmt.Sprintf("✓ Source profile dimuat: %s", sourceInfo.Name))
	}

	return nil
}

func (e *Executor) setupTargetProfileFromSource(cloneOpts *profilemodel.ProfileCloneOptions) {
	sourceInfo := cloneOpts.SourceInfo
	if sourceInfo == nil {
		return
	}

	// IMPORTANT:
	// Clone bukan edit. Kalau OriginalProfileInfo masih terisi dari proses selection/load,
	// display layer akan menganggap ini mode "show" dan menampilkan profil source (bukan target).
	// Maka, kosongkan baseline original supaya UI selalu menampilkan target clone yang sedang disusun.
	e.State.OriginalProfileInfo = nil
	e.State.OriginalProfileName = ""

	// Clone all fields dari source
	e.State.ProfileInfo.DBInfo = domain.DBInfo{
		Host:     sourceInfo.DBInfo.Host,
		Port:     sourceInfo.DBInfo.Port,
		User:     sourceInfo.DBInfo.User,
		Password: sourceInfo.DBInfo.Password,
	}

	e.State.ProfileInfo.SSHTunnel = domain.SSHTunnelConfig{
		Enabled:      sourceInfo.SSHTunnel.Enabled,
		Host:         sourceInfo.SSHTunnel.Host,
		Port:         sourceInfo.SSHTunnel.Port,
		User:         sourceInfo.SSHTunnel.User,
		Password:     sourceInfo.SSHTunnel.Password,
		IdentityFile: sourceInfo.SSHTunnel.IdentityFile,
		LocalPort:    sourceInfo.SSHTunnel.LocalPort,
	}

	// Apply overrides dari flags
	if cloneOpts.TargetName != "" {
		e.State.ProfileInfo.Name = cloneOpts.TargetName
	} else {
		// Default: tambahkan suffix -clone
		e.State.ProfileInfo.Name = sourceInfo.Name + "-clone"
	}

	if cloneOpts.TargetHost != "" {
		e.State.ProfileInfo.DBInfo.Host = cloneOpts.TargetHost
	}

	if cloneOpts.TargetPort > 0 {
		e.State.ProfileInfo.DBInfo.Port = cloneOpts.TargetPort
	}

	// Encryption key
	if cloneOpts.NewProfileKey != "" {
		e.State.ProfileInfo.EncryptionKey = cloneOpts.NewProfileKey
	} else if cloneOpts.ProfileKey != "" {
		e.State.ProfileInfo.EncryptionKey = cloneOpts.ProfileKey
	} else {
		e.State.ProfileInfo.EncryptionKey = sourceInfo.EncryptionKey
	}
}
