package main

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
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
