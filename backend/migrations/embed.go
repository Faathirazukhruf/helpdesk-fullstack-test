package migrations

import "embed"

// FS menampung seluruh file SQL migrasi yang dikompilasi langsung ke dalam binary aplikasi.
//
//go:embed *.sql
var FS embed.FS
