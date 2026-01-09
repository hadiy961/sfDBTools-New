package consts

// File : pkg/consts/consts_backup.go
// Deskripsi : Constants related to backup operations
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 14 November 2025

// Fixed pattern yang digunakan untuk filename backup.
const FixedBackupPattern = "{database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}"

// BackupWriterBufferSize adalah ukuran buffer untuk buffered I/O saat menulis output backup.
const BackupWriterBufferSize = 256 * 1024

// FilenameGenerateErrorPlaceholder digunakan saat preview filename gagal dibuat.
const FilenameGenerateErrorPlaceholder = "error_generating_filename"

// MaxDisplayDatabases adalah jumlah maksimal database yang ditampilkan secara detail dalam output combined backup.
const MaxDisplayDatabases = 10
