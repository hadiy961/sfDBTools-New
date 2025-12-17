package parsing

import (
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingBase64EncodeOptions membaca flag base64 encode
func ParsingBase64EncodeOptions(cmd *cobra.Command) types.Base64EncodeOptions {
	return types.Base64EncodeOptions{
		InputText:  helper.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingBase64DecodeOptions membaca flag base64 decode
func ParsingBase64DecodeOptions(cmd *cobra.Command) types.Base64DecodeOptions {
	return types.Base64DecodeOptions{
		InputData:  helper.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingEncryptFileOptions membaca flag encrypt file
func ParsingEncryptFileOptions(cmd *cobra.Command) types.EncryptFileOptions {
	return types.EncryptFileOptions{
		InputPath:  helper.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptFileOptions membaca flag decrypt file
func ParsingDecryptFileOptions(cmd *cobra.Command) types.DecryptFileOptions {
	return types.DecryptFileOptions{
		InputPath:  helper.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingEncryptTextOptions membaca flag encrypt text
func ParsingEncryptTextOptions(cmd *cobra.Command) types.EncryptTextOptions {
	return types.EncryptTextOptions{
		InputText:  helper.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptTextOptions membaca flag decrypt text
func ParsingDecryptTextOptions(cmd *cobra.Command) types.DecryptTextOptions {
	return types.DecryptTextOptions{
		InputData:  helper.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}
