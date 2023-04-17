package main

import (
	"fmt"
	"sync"
	"reflect"
	"time"
)

type Promise interface {
	Resolve(v interface{})
	Reject(err error)
	Catch(f func(err error) interface{}) Promise
	Then(f func(v interface{}) interface{}) Promise
	Finally(f func()) Promise
	HandlePanic()
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
		} else {
			p.err = fmt.Errorf("Resolve: cannot assign %v (type %v) to type *%v", value, val.Type(), reflect.TypeOf((*T)(nil)).Elem())
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

// newPromise.HandlePanic undefined (type Promise has no field or method HandlePanic)
func (p *promise[T]) Then(f func(interface{}) interface{}) Promise {

	newPromise := NewPromise[T]()
	go func() {
		defer newPromise.HandlePanic()
		<-p.CH
		if p.err != nil {
			newPromise.Reject(p.err)
		} else {
			result := f(p.value)
			switch val := result.(type) {
			case error:
				newPromise.Reject(val)
			default:
				newPromise.Resolve(result)
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
		newPromise.Resolve(p.value)
	}()

	return newPromise
}

func (p *promise[T]) HandlePanic() {
	if r := recover(); r != nil {
		p.Reject(r.(error))
	}
}


func main() {
	// create a new promise

	fmt.Println("Hello Go, we doing promises today!")
	p := NewPromise[int]()

	// chain promises with Then and Finally
	p.Then(func(val interface{}) interface{} {
		// multiply the value by 2
		return val.(int) * 2
	}).Finally(func() {
		// print "done" when the promise is resolved or rejected
		fmt.Println("done")
	})

	// resolve the promise with a value of 5
	p.Resolve(5)

	// wait for the promise to be resolved or rejected
	<-p.(*promise[int]).CH

	time.Sleep(1 * time.Second)
}
