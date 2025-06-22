package main

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

// Test 1: context.AfterFunc

// non-synctest version
func TestAfterFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calledCh := make(chan struct{}) // closed when AfterFunc is called
	context.AfterFunc(ctx, func() {
		close(calledCh)
	})

	// funcCalled reports whether the function was called.
	funcCalled := func() bool {
		select {
		case <-calledCh:
			return true
		case <-time.After(10 * time.Millisecond):
			return false
		}
	}

	if funcCalled() {
		t.Fatalf("AfterFunc function called before context is canceled")
	}

	cancel()

	if !funcCalled() {
		t.Fatalf("AfterFunc function not called after context is canceled")
	}
}

// synctest version
func TestAfterFuncSyncTest(t *testing.T) {
	synctest.Run(func() {
		ctx, cancel := context.WithCancel(context.Background())

		funcCalled := false
		context.AfterFunc(ctx, func() { funcCalled = true })

		// Warte bis alle Goroutinen blockiert sind
		synctest.Wait()
		if funcCalled {
			t.Fatalf("AfterFunc function called before context is canceled")
		}

		cancel()

		synctest.Wait()
		if !funcCalled {
			t.Fatalf("AfterFunc function not called after context is canceled")
		}
	})
}
