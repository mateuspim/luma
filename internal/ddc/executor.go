package ddc

import (
	"context"
	"os/exec"
	"sync/atomic"
	"time"
)

// Executor serializes ddcutil calls with a semaphore and per-call timeout.
// Commands are dropped (not queued) when the semaphore is full.
type Executor struct {
	sem     chan struct{}
	timeout time.Duration
	inFlight atomic.Int32
}

// NewExecutor creates an Executor. maxProcs is clamped to [1,2].
func NewExecutor(maxProcs int, timeoutS int) *Executor {
	if maxProcs < 1 {
		maxProcs = 1
	}
	if maxProcs > 2 {
		maxProcs = 2
	}
	if timeoutS < 3 {
		timeoutS = 3
	}
	if timeoutS > 30 {
		timeoutS = 30
	}
	return &Executor{
		sem:     make(chan struct{}, maxProcs),
		timeout: time.Duration(timeoutS) * time.Second,
	}
}

// Run executes ddcutil with the given args.
// Returns ErrBusy if the semaphore is full (command dropped).
// Returns context.DeadlineExceeded if the command hangs beyond the timeout.
func (e *Executor) Run(ctx context.Context, args ...string) (string, error) {
	// Try to acquire semaphore without blocking.
	select {
	case e.sem <- struct{}{}:
	default:
		return "", ErrBusy
	}
	e.inFlight.Add(1)
	defer func() {
		<-e.sem
		e.inFlight.Add(-1)
	}()

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ddcutil", args...).Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", context.DeadlineExceeded
		}
		return "", err
	}
	return string(out), nil
}

// IsBusy returns true when at least one command is in flight.
func (e *Executor) IsBusy() bool {
	return e.inFlight.Load() > 0
}
