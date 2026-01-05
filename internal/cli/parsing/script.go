package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"strings"

	"github.com/spf13/cobra"
)

func ParsingScriptEncryptOptions(cmd *cobra.Command) types.ScriptEncryptOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	deleteSource, _ := cmd.Flags().GetBool("delete-source")
	return types.ScriptEncryptOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		Mode:          helper.GetStringFlagOrEnv(cmd, "mode", ""),
		OutputPath:    helper.GetStringFlagOrEnv(cmd, "output", ""),
		DeleteSource:  deleteSource,
	}
}

func ParsingScriptRunOptions(cmd *cobra.Command) types.ScriptRunOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return types.ScriptRunOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}

func ParsingScriptExtractOptions(cmd *cobra.Command) types.ScriptExtractOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return types.ScriptExtractOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		OutDir:        helper.GetStringFlagOrEnv(cmd, "out-dir", ""),
	}
}

func ParsingScriptInfoOptions(cmd *cobra.Command) types.ScriptInfoOptions {
	key := helper.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return types.ScriptInfoOptions{
		FilePath:      helper.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}
