// File : pkg/fsops/fsops_check.go
// Deskripsi : Helper functions untuk file dan directory existence checks
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025

package fsops

import (
	"os"
)

// FileExists memeriksa apakah file ada dan merupakan file (bukan directory)
// Returns: true jika file ada dan merupakan file biasa
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists memeriksa apakah directory ada dan merupakan directory
// Returns: true jika path ada dan merupakan directory
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// PathExists memeriksa apakah path (file atau directory) ada
// Returns: true jika path ada (tidak peduli file atau directory)
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileExistsWithInfo memeriksa apakah file ada dan mengembalikan os.FileInfo jika ada
// Returns: (info, true) jika file ada, (nil, false) jika tidak ada atau ada error
func FileExistsWithInfo(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if info.IsDir() {
		return nil, false
	}
	return info, true
}

// DirExistsWithInfo memeriksa apakah directory ada dan mengembalikan os.FileInfo jika ada
// Returns: (info, true) jika directory ada, (nil, false) jika tidak ada atau ada error
func DirExistsWithInfo(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if !info.IsDir() {
		return nil, false
	}
	return info, true
}
