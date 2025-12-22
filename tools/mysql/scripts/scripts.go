package scripts

import (
	_ "embed"
)

var (
	//go:embed 2-outbox.sql
	CreateOutboxTable []byte
)
