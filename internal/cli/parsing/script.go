package parsing

import (
	scriptmodel "sfDBTools/internal/app/script/model"
	"sfDBTools/pkg/helper"
	"strings"

	"github.com/spf13/cobra"
)

func ParsingScriptEncryptOptions(cmd *cobra.Command) scriptmodel.ScriptEncryptOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	deleteSource, _ := cmd.Flags().GetBool("delete-source")
	return scriptmodel.ScriptEncryptOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		Mode:          helper.GetStringFlagOrEnv(cmd, "mode", ""),
		OutputPath:    helper.GetStringFlagOrEnv(cmd, "output", ""),
		DeleteSource:  deleteSource,
	}
}

func ParsingScriptRunOptions(cmd *cobra.Command) scriptmodel.ScriptRunOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return scriptmodel.ScriptRunOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}

func ParsingScriptExtractOptions(cmd *cobra.Command) scriptmodel.ScriptExtractOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return scriptmodel.ScriptExtractOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		OutDir:        helper.GetStringFlagOrEnv(cmd, "out-dir", ""),
	}
}

func ParsingScriptInfoOptions(cmd *cobra.Command) scriptmodel.ScriptInfoOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return scriptmodel.ScriptInfoOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}
