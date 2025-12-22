package compress

import "sfDBTools/pkg/consts"

// --- Compression Ratio Data ---

// CompressionRatios menyimpan data rasio kompresi berdasarkan pengalaman empiris.
// Dipisahkan dari logika untuk kemudahan pembacaan dan pemeliharaan.
var CompressionRatios = map[CompressionType]map[CompressionLevel]float64{
	CompressionType(consts.CompressionTypeNone): {
		CompressionLevel(consts.CompressionLevelDefault): 1.0, // Tanpa kompresi, ukuran tetap 100%
	},
	CompressionType(consts.CompressionTypeGzip): {
		// Standard gzip - pilihan seimbang yang umum [4]
		CompressionLevel(consts.CompressionLevelBestSpeed): 0.30, // Level 1
		CompressionLevel(consts.CompressionLevelFast):      0.28, // Level 3
		CompressionLevel(consts.CompressionLevelDefault):   0.25, // Level 6 [5]
		CompressionLevel(consts.CompressionLevelBetter):    0.245,
		CompressionLevel(consts.CompressionLevelBest):      0.243, // Level 9, pengembalian yang semakin berkurang [5]
	},
	CompressionType(consts.CompressionTypePgzip): {
		// Parallel gzip - performa sama dengan gzip, tapi lebih cepat untuk file besar
		CompressionLevel(consts.CompressionLevelBestSpeed): 0.30,
		CompressionLevel(consts.CompressionLevelFast):      0.28,
		CompressionLevel(consts.CompressionLevelDefault):   0.25,
		CompressionLevel(consts.CompressionLevelBetter):    0.245,
		CompressionLevel(consts.CompressionLevelBest):      0.243,
	},
	CompressionType(consts.CompressionTypeZlib): {
		// Zlib/DEFLATE - sedikit lebih baik dari gzip
		CompressionLevel(consts.CompressionLevelBestSpeed): 0.29,
		CompressionLevel(consts.CompressionLevelFast):      0.27,
		CompressionLevel(consts.CompressionLevelDefault):   0.24,
		CompressionLevel(consts.CompressionLevelBetter):    0.235,
		CompressionLevel(consts.CompressionLevelBest):      0.23,
	},
	CompressionType(consts.CompressionTypeZstd): {
		// Zstandard - keseimbangan terbaik antara kecepatan dan rasio kompresi [6, 4]
		CompressionLevel(consts.CompressionLevelBestSpeed): 0.28, // Level 1
		CompressionLevel(consts.CompressionLevelFast):      0.25, // Level 3
		CompressionLevel(consts.CompressionLevelDefault):   0.22, // Level 5-6
		CompressionLevel(consts.CompressionLevelBetter):    0.20, // Level 7-9
		CompressionLevel(consts.CompressionLevelBest):      0.18, // Level 19+
	},
	CompressionType(consts.CompressionTypeXz): {
		// XZ/LZMA - rasio kompresi terbaik, tetapi sangat lambat dan intensif CPU [7, 4, 5]
		CompressionLevel(consts.CompressionLevelBestSpeed): 0.24, // Level 0
		CompressionLevel(consts.CompressionLevelFast):      0.18, // Level 3
		CompressionLevel(consts.CompressionLevelDefault):   0.16, // Level 6 [5]
		CompressionLevel(consts.CompressionLevelBetter):    0.159,
		CompressionLevel(consts.CompressionLevelBest):      0.157, // Level 9
	},
}
