package parsing

import (
	scriptmodel "sfdbtools/internal/app/script/model"
	resolver "sfdbtools/internal/cli/resolver"
	"strings"

	"github.com/spf13/cobra"
)

func ParsingScriptEncryptOptions(cmd *cobra.Command) scriptmodel.ScriptEncryptOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	deleteSource, _ := cmd.Flags().GetBool("delete-source")
	return scriptmodel.ScriptEncryptOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		Mode:          resolver.GetStringFlagOrEnv(cmd, "mode", ""),
		OutputPath:    resolver.GetStringFlagOrEnv(cmd, "output", ""),
		DeleteSource:  deleteSource,
	}
}

func ParsingScriptRunOptions(cmd *cobra.Command) scriptmodel.ScriptRunOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return scriptmodel.ScriptRunOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}

func ParsingScriptExtractOptions(cmd *cobra.Command) scriptmodel.ScriptExtractOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return scriptmodel.ScriptExtractOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		OutDir:        resolver.GetStringFlagOrEnv(cmd, "out-dir", ""),
	}
}

func ParsingScriptInfoOptions(cmd *cobra.Command) scriptmodel.ScriptInfoOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return scriptmodel.ScriptInfoOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}
