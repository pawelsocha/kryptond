package config

import "errors"

var (
	ErrKryptonSectionRead  = errors.New("Can't read section krypton.")
	ErrDatabaseSectionRead = errors.New("Can't read database config.")
	ErrMikrotikSectionRead = errors.New("Can't read s.")
)
