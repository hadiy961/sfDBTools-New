package parsing

import (
	"sfdbtools/internal/app/script"
	resolver "sfdbtools/internal/cli/resolver"
	"strings"

	"github.com/spf13/cobra"
)

func ParsingScriptEncryptOptions(cmd *cobra.Command) script.EncryptOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	deleteSource, _ := cmd.Flags().GetBool("delete-source")
	return script.EncryptOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		Mode:          resolver.GetStringFlagOrEnv(cmd, "mode", ""),
		OutputPath:    resolver.GetStringFlagOrEnv(cmd, "output", ""),
		DeleteSource:  deleteSource,
	}
}

func ParsingScriptRunOptions(cmd *cobra.Command) script.RunOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return script.RunOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}

func ParsingScriptExtractOptions(cmd *cobra.Command) script.ExtractOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return script.ExtractOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
		OutDir:        resolver.GetStringFlagOrEnv(cmd, "out-dir", ""),
	}
}

func ParsingScriptInfoOptions(cmd *cobra.Command) script.InfoOptions {
	key := resolver.GetStringFlagOrEnv(cmd, "key", "")
	if strings.TrimSpace(key) == "" {
		key = resolver.GetStringFlagOrEnv(cmd, "encryption-key", "")
	}
	return script.InfoOptions{
		FilePath:      resolver.GetStringFlagOrEnv(cmd, "file", ""),
		EncryptionKey: key,
	}
}
