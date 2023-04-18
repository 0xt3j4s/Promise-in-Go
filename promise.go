package main

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Defining the Promise Interface and the Promise Struct

type Promise interface {
	Resolve(v interface{}) // Resolve Method
	Reject(err error) // Reject Method
	Catch(f func(err error) interface{}) Promise // Catch Method
	Then(f func(v interface{}) interface{}) Promise // Then Method
	Finally(f func()) Promise // Finally Method
	HandlePanic() // HandlePanic Method
}

type promise[v any] struct {

	value *v
	err   error
	state string
	CH chan struct{}
	once  sync.Once

}

func NewPromise[T any]() Promise {

	return &promise[T]{
		value: nil,
		err:   nil,
		state: "pending",
		CH: make(chan struct{}),
	}
}


func (p *promise[T]) Resolve(value interface{}) {
	if p.state != "pending" {
		return
	}

	p.once.Do(func() {
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.IsValid() && val.Type().AssignableTo(reflect.TypeOf((*T)(nil)).Elem()) {
			v := val.Interface().(T)
			p.value = &v
			p.state = "fulfilled"
		} else {
			p.err = fmt.Errorf("Resolve: cannot assign %v (type %v) to type *%v", value, val.Type(), reflect.TypeOf((*T)(nil)).Elem())
			p.state = "rejected"
		}

		close(p.CH)
	})
}



func (p *promise[T]) Reject(err error) {
	if (p.state != "pending") {
		return
	}

	p.once.Do(func() {
		p.err = err
		close(p.CH)
	})
}

func (p *promise[T]) Catch(f func(error) interface{}) Promise {
	newPromise := NewPromise[T]()
	go func() {
		defer newPromise.HandlePanic()
		<-p.CH
		if p.err != nil {
			result := f(p.err)
			if (reflect.TypeOf(result) != reflect.TypeOf(*p.value)) {
				fmt.Println("Error: Wrong return type from 'Catch' method...")
			}
			switch val := result.(type) {
			case error:
				newPromise.Reject(val)
			default:
				newPromise.Resolve(result)
			}
		} else {
			newPromise.Resolve(p.value)
		}
	}()
	
	// if promise is already resolved or rejected, skip the function
	if p.state != "pending" {
		go func() {
			<-newPromise.(*promise[T]).CH
		}()
	}

	return newPromise
}


func (p *promise[T]) Then(f func(interface{}) interface{}) Promise {

	newPromise := NewPromise[T]()
	go func() {
		defer newPromise.HandlePanic()
		<-p.CH
		if p.err != nil {
			newPromise.Reject(p.err)
		} else {
			result := f(*p.value) // Reading the value at the pointer
			if (reflect.TypeOf(result) != reflect.TypeOf(*p.value)) {
				fmt.Println("Error: Wrong return type from 'Then' method.")
			}
			switch val := result.(type) {
			case error:
				newPromise.Reject(val) 	// if the result is an error, reject the promise
			default:
				newPromise.Resolve(result) // if the result is not an error, resolve the promise
			}
		}
	}()

	// if promise is already resolved or rejected, skip the function
	if p.state != "pending" {
		go func() {
			<-newPromise.(*promise[T]).CH
		}()
	}

	return newPromise
}

func (p *promise[T]) Finally(f func()) Promise {
	newPromise := NewPromise[T]()
	go func() {
		defer newPromise.HandlePanic()
		<-p.CH
		f()
		if p.err != nil {
			newPromise.Reject(p.err)
		} else {
			newPromise.Resolve(p.value)
		}
	}()

	return newPromise
}

func (p *promise[T]) HandlePanic() {
	if r := recover(); r != nil {
		p.Reject(r.(error))
	}
}


func main() {

	// First Test case
	
	p1 := NewPromise[int]()
	
	p1.Then(func(v interface{}) interface{} {
		fmt.Println("First promise resolved with value: ", v)
		return 10
	}).Then(func(v interface{}) interface{} {
		fmt.Println("Second promise resolved with value: ", v)
		return 20
	}).Finally(func() {
		fmt.Printf("Promise execution finished\n\n")
	})

	p1.Resolve(5)	


	// Second Test case 

	p2 := NewPromise[int]()

	p2.Then(func(v interface{}) interface{} {
		fmt.Println("First promise resolved with value: ", v)
		return fmt.Errorf("something went wrong") // Intentionally throwing an error to check the catch method
	}).Catch(func(err error) interface{} {
		fmt.Println("Error: ", err)
		return -1
	}).Finally(func() {
		fmt.Printf("Promise execution finished\n\n")
	})

	p2.Resolve(10)


	// Third Test case
	
	p3 := NewPromise[string]()

	name, err := fetchName()
	if err != nil {
		p3.Reject(err)
	} else {
		p3.Resolve(name)
	}


	p3.Then(func(v interface{}) interface{} {
		fmt.Println("My name is: ", v)
		// return "" // Uncomment this to run wihout catch method
		return 5 // Intentionally returning an int to check the catch method
	}).Catch(func(err error) interface{} {
		fmt.Println("Error: ", err)
		return ""
	}).Finally(func() {
		fmt.Println("Promise execution finished")
	})

	
	
	time.Sleep(1 * time.Second)
	
 	// wait for the promise to be resolved or rejected
}


func fetchName() (string, error) {
	return "John", nil
}