package main

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

/*
	Тест, предложенный одним из учащихся курса, Ilya Boltnev
	https://www.coursera.org/learn/golang-webservices-1/discussions/weeks/2/threads/kI2PR_XtEeeWKRIdN7jcig

	В чем его преимущество по сравнению с TestPipeline?
	1. Он проверяет то, что все функции действительно выполнились
	2. Он дает представление о влиянии time.Sleep в одном из звеньев конвейера на время работы

	возможно кому-то будет легче с ним
	при правильной реализации ваш код конечно же должен его проходить
*/

func first(in, out chan interface{}) {
	//runtime.Breakpoint()
	out <- uint32(1)
	out <- uint32(3)
	out <- uint32(4)
}

func second(in, out chan interface{}) {
	//runtime.Breakpoint()
	for val := range in {
		out <- val.(uint32) * 3
		time.Sleep(time.Millisecond * 100)
	}
}

func third(in, out chan interface{}, received *uint32) {
	//runtime.Breakpoint()
	for val := range in {
		fmt.Println("collected", val)
		atomic.AddUint32(received, val.(uint32))
	}
}

func TestByIlia(t *testing.T) {

	var received uint32
	freeFlowJobs := []job{
		job(first),
		job(second),
		job(func(in, out chan interface{}){
			third(in, out, &received)
		}),
	}

	start := time.Now()

	ExecutePipeline(freeFlowJobs...)

	end := time.Since(start)

	expectedTime := time.Millisecond * 350

	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}

	if received != (1+3+4)*3 {
		t.Errorf("f3 have not collected inputs, recieved = %d", received)
	}
}
