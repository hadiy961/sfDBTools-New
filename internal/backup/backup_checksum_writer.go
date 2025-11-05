// File : internal/backup/backup_checksum_writer.go
// Deskripsi : Streaming checksum writer untuk menghitung hash saat backup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package backup

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

// MultiHashWriter adalah writer yang menghitung multiple hash secara bersamaan
// saat data ditulis (streaming), memory-efficient untuk file besar
type MultiHashWriter struct {
	writer       io.Writer
	sha256Hash   hash.Hash
	md5Hash      hash.Hash
	bytesWritten int64
}

// NewMultiHashWriter membuat instance baru MultiHashWriter
func NewMultiHashWriter(w io.Writer) *MultiHashWriter {
	return &MultiHashWriter{
		writer:     w,
		sha256Hash: sha256.New(),
		md5Hash:    md5.New(),
	}
}

// Write menulis data ke underlying writer dan update hash
func (m *MultiHashWriter) Write(p []byte) (n int, err error) {
	// Tulis ke underlying writer
	n, err = m.writer.Write(p)
	if err != nil {
		return n, err
	}

	// Update hash calculators
	m.sha256Hash.Write(p[:n])
	m.md5Hash.Write(p[:n])
	m.bytesWritten += int64(n)

	return n, nil
}

// GetSHA256 mengembalikan SHA256 hash dalam format hex
func (m *MultiHashWriter) GetSHA256() string {
	return hex.EncodeToString(m.sha256Hash.Sum(nil))
}

// GetMD5 mengembalikan MD5 hash dalam format hex
func (m *MultiHashWriter) GetMD5() string {
	return hex.EncodeToString(m.md5Hash.Sum(nil))
}

// GetBytesWritten mengembalikan total bytes yang ditulis
func (m *MultiHashWriter) GetBytesWritten() int64 {
	return m.bytesWritten
}

// Close implements io.Closer (if underlying writer is closer)
func (m *MultiHashWriter) Close() error {
	if closer, ok := m.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
