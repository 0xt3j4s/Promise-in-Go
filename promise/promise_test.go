package promise_test

import (
	"fmt"
	"sync"
	"testing"

	"promise/promise"
)

// Testing the promise package

func TestPromise(t *testing.T) {
	// Create a WaitGroup to synchronize the test function with the promises
	var wg sync.WaitGroup

	// Increment the WaitGroup counter to account for the promises
	wg.Add(1)

	// Testing resolving a promise with a valid value
	p1 := promise.NewPromise[int]()

	p1.Then(func(v interface{}) interface{} {
		actual := v.(int)
		expected := 5
		if actual != expected {
			t.Errorf("Expected: %d, Got: %d", expected, actual)
		}
		return 10
	}).Then(func(v interface{}) interface{} {
		actual := v.(int)
		expected := 10
		if actual != expected {
			t.Errorf("Expected: %d, Got: %d", expected, actual)
		}
		return 20
	}).Finally(func() {
		fmt.Printf("Promise 1 execution finished\n\n")
		wg.Done() // Decrement the WaitGroup counter when the promise is resolved
	})

	p1.Resolve(5)

	// Wait for all promises to be resolved before completing the test
	wg.Wait()
}
