// File : internal/app/profile/process/ssh_ports.go
// Deskripsi : Helper port untuk SSH tunnel
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package process

import (
	"net"
	"strconv"
)

func pickLocalPort(requested int) (int, error) {
	if requested > 0 {
		return requested, nil
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr := ln.Addr().String()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return p, nil
}
