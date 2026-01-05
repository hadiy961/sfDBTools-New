// File : internal/ui/doc.go
// Deskripsi : Package facade UI untuk seluruh aplikasi (pintu tunggal UI)
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

// Package ui adalah entry point facade UI.
//
// Struktur:
// - internal/ui/print|table|progress|text|style: output
// - internal/ui/prompt: public API untuk layer app
// - internal/ui/input: engine prompt (berbasis survey)
//
// Package internal ini menjadi pintu tunggal UI untuk layer app/cli/services.
package ui
