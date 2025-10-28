package types

import "errors"

// ErrUserCancelled adalah sentinel error untuk menandai pembatalan oleh pengguna.
var ErrUserCancelled = errors.New("user_cancelled")

// ErrConnectionFailedRetry adalah sentinel error untuk menandai kegagalan koneksi dengan permintaan retry.
var ErrConnectionFailedRetry = errors.New("connection_failed_retry")
