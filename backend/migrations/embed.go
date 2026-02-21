package migrations

import "embed"

// FS holds all migration *.sql files in this directory for use with golang-migrate iofs driver.
//
//go:embed *.sql
var FS embed.FS
