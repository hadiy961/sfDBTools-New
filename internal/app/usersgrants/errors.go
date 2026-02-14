package usersgrants

import "errors"

// ErrNoUserWithGrants menandakan tidak ada user yang memiliki grants relevan
// terhadap database filter yang diminta.
// Caller dapat menggunakan errors.Is(err, ErrNoUserWithGrants) untuk menangani
// kasus ini sebagai non-fatal (mis. skip export grants).
var ErrNoUserWithGrants = errors.New("tidak ada user dengan grants")
