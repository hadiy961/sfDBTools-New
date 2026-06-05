package connectivity

import (
	"context"
	"net"
	"time"
)

// IsInternetAvailable checks if the internet is accessible by dialing common public endpoints.
func IsInternetAvailable(ctx context.Context, timeout time.Duration) bool {
	addrs := []string{
		"1.1.1.1:443",
		"8.8.8.8:53",
		"api.github.com:443",
		"google.com:443",
	}

	d := net.Dialer{Timeout: timeout}
	for _, addr := range addrs {
		cctx, cancel := context.WithTimeout(ctx, timeout)
		conn, err := d.DialContext(cctx, "tcp", addr)
		cancel()
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}
