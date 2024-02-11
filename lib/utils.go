package lib

import (
	"errors"
)

var (
	ErrMainFileNotFound    error = getErr("specified main file not found")
	ErrNoMainFile          error = getErr("no main file specified in config or command")
	ErrEmptyAllowList      error = getErr("cannot have empty allow list")
	ErrReadingConfigFile   error = getErr("error reading config file")
	ErrParsingConfigFile   error = getErr("error while parsing config file")
	ErrStartChildProcess   error = getErr("error while starting program")
	ErrRestartChildProcess error = getErr("error while restarting program")
	ErrKillChildProcess    error = getErr("error while killing program")
	ErrMainFileIsDir       error = getErr("main file should'nt be a directory")
	ErrSettingPWD          error = getErr("error while setting present working directory")
	ErrWatcherSetup        error = getErr("error while setting up watcher")
	ErrWalkingFS           error = getErr("error while walking file system")
)

func getErr(err string) error {
	return errors.New(BoldRedColor + err + ResetColor)
}

var MONTRE_LOG string = BoldCyanColor + "[montre]" + ResetColor + " "

var (
	ResetColor    string = "\033[0m"
	GreenColor    string = "\033[0;32m"
	YellowColor   string = "\033[0;33m"
	BoldRedColor  string = "\033[1;31m"
	BoldCyanColor string = "\033[1;36m"
)

func MontreLog(log string) string {
	return MONTRE_LOG + log
}

func RedLog(log string) string {
	return BoldRedColor + log + ResetColor
}

func YellowLog(log string) string {
	return YellowColor + log + ResetColor
}

func GreenLog(log string) string {
	return GreenColor + log + ResetColor
}
