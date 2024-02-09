package lib

import (
	"errors"
)

var (
	ErrMainFileNotFound  error = errors.New(BoldRedColor + "specified main file not found" + ResetColor)
	ErrNoMainFile        error = errors.New(BoldRedColor + "no main file specified in config or command" + ResetColor)
	ErrEmptyAllowList    error = errors.New(BoldRedColor + "cannot have allow list" + ResetColor)
	ErrReadingConfigFile error = errors.New(BoldRedColor + "error reading config file" + ResetColor)
	ErrParsingConfigFile error = errors.New(BoldRedColor + "error while parsing config file" + ResetColor)
	ErrRestartingProcess error = errors.New(BoldRedColor + "error while restarting program" + ResetColor)
)

var (
	ResetColor   string = "\033[0m"
	GreenColor   string = "\033[0;32m"
	YellowColor  string = "\033[0;33m"
	BoldRedColor string = "\033[1;31m"
)
