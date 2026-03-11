package migrations

import "embed"

// Files contains SQL migrations for the application.
//
//go:embed *.sql
var Files embed.FS
