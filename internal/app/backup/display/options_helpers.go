package display

import (
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/text"
)

func (d *OptionsDisplayer) isSingleMode() bool {
	return d.options.Mode == consts.ModeSingle || d.options.Mode == consts.ModePrimary || d.options.Mode == consts.ModeSecondary
}

func (d *OptionsDisplayer) isSeparatedMode() bool {
	return d.options.Mode == consts.ModeSeparated
}

func (d *OptionsDisplayer) getExportUserStatus() string {
	if d.options.ExcludeUser {
		return "No"
	}
	return text.ColorText("Yes", consts.UIColorGreen)
}

func (d *OptionsDisplayer) getFilenameLabel() string {
	switch d.options.Mode {
	case consts.ModeAll:
		return "Output File (All)"
	case consts.ModeCombined:
		return "Output File (Filter)"
	default:
		return "Backup Filename"
	}
}
