package parsing

import (
	resolver "sfdbtools/internal/cli/resolver"
	cryptomodel "sfdbtools/internal/services/crypto/model"

	"github.com/spf13/cobra"
)

// ParsingBase64EncodeOptions membaca flag base64 encode
func ParsingBase64EncodeOptions(cmd *cobra.Command) cryptomodel.Base64EncodeOptions {
	return cryptomodel.Base64EncodeOptions{
		InputText:  resolver.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingBase64DecodeOptions membaca flag base64 decode
func ParsingBase64DecodeOptions(cmd *cobra.Command) cryptomodel.Base64DecodeOptions {
	return cryptomodel.Base64DecodeOptions{
		InputData:  resolver.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingEncryptFileOptions membaca flag encrypt file
func ParsingEncryptFileOptions(cmd *cobra.Command) cryptomodel.EncryptFileOptions {
	return cryptomodel.EncryptFileOptions{
		InputPath:  resolver.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptFileOptions membaca flag decrypt file
func ParsingDecryptFileOptions(cmd *cobra.Command) cryptomodel.DecryptFileOptions {
	return cryptomodel.DecryptFileOptions{
		InputPath:  resolver.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingEncryptTextOptions membaca flag encrypt text
func ParsingEncryptTextOptions(cmd *cobra.Command) cryptomodel.EncryptTextOptions {
	return cryptomodel.EncryptTextOptions{
		InputText:  resolver.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptTextOptions membaca flag decrypt text
func ParsingDecryptTextOptions(cmd *cobra.Command) cryptomodel.DecryptTextOptions {
	return cryptomodel.DecryptTextOptions{
		InputData:  resolver.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: resolver.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        resolver.GetStringFlagOrEnv(cmd, "key", ""),
	}
}
