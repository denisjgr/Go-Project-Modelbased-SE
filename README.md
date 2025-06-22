# Testing concurrent code with testing/synctest

Link topic from The Go Blog: https://go.dev/blog/synctest

Used Go version: go 1.24.1 windows/amd64

Used Git version: 2.21.0.windows.1

## What issue is meant to be solved?
### Issue
The synctest package addresses the difficulty of writing correct concurrency tests for 
data structures from the sync package (e.g., sync.Map, sync.Pool, sync.WaitGroup). Traditional tests 
for concurrent operations are error-prone because they rely on timing-based primitives like time.Sleep 
or complex synchronization mechanisms that:
    
    - Cause flaky tests (non-deterministic behavior)
    - Are difficult to debug
    - Enable incomplete test coverage

### Solution
synctest provides deterministic testing utilities that:

    - Enable control over goroutine scheduling in tests
    - Deliberately provoke and verify race conditions
    - Enforce complete coverage of all execution paths


## What are typical use cases?

    - Testing sync.Map operations (Store, Load, Delete) under race conditions.
    - Validating sync.Pool behavior during concurrent access.
    - Ensuring sync.WaitGroup correctly handles Done()/Wait()
    - Reproducing rare race conditions in a controlled environment

## Is the solution provided fully satisfactory?
### Advantages

    - Determinism: Tests are reproducible and flaky-free
    - Simplification: Replaces complex time.Sleep / channel constructs with clear APIs
    - Complete coverage: Enforces all possible scheduling paths (e.g., via synctest.Explore)
    - Integration: Directly available in the standard library from Go 1.21

### Limitations

    - Learning curve: New concepts like "scheduling traces" require rethinking
    - Overhead: Tests run slower as all scheduling paths are executed
    - Not a replacement for -race: Static analysis via go test -race remains necessary

### Conclusion
The solution is practical and valuable but not a panacea. It ideally complements existing tools for 
deterministic concurrency testing.

## Is the topic of great importance?
Yes, for three reasons:
### 1. Critical safety:
Concurrency bugs cause severe issues (data races, deadlocks). synctest enables systematic detection.
### 2. Productivity boost:
Developers save time through

    - Elimination of flaky tests
    - Targeted reproduction of race conditions
    - Automated coverage of all scheduling paths
### 3. Ecosystem impact:
Standardized testing of sync components promotes more robust libraries (e.g., database clients, caches).


# Setting up the programming environment
    - Installing Go version 1.24.1 windows/amd64
    - Set up GoLand IDE
    - Create new Project and link it with new GitHub repository
    - Set environmental variable for the use of synctest (GOEXPERIMENT=synctest)

# Code examples
In sync_test.go there are the examples from the Go Blog post.
## TestAfterFunc and TestAfterFuncSyncTest
Here we are testing the context.AfterFunc function with and without synctest.
## TestWithTimeout
Here we are testing the context.WithTimeout function with the help of bubbles and their fake clocks.
## TestHTTPExpectContinue
Here we are testing the net/http package's handling of the 100 Continue response.
## TestOnceDo
Here we are testing the sync.Once object which will perform exactly one action.
## TestMutexLockUnlock
Here we are testing the sync.Mutex Lock/Unlock capabilities.
## TestWaitGroup
Here we are testing sync.WaitGroup which waits for a collection of goroutines to finish.