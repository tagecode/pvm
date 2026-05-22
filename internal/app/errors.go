package app

import "fmt"

const (
	ExitOK       = 0
	ExitUser     = 1
	ExitSystem   = 2
	ExitPartial  = 3
	ExitCanceled = 130
)

type ExitError struct {
	Code    int
	Message string
	Hint    string
}

func (e *ExitError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s (hint: %s)", e.Message, e.Hint)
	}
	return e.Message
}

func UserError(msg, hint string) *ExitError {
	return &ExitError{Code: ExitUser, Message: msg, Hint: hint}
}

func SystemError(msg, hint string) *ExitError {
	return &ExitError{Code: ExitSystem, Message: msg, Hint: hint}
}

var (
	ErrAlreadyInstalled = UserError("version already installed", "use --reinstall to force")
	ErrNotInstalled     = UserError("version not installed", "run pvm install <version>")
	ErrActiveVersion    = UserError("cannot remove active version", "run pvm use <other> first")
	ErrNotFound         = UserError("version not found", "run pvm list to see installed versions")
)
