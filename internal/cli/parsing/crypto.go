package parsing

import (
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/crypto"

	"github.com/spf13/cobra"
)

// ParsingBase64EncodeOptions membaca flag base64 encode
func ParsingBase64EncodeOptions(cmd *cobra.Command) crypto.Base64EncodeOptions {
	return crypto.Base64EncodeOptions{
		InputText:  resolver.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingBase64DecodeOptions membaca flag base64 decode
func ParsingBase64DecodeOptions(cmd *cobra.Command) crypto.Base64DecodeOptions {
	return crypto.Base64DecodeOptions{
		InputData:  resolver.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingEncryptFileOptions membaca flag encrypt file
func ParsingEncryptFileOptions(cmd *cobra.Command) crypto.EncryptFileOptions {
	return crypto.EncryptFileOptions{
		InputPath:  resolver.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptFileOptions membaca flag decrypt file
func ParsingDecryptFileOptions(cmd *cobra.Command) crypto.DecryptFileOptions {
	return crypto.DecryptFileOptions{
		InputPath:  resolver.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingEncryptTextOptions membaca flag encrypt text
func ParsingEncryptTextOptions(cmd *cobra.Command) crypto.EncryptTextOptions {
	return crypto.EncryptTextOptions{
		InputText:  resolver.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptTextOptions membaca flag decrypt text
func ParsingDecryptTextOptions(cmd *cobra.Command) crypto.DecryptTextOptions {
	return crypto.DecryptTextOptions{
		InputData:  resolver.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}
