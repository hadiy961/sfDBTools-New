package servicehelper

// File : pkg/servicehelper/servicehelper_base.go
// Deskripsi : Base helper untuk service operations dengan graceful shutdown support
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025

import (
	"context"
	"sync"
)

// BaseService menyediakan functionality umum untuk service dengan graceful shutdown
// Embed struct ini ke dalam service struct untuk mendapatkan functionality-nya
type BaseService struct {
	cancelFunc context.CancelFunc
	mu         sync.Mutex
}

// SetCancelFunc menyimpan cancel function untuk graceful shutdown
func (b *BaseService) SetCancelFunc(cancel context.CancelFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cancelFunc = cancel
}

// GetCancelFunc mengambil cancel function (dengan lock protection)
func (b *BaseService) GetCancelFunc() context.CancelFunc {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cancelFunc
}

// Cancel memanggil cancel function jika ada
func (b *BaseService) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
}

// WithLock menjalankan function dengan mutex lock protection
// Berguna untuk operasi custom yang butuh thread-safety
func (b *BaseService) WithLock(fn func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fn()
}
