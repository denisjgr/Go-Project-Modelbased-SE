package main

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
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

// Test 2: context.WithTimeout
func TestWithTimeout(t *testing.T) {
	synctest.Run(func() {
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		time.Sleep(timeout - time.Nanosecond)
		synctest.Wait()
		if err := ctx.Err(); err != nil {
			t.Fatalf("before timeout, ctx.Err() = %v; want nil", err)
		}

		time.Sleep(time.Nanosecond)
		synctest.Wait()
		if err := ctx.Err(); err != context.DeadlineExceeded {
			t.Fatalf("after timeout, ctx.Err() = %v; want DeadlineExceeded", err)
		}
	})
}

// Test 3: HTTP Expect: 100-continue Mechanismus
func TestHTTPExpectContinue(t *testing.T) {
	synctest.Run(func() {
		srvConn, cliConn := net.Pipe()
		defer func(srvConn net.Conn) {
			err := srvConn.Close()
			if err != nil {

			}
		}(srvConn)
		defer func(cliConn net.Conn) {
			err := cliConn.Close()
			if err != nil {

			}
		}(cliConn)

		tr := &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return cliConn, nil
			},
			ExpectContinueTimeout: 5 * time.Second,
		}

		body := "request body"
		go func() {
			req, _ := http.NewRequest("PUT", "http://test.tld/", strings.NewReader(body))
			req.Header.Set("Expect", "100-continue")
			resp, err := tr.RoundTrip(req)
			if err != nil {
				t.Errorf("RoundTrip: unexpected error %v", err)
			} else {
				err := resp.Body.Close()
				if err != nil {
					return
				}
			}
		}()

		req, err := http.ReadRequest(bufio.NewReader(srvConn))
		if err != nil {
			t.Fatalf("ReadRequest: %v", err)
		}

		var gotBody strings.Builder
		go func() {
			_, err := io.Copy(&gotBody, req.Body)
			if err != nil {

			}
		}()
		synctest.Wait()
		if got := gotBody.String(); got != "" {
			t.Fatalf("before sending 100 Continue, unexpectedly read body: %q", got)
		}

		_, err = srvConn.Write([]byte("HTTP/1.1 100 Continue\r\n\r\n"))
		if err != nil {
			return
		}
		synctest.Wait()
		if got := gotBody.String(); got != body {
			t.Fatalf("after sending 100 Continue, read body %q, want %q", got, body)
		}

		_, err = srvConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			return
		}
	})
}

// Test 4: sync.Once: Verify, dass Do genau einmal ausgeführt wird
func TestOnceDo(t *testing.T) {
	synctest.Run(func() {
		var once sync.Once
		counter := 0

		// mehrfach aufrufen
		once.Do(func() { counter++ })
		once.Do(func() { counter++ })
		once.Do(func() { counter++ })

		if counter != 1 {
			t.Fatalf("expected Do to run once, but ran %d times", counter)
		}
	})
}

// Test 5: sync.Mutex: Stelle sicher, dass Unlock auch nach Lock funktioniert
func TestMutexLockUnlock(t *testing.T) {
	synctest.Run(func() {
		var mu sync.Mutex
		locked := false

		// Goroutine 1: lock, set flag, unlock
		go func() {
			mu.Lock()
			locked = true
			mu.Unlock()
		}()

		// Goroutine 2: versuche zu locken, erst nach Unlock möglich
		synctest.Wait()
		if !locked {
			t.Fatalf("expected first goroutine to acquire lock before waiting")
		}

		synctest.Wait() // jetzt blockiert, bis Unlock wieder kommt
		// Wenn wir hier ankommen, heißt das: keine Deadlocks mehr
	})
}

// Test 6: sync.WaitGroup: Warte auf alle Done-Calls
func TestWaitGroup(t *testing.T) {
	synctest.Run(func() {
		var wg sync.WaitGroup
		const routines = 3
		counter := 0

		wg.Add(routines)
		for i := 0; i < routines; i++ {
			go func() {
				defer wg.Done()
				counter++
			}()
		}

		// bis alle Done() aufgerufen wurden, bleibt Wait() blockiert
		synctest.Wait()
		// nachdem alle wg.Done waren, sollte counter==routines sein
		if counter != routines {
			t.Fatalf("expected counter %d, got %d", routines, counter)
		}
	})
}

// Test 7: Kanäle: Buffered Channel send/receive ohne Deadlock
func TestChannelBuffer(t *testing.T) {
	synctest.Run(func() {
		ch := make(chan int, 1)
		ch <- 42

		val := <-ch
		if val != 42 {
			t.Fatalf("expected 42, got %d", val)
		}
	})
}
