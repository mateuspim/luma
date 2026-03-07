package ddc

import "errors"

// ErrBusy is returned when the executor drops a command because it is at capacity.
var ErrBusy = errors.New("ddc: executor busy, command dropped")

// ErrNotFound is returned when ddcutil is not installed.
var ErrNotFound = errors.New("ddc: ddcutil not found, please install it")
