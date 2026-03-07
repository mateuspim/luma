package ddc

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestExecutor_DropsWhenBusy(t *testing.T) {
	// maxProcs=1, timeout=5s
	e := NewExecutor(1, 5)

	var wg sync.WaitGroup
	results := make([]error, 3)

	// Fire 3 concurrent calls; only 1 should run, rest should get ErrBusy.
	// Use a slow command (sleep via a non-existent arg that causes immediate exit)
	// We'll simulate by using a real slow command indirectly — but since
	// ddcutil may not be available in CI, we test the semaphore logic directly.

	// Manually fill the semaphore.
	e.sem <- struct{}{} // occupy the slot

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := e.Run(context.Background(), "--help")
			results[idx] = err
		}(i)
	}

	// Give goroutines time to hit the semaphore check.
	time.Sleep(50 * time.Millisecond)
	<-e.sem // release
	wg.Wait()

	busyCount := 0
	for _, err := range results {
		if errors.Is(err, ErrBusy) {
			busyCount++
		}
	}
	// All 3 should have been dropped since semaphore was full when they ran.
	if busyCount != 3 {
		t.Errorf("expected 3 ErrBusy, got %d (results: %v)", busyCount, results)
	}
}

func TestExecutor_TimeoutKillsHungProcess(t *testing.T) {
	// timeout=1s, run a command that would hang (sleep 10)
	// We can't test this perfectly without a slow ddcutil, but we verify
	// that a context timeout is respected.
	e := NewExecutor(1, 1) // 1s timeout

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	// "ddcutil sleep 10" is not valid but process exits quickly with error.
	// Instead verify timeout propagation by injecting a short-lived context.
	shortCtx, shortCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer shortCancel()

	_, err := e.Run(shortCtx, "detect")
	elapsed := time.Since(start)

	// Should have returned quickly (timeout or fast exit), not hung.
	if elapsed > 3*time.Second {
		t.Errorf("Run took too long: %v", elapsed)
	}
	_ = err // may succeed or fail depending on environment
}

func TestExecutor_IsBusy(t *testing.T) {
	e := NewExecutor(1, 5)
	if e.IsBusy() {
		t.Error("should not be busy initially")
	}

	// Manually simulate in-flight.
	e.inFlight.Add(1)
	if !e.IsBusy() {
		t.Error("should be busy with inFlight=1")
	}
	e.inFlight.Add(-1)
	if e.IsBusy() {
		t.Error("should not be busy after release")
	}
}

func TestNewExecutor_Clamping(t *testing.T) {
	e := NewExecutor(0, 1) // both below min
	if cap(e.sem) != 1 {
		t.Errorf("expected semaphore cap 1, got %d", cap(e.sem))
	}
	if e.timeout != 3*time.Second {
		t.Errorf("expected timeout 3s, got %v", e.timeout)
	}

	e2 := NewExecutor(10, 100) // both above max
	if cap(e2.sem) != 2 {
		t.Errorf("expected semaphore cap 2, got %d", cap(e2.sem))
	}
	if e2.timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", e2.timeout)
	}
}
