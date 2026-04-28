package verify

import (
	"fmt"
	"strings"
)

// ContentStep menangani Checksum dan Header/Footer validation,
// berpotensi menggunakan Single-Pass Read jika keduanya aktif.
type ContentStep struct{}

// Name returns the name of the step
func (s *ContentStep) Name() string {
	return "Content Validation (Checksum, Header/Footer)"
}

// Execute menjalankan checksum dan/atau header footer validation
func (s *ContentStep) Execute(ctx *VerifyContext) error {
	// Jika keduanya tidak diminta, skip
	if !ctx.Opts.Checksum && !ctx.Opts.HeaderFooter {
		return nil
	}

	// 1. Single-Pass Read jika Checksum dan HeaderFooter keduanya aktif
	if ctx.Opts.Checksum && ctx.Opts.HeaderFooter {
		return s.executeSinglePass(ctx)
	}

	// 2. Jika hanya Checksum
	if ctx.Opts.Checksum {
		return s.executeChecksumOnly(ctx)
	}

	// 3. Jika hanya Header/Footer
	if ctx.Opts.HeaderFooter {
		return s.executeHeaderFooterOnly(ctx)
	}

	return nil
}

func (s *ContentStep) executeChecksumOnly(ctx *VerifyContext) error {
	algo := ctx.Opts.ChecksumAlgo
	if algo == "" {
		algo = "sha256"
	}
	ctx.Result.ChecksumAlgo = algo

	if ctx.Opts.ExpectedHash != "" {
		match, actual, err := CompareChecksum(ctx.FilePath, algo, ctx.Opts.ExpectedHash)
		if err != nil {
			return fmt.Errorf("failed to compare checksum: %w", err)
		}
		ctx.Result.ChecksumHash = actual
		if !match {
			ctx.Result.VerifyStatus = "failed"
			ctx.Result.FailureReason = fmt.Sprintf("Checksum mismatch: expected %s, got %s", ctx.Opts.ExpectedHash, actual)
			return nil
		}
	} else {
		hash, err := GenerateChecksum(ctx.FilePath, algo)
		if err != nil {
			return fmt.Errorf("failed to generate checksum: %w", err)
		}
		ctx.Result.ChecksumHash = hash
	}
	return nil
}

func (s *ContentStep) executeHeaderFooterOnly(ctx *VerifyContext) error {
	reader, closers, err := OpenVerifyReader(ctx.FilePath, ctx.Opts.EncryptionKey, nil)
	if err != nil {
		return fmt.Errorf("failed to open verify reader: %w", err)
	}
	defer CloseReaders(closers)

	headerOK, footerOK, err := ValidateHeaderFooter(reader)
	if err != nil {
		return fmt.Errorf("error during header/footer validation: %w", err)
	}

	s.populateHeaderFooterResult(ctx, headerOK, footerOK)
	return nil
}

func (s *ContentStep) executeSinglePass(ctx *VerifyContext) error {
	algo := ctx.Opts.ChecksumAlgo
	if algo == "" {
		algo = "sha256"
	}
	ctx.Result.ChecksumAlgo = algo

	hasher, err := GetHasher(algo)
	if err != nil {
		return fmt.Errorf("failed to get hasher: %w", err)
	}

	// Buka verify reader dengan menyuntikkan hasher untuk di-Tee
	reader, closers, err := OpenVerifyReader(ctx.FilePath, ctx.Opts.EncryptionKey, hasher)
	if err != nil {
		return fmt.Errorf("failed to open verify reader: %w", err)
	}
	defer CloseReaders(closers)

	headerOK, footerOK, err := ValidateHeaderFooter(reader)
	if err != nil {
		return fmt.Errorf("error during header/footer validation: %w", err)
	}

	// Setelah reader selesai membaca stream, hasher akan memiliki final hash
	actualHash := HashToString(hasher)
	ctx.Result.ChecksumHash = actualHash

	if ctx.Opts.ExpectedHash != "" && actualHash != ctx.Opts.ExpectedHash {
		ctx.Result.VerifyStatus = "failed"
		ctx.Result.FailureReason = fmt.Sprintf("Checksum mismatch: expected %s, got %s", ctx.Opts.ExpectedHash, actualHash)
		return nil
	}

	s.populateHeaderFooterResult(ctx, headerOK, footerOK)
	return nil
}

func (s *ContentStep) populateHeaderFooterResult(ctx *VerifyContext, headerOK, footerOK bool) {
	ctx.Result.HeaderValid = &headerOK
	ctx.Result.FooterValid = &footerOK

	if !headerOK || !footerOK {
		ctx.Result.VerifyStatus = "failed"
		var reasons []string
		if !headerOK {
			reasons = append(reasons, "Invalid SQL Header")
		}
		if !footerOK {
			reasons = append(reasons, "Invalid SQL Footer (Dump not completed)")
		}
		
		if ctx.Result.FailureReason != "" {
			ctx.Result.FailureReason += " | " + strings.Join(reasons, ", ")
		} else {
			ctx.Result.FailureReason = strings.Join(reasons, ", ")
		}
	}
}
