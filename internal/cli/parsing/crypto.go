package parsing

import (
	cryptomodel "sfdbtools/internal/services/crypto/model"
	"sfdbtools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingBase64EncodeOptions membaca flag base64 encode
func ParsingBase64EncodeOptions(cmd *cobra.Command) cryptomodel.Base64EncodeOptions {
	return cryptomodel.Base64EncodeOptions{
		InputText:  helper.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingBase64DecodeOptions membaca flag base64 decode
func ParsingBase64DecodeOptions(cmd *cobra.Command) cryptomodel.Base64DecodeOptions {
	return cryptomodel.Base64DecodeOptions{
		InputData:  helper.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
	}
}

// ParsingEncryptFileOptions membaca flag encrypt file
func ParsingEncryptFileOptions(cmd *cobra.Command) cryptomodel.EncryptFileOptions {
	return cryptomodel.EncryptFileOptions{
		InputPath:  helper.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptFileOptions membaca flag decrypt file
func ParsingDecryptFileOptions(cmd *cobra.Command) cryptomodel.DecryptFileOptions {
	return cryptomodel.DecryptFileOptions{
		InputPath:  helper.GetStringFlagOrEnv(cmd, "in", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingEncryptTextOptions membaca flag encrypt text
func ParsingEncryptTextOptions(cmd *cobra.Command) cryptomodel.EncryptTextOptions {
	return cryptomodel.EncryptTextOptions{
		InputText:  helper.GetStringFlagOrEnv(cmd, "text", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}

// ParsingDecryptTextOptions membaca flag decrypt text
func ParsingDecryptTextOptions(cmd *cobra.Command) cryptomodel.DecryptTextOptions {
	return cryptomodel.DecryptTextOptions{
		InputData:  helper.GetStringFlagOrEnv(cmd, "data", ""),
		OutputPath: helper.GetStringFlagOrEnv(cmd, "out", ""),
		Key:        helper.GetStringFlagOrEnv(cmd, "key", ""),
	}
}
