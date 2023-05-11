package main

import (
	"fmt"
	"time"
	"promise/promise"
)

func main() {

	// First Test case

	p1 := promise.NewPromise[int]()
	
	p1.Then(func(v interface{}) interface{} {
		fmt.Println("First promise resolved with value: ", v)
		return 10
	}).Then(func(v interface{}) interface{} {
		fmt.Println("Second promise resolved with value: ", v)
		return 20
	}).Finally(func() {
		fmt.Printf("Promise 1 execution finished\n\n")
	})

	p1.Resolve(5)	


	// Second Test case 

	p2 := promise.NewPromise[int]()

	p2.Then(func(v interface{}) interface{} {
		fmt.Println("First promise resolved with value: ", v)
		return fmt.Errorf("something went wrong") // Intentionally throwing an error to check the catch method
	}).Catch(func(err error) interface{} {
		fmt.Println("Error: ", err)
		return -1
	}).Finally(func() {
		fmt.Printf("Promise 2 execution finished\n\n")
	})

	p2.Resolve(10)


	// Third Test case
	
	p3 := promise.NewPromise[string]()

	name, err := fetchName()
	if err != nil {
		p3.Reject(err)
	} else {
		p3.Resolve(name)
	}


	p3.Then(func(v interface{}) interface{} {
		fmt.Println("My name is: ", v)
		// return "" // Uncomment this to run without catch method
		return 5 // Intentionally returning an int to check the catch method
	}).Catch(func(err error) interface{} {
		fmt.Println("Error: ", err)
		return ""
	}).Finally(func() {
		fmt.Println("Promise 3 execution finished")
	})

	time.Sleep(1 * time.Second)
	
 	// wait for the promise to be resolved or rejected
}


func fetchName() (string, error) {
	return "John", nil
}