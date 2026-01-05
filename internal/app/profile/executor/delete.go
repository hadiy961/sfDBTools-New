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

func filterProfileConfigFiles(files []string) []string {
	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func (e *Executor) collectValidPathsFromFlags(profiles []string) (validPaths []string, displayNames []string, err error) {
	validPaths = make([]string, 0, len(profiles))
	displayNames = make([]string, 0, len(profiles))

	for _, p := range profiles {
		if p == "" {
			continue
		}

		absPath, name, err := helper.ResolveConfigPath(p)
		if err != nil {
			return nil, nil, err
		}
		if !fsops.PathExists(absPath) {
			return nil, nil, fmt.Errorf("%s (name: %s)", fmt.Sprintf(consts.ProfileErrConfigFileNotFoundFmt, absPath), name)
		}

		validPaths = append(validPaths, absPath)
		displayNames = append(displayNames, fmt.Sprintf("%s (%s)", name, absPath))
	}

	return validPaths, displayNames, nil
}

func (e *Executor) deletePaths(paths []string, logSuccessFmt string, showErrorOnFail bool, logFailAsError bool) {
	for _, p := range paths {
		if err := fsops.RemoveFile(p); err != nil {
			if e.Log != nil {
				if logFailAsError {
					e.Log.Error(fmt.Sprintf(consts.ProfileDeleteFailedFmt, p, err))
				} else {
					e.Log.Errorf(consts.ProfileLogDeleteFileFailedFmt, p, err)
				}
			}
			if showErrorOnFail {
				ui.PrintError(fmt.Sprintf(consts.ProfileDeleteFailedFmt, p, err))
			}
			continue
		}

		if e.Log != nil {
			e.Log.Info(fmt.Sprintf(logSuccessFmt, p))
		}
		// UI success selalu pakai message yang sama seperti sebelumnya.
		ui.PrintSuccess(fmt.Sprintf(consts.ProfileDeleteDeletedFmt, p))
	}
}

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
		validPaths, displayNames, err := e.collectValidPathsFromFlags(e.ProfileDelete.Profiles)
		if err != nil {
			return err
		}

		if len(validPaths) == 0 {
			ui.PrintInfo(consts.ProfileDeleteNoValidProfiles)
			return nil
		}

		if e.ProfileDelete.Force {
			e.deletePaths(validPaths, consts.ProfileDeleteForceDeletedFmt, false, false)
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

		e.deletePaths(validPaths, consts.ProfileDeleteDeletedFmt, true, false)
		return nil
	}

	// 2) Interactive selection
	files, err := fsops.ReadDirFiles(e.ConfigDir)
	if err != nil {
		return fmt.Errorf(consts.ProfileDeleteReadConfigDirFailedFmt, err)
	}

	filtered := filterProfileConfigFiles(files)
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

	e.deletePaths(selected, consts.ProfileDeleteDeletedFmt, true, true)

	return nil
}
