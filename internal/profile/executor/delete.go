// File : internal/profile/executor/delete.go
// Deskripsi : Eksekusi hapus profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package executor

import (
	"fmt"
	"path/filepath"

	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (e *Executor) PromptDeleteProfile() error {
	ui.Headers(consts.ProfileUIHeaderDelete)
	isInteractive := e.isInteractiveMode()

	force := false
	if e.ProfileDelete != nil {
		force = e.ProfileDelete.Force
	}

	if !isInteractive {
		if e.ProfileDelete == nil || len(e.ProfileDelete.Profiles) == 0 {
			return fmt.Errorf(
				"%sflag --profile wajib disertakan: %w",
				consts.ProfileMsgNonInteractivePrefix,
				validation.ErrNonInteractive,
			)
		}
		if !force {
			return fmt.Errorf(
				"%sflag --force wajib disertakan: %w",
				consts.ProfileMsgNonInteractivePrefix,
				validation.ErrNonInteractive,
			)
		}
	}

	// 1) Jika profile path disediakan via flag --profile (support multiple)
	if e.ProfileDelete != nil && len(e.ProfileDelete.Profiles) > 0 {
		validPaths := make([]string, 0, len(e.ProfileDelete.Profiles))
		displayNames := make([]string, 0, len(e.ProfileDelete.Profiles))

		for _, p := range e.ProfileDelete.Profiles {
			if p == "" {
				continue
			}

			absPath, name, err := helper.ResolveConfigPath(p)
			if err != nil {
				return err
			}
			if !fsops.PathExists(absPath) {
				return fmt.Errorf("%s (name: %s)", fmt.Sprintf(consts.ProfileErrConfigFileNotFoundFmt, absPath), name)
			}

			validPaths = append(validPaths, absPath)
			displayNames = append(displayNames, fmt.Sprintf("%s (%s)", name, absPath))
		}

		if len(validPaths) == 0 {
			ui.PrintInfo(consts.ProfileDeleteNoValidProfiles)
			return nil
		}

		if e.ProfileDelete.Force {
			for _, path := range validPaths {
				if err := fsops.RemoveFile(path); err != nil {
					if e.Log != nil {
						e.Log.Errorf(consts.ProfileLogDeleteFileFailedFmt, path, err)
					}
					continue
				}
				if e.Log != nil {
					e.Log.Info(fmt.Sprintf(consts.ProfileDeleteForceDeletedFmt, path))
				}
				ui.PrintSuccess(fmt.Sprintf(consts.ProfileDeleteDeletedFmt, path))
			}
			return nil
		}

		ui.PrintWarning(consts.ProfileDeleteWillDeleteHeader)
		for _, d := range displayNames {
			ui.PrintWarning(consts.ProfileDeleteListPrefix + d)
		}

		ok, err := input.AskYesNo(fmt.Sprintf(consts.ProfileDeleteConfirmCountFmt, len(validPaths)), false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if !ok {
			ui.PrintInfo(consts.ProfileDeleteCancelledByUser)
			return nil
		}

		for _, path := range validPaths {
			if err := fsops.RemoveFile(path); err != nil {
				if e.Log != nil {
					e.Log.Errorf(consts.ProfileLogDeleteFileFailedFmt, path, err)
				}
				ui.PrintError(fmt.Sprintf(consts.ProfileDeleteFailedFmt, path, err))
				continue
			}
			if e.Log != nil {
				e.Log.Info(fmt.Sprintf(consts.ProfileDeleteDeletedFmt, path))
			}
			ui.PrintSuccess(fmt.Sprintf(consts.ProfileDeleteDeletedFmt, path))
		}
		return nil
	}

	// 2) Interactive selection
	files, err := fsops.ReadDirFiles(e.ConfigDir)
	if err != nil {
		return fmt.Errorf(consts.ProfileDeleteReadConfigDirFailedFmt, err)
	}

	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f {
			filtered = append(filtered, f)
		}
	}
	if len(filtered) == 0 {
		ui.PrintInfo(consts.ProfileDeleteNoConfigFiles)
		return nil
	}

	idxs, err := input.ShowMultiSelect(consts.ProfileDeleteSelectFilesPrompt, filtered)
	if err != nil {
		return validation.HandleInputError(err)
	}

	selected := make([]string, 0, len(idxs))
	for _, i := range idxs {
		if i >= 1 && i <= len(filtered) {
			selected = append(selected, filepath.Join(e.ConfigDir, filtered[i-1]))
		}
	}

	if len(selected) == 0 {
		ui.PrintInfo(consts.ProfileDeleteNoFilesSelected)
		return nil
	}

	if !force {
		ok, err := input.AskYesNo(fmt.Sprintf(consts.ProfileDeleteConfirmFilesCountFmt, len(selected)), false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if !ok {
			ui.PrintInfo(consts.ProfileDeleteCancelledByUser)
			return nil
		}
	}

	for _, p := range selected {
		if err := fsops.RemoveFile(p); err != nil {
			if e.Log != nil {
				e.Log.Error(fmt.Sprintf(consts.ProfileDeleteFailedFmt, p, err))
			}
			ui.PrintError(fmt.Sprintf(consts.ProfileDeleteFailedFmt, p, err))
			continue
		}
		if e.Log != nil {
			e.Log.Info(fmt.Sprintf(consts.ProfileDeleteDeletedFmt, p))
		}
		ui.PrintSuccess(fmt.Sprintf(consts.ProfileDeleteDeletedFmt, p))
	}

	return nil
}
