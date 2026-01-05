package helpers

import "context"

// DatabaseDetailJob untuk worker pattern
type DatabaseDetailJob struct {
	DatabaseName string
}

// DetailCollectOptions mengizinkan override perilaku koleksi metrik tertentu.
type DetailCollectOptions struct {
	// SizeProvider, bila diset, digunakan untuk menghitung ukuran database
	// sebagai pengganti query GetDatabaseSize. Kembalikan (size,nil) untuk sukses
	// atau error untuk menandai kegagalan metrik.
	SizeProvider func(ctx context.Context, dbName string) (int64, error)
}
