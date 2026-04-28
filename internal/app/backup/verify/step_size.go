package verify

import (
	"fmt"
)

// SizeStep memverifikasi ukuran file minimum
type SizeStep struct{}

// Name returns the name of the step
func (s *SizeStep) Name() string {
	return "Size Validation"
}

// Execute menjalankan size validation
func (s *SizeStep) Execute(ctx *VerifyContext) error {
	if ctx.Opts.SizeCheck {
		valid, size, err := ValidateSize(ctx.FilePath, ctx.Opts.MinFileSize)
		if err != nil {
			return fmt.Errorf("failed to get file size: %w", err)
		}
		ctx.Result.FileSizeBytes = size
		sizeValid := valid
		ctx.Result.SizeValid = &sizeValid

		if !valid {
			ctx.Result.VerifyStatus = "failed"
			ctx.Result.FailureReason = fmt.Errorf("file size %d is below minimum %d", size, ctx.Opts.MinFileSize).Error()
			return nil // Soft error, let the pipeline know it failed
		}
	} else {
		_, size, err := ValidateSize(ctx.FilePath, 0)
		if err == nil {
			ctx.Result.FileSizeBytes = size
		}
	}

	if ctx.Result.FileSizeBytes == 0 {
		ctx.Result.VerifyStatus = "failed"
		ctx.Result.FailureReason = "File is empty"
		return nil
	}

	return nil
}
