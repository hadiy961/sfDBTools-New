package compress

// --- Compression Ratio Data ---

// CompressionRatios menyimpan data rasio kompresi berdasarkan pengalaman empiris.
// Dipisahkan dari logika untuk kemudahan pembacaan dan pemeliharaan.
var CompressionRatios = map[CompressionType]map[CompressionLevel]float64{
	CompressionNone: {
		LevelDefault: 1.0, // Tanpa kompresi, ukuran tetap 100%
	},
	CompressionGzip: {
		// Pilihan seimbang yang umum, tetapi telah dilampaui oleh zstd [4]
		LevelBestSpeed: 0.30, // Level -1
		LevelFast:      0.28, // Level -3
		LevelDefault:   0.25, // Level -6 [5]
		LevelBetter:    0.245,
		LevelBest:      0.243, // Level -9, pengembalian yang semakin berkurang [5]
	},
	CompressionBzip2: {
		// Rasio lebih baik dari gzip, tetapi sangat lambat untuk kompresi dan dekompresi [6, 5]
		LevelBestSpeed: 0.23, // Level -1
		LevelFast:      0.21,
		LevelDefault:   0.20, // Level -9 [5]
		LevelBetter:    0.198,
		LevelBest:      0.197,
	},
	CompressionLz4: {
		// Sangat cepat, tetapi rasio kompresi paling rendah, baik untuk kecepatan maksimal [1, 7, 4]
		LevelBestSpeed: 0.40,
		LevelFast:      0.38,
		LevelDefault:   0.36,
		LevelBetter:    0.35,
		LevelBest:      0.34,
	},
	CompressionXz: {
		// Rasio kompresi terbaik, tetapi sangat lambat dan intensif CPU saat kompresi [7, 4, 5]
		LevelBestSpeed: 0.24, // Level -0
		LevelFast:      0.18, // Level -3
		LevelDefault:   0.16, // Level -6 [5]
		LevelBetter:    0.159,
		LevelBest:      0.157, // Level -9
	},
	CompressionZlib: {
		LevelBestSpeed: 0.21,
		LevelFast:      0.19,
		LevelDefault:   0.16,
		LevelBetter:    0.14,
		LevelBest:      0.13,
	},
	CompressionZstd: {
		// Keseimbangan terbaik antara kecepatan dan rasio kompresi untuk sebagian besar kasus penggunaan [6, 4]
		LevelBestSpeed: 0.28, // Level 1
		LevelFast:      0.25,
		LevelDefault:   0.22, // Level 3-5
		LevelBetter:    0.20,
		LevelBest:      0.18, // Level 19+
	},
}
