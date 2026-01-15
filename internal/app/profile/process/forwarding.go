// File : internal/app/profile/process/forwarding.go
// Deskripsi : Port selection dan forwarding loop untuk SSH tunnel
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 15 Januari 2026

package process

import (
	"context"
	"io"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh"
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

func forward(ctx context.Context, ln net.Listener, client *ssh.Client, remoteAddr string) {
	for {
		conn, aerr := ln.Accept()
		if aerr != nil {
			select {
			case <-ctx.Done():
				return
			default:
				return
			}
		}
		go func(c net.Conn) {
			defer c.Close()
			rc, derr := client.Dial("tcp", remoteAddr)
			if derr != nil {
				return
			}
			defer rc.Close()

			// Bidirectional copy
			done := make(chan struct{}, 2)
			go func() { _, _ = io.Copy(rc, c); done <- struct{}{} }()
			go func() { _, _ = io.Copy(c, rc); done <- struct{}{} }()
			<-done
		}(conn)
	}
}
