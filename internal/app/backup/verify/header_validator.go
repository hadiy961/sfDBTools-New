package verify

import (
	"bytes"
	"fmt"
	"io"
)

// ValidateHeader membaca N bytes pertama dari stream dan cek apakah mengandung
// signature MariaDB/MySQL dump (e.g. "-- MariaDB dump", "-- MySQL dump")
func ValidateHeader(reader io.Reader) (bool, error) {
	buf := make([]byte, 1024)
	n, err := io.ReadFull(reader, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, fmt.Errorf("failed to read header: %w", err)
	}

	content := buf[:n]
	if bytes.Contains(content, []byte("-- MariaDB dump")) || bytes.Contains(content, []byte("-- MySQL dump")) {
		return true, nil
	}

	return false, nil
}

// ValidateFooter membaca stream sampai habis dan cek apakah N bytes terakhir
// mengandung "-- Dump completed" marker
func ValidateFooter(reader io.Reader) (bool, error) {
	const bufSize = 4096
	tailBuf := make([]byte, bufSize)
	tempBuf := make([]byte, bufSize)

	var totalRead int
	for {
		n, err := reader.Read(tempBuf)
		if n > 0 {
			if n >= bufSize {
				copy(tailBuf, tempBuf[n-bufSize:n])
			} else {
				shift := bufSize - n
				copy(tailBuf, tailBuf[n:])
				copy(tailBuf[shift:], tempBuf[:n])
			}
			totalRead += n
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("error reading stream for footer: %w", err)
		}
	}

	if totalRead == 0 {
		return false, nil
	}

	if bytes.Contains(tailBuf, []byte("-- Dump completed")) {
		return true, nil
	}

	return false, nil
}

// ValidateHeaderFooter menjalankan keduanya dalam satu pass
func ValidateHeaderFooter(reader io.Reader) (headerOK bool, footerOK bool, err error) {
	headBuf := make([]byte, 1024)
	n, err := io.ReadFull(reader, headBuf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, false, fmt.Errorf("failed to read header: %w", err)
	}

	content := headBuf[:n]
	if bytes.Contains(content, []byte("-- MariaDB dump")) || bytes.Contains(content, []byte("-- MySQL dump")) {
		headerOK = true
	}

	if err == io.EOF || err == io.ErrUnexpectedEOF {
		if bytes.Contains(content, []byte("-- Dump completed")) {
			footerOK = true
		}
		return headerOK, footerOK, nil
	}

	const tailSize = 4096
	tailBuf := make([]byte, tailSize)
	tempBuf := make([]byte, tailSize)

	if n > 0 {
		if n >= tailSize {
			copy(tailBuf, headBuf[n-tailSize:n])
		} else {
			shift := tailSize - n
			copy(tailBuf[shift:], headBuf[:n])
		}
	}

	for {
		nRead, err := reader.Read(tempBuf)
		if nRead > 0 {
			if nRead >= tailSize {
				copy(tailBuf, tempBuf[nRead-tailSize:nRead])
			} else {
				shift := tailSize - nRead
				copy(tailBuf, tailBuf[nRead:])
				copy(tailBuf[shift:], tempBuf[:nRead])
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return headerOK, false, fmt.Errorf("error reading stream for footer: %w", err)
		}
	}

	if bytes.Contains(tailBuf, []byte("-- Dump completed")) {
		footerOK = true
	}

	return headerOK, footerOK, nil
}
