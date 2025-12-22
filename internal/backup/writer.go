// File : internal/backup/writer.go
// Deskripsi : Writer dan mysqldump execution logic
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 2025-12-05

package backup

// Split from writer.go into:
// - writer_mysqldump.go (mysqldump execution)
// - writer_pipeline.go (compression/encryption pipeline)
// - writer_key.go (key resolution)
// - writer_output.go (buffered output file)
// - writer_monitor.go (progress monitor writer)
