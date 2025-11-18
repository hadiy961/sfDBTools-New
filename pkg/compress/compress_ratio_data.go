package compress

// --- Compression Ratio Data ---

// CompressionRatios menyimpan data rasio kompresi berdasarkan pengalaman empiris.
// Dipisahkan dari logika untuk kemudahan pembacaan dan pemeliharaan.
var CompressionRatios = map[CompressionType]map[CompressionLevel]float64{
	CompressionNone: {
		LevelDefault: 1.0, // Tanpa kompresi, ukuran tetap 100%
	},
	CompressionGzip: {
		// Standard gzip - pilihan seimbang yang umum [4]
		LevelBestSpeed: 0.30, // Level 1
		LevelFast:      0.28, // Level 3
		LevelDefault:   0.25, // Level 6 [5]
		LevelBetter:    0.245,
		LevelBest:      0.243, // Level 9, pengembalian yang semakin berkurang [5]
	},
	CompressionPgzip: {
		// Parallel gzip - performa sama dengan gzip, tapi lebih cepat untuk file besar
		LevelBestSpeed: 0.30,
		LevelFast:      0.28,
		LevelDefault:   0.25,
		LevelBetter:    0.245,
		LevelBest:      0.243,
	},
	CompressionZlib: {
		// Zlib/DEFLATE - sedikit lebih baik dari gzip
		LevelBestSpeed: 0.29,
		LevelFast:      0.27,
		LevelDefault:   0.24,
		LevelBetter:    0.235,
		LevelBest:      0.23,
	},
	CompressionZstd: {
		// Zstandard - keseimbangan terbaik antara kecepatan dan rasio kompresi [6, 4]
		LevelBestSpeed: 0.28, // Level 1
		LevelFast:      0.25, // Level 3
		LevelDefault:   0.22, // Level 5-6
		LevelBetter:    0.20, // Level 7-9
		LevelBest:      0.18, // Level 19+
	},
	CompressionXz: {
		// XZ/LZMA - rasio kompresi terbaik, tetapi sangat lambat dan intensif CPU [7, 4, 5]
		LevelBestSpeed: 0.24, // Level 0
		LevelFast:      0.18, // Level 3
		LevelDefault:   0.16, // Level 6 [5]
		LevelBetter:    0.159,
		LevelBest:      0.157, // Level 9
	},
}
