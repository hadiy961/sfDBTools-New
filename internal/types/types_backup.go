package types

import "time"

// BackupFileInfo menyimpan informasi ringkas tentang file backup.
type BackupFileInfo struct {
	Path    string
	ModTime time.Time
	Size    int64
}
