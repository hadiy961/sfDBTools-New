package consts

// Compression type identifiers.
const (
	CompressionTypeNone  = "none"
	CompressionTypeGzip  = "gzip"
	CompressionTypePgzip = "pgzip"
	CompressionTypeZlib  = "zlib"
	CompressionTypeZstd  = "zstd"
	CompressionTypeXz    = "xz"
)

// Compression level values (1-9).
const (
	CompressionLevelBestSpeed = 1
	CompressionLevelFast      = 3
	CompressionLevelDefault   = 6
	CompressionLevelBetter    = 7
	CompressionLevelBest      = 9
)
