package admin

import (
	"testing"
)

const ITERATIONS = 16

func TestLogInterceptor(t *testing.T) {
	interceptor := newInterceptor()
	stop := false
	go func() {
		for !stop {
			interceptor.Write([]byte("hello world"))
		}
	}()
	test := func() {
		var done [ITERATIONS]chan bool
		for i := 0; i < ITERATIONS; i++ {
			done[i] = make(chan bool, 1)
			go func(d chan bool) {
				logs, unsubscribe := interceptor.Subscribe()
				defer unsubscribe()
				count := 0
				for log := range logs {
					t.Log(string(log))
					count++
					if count == ITERATIONS {
						d <- true
						close(d)
						return
					}
				}
			}(done[i])
		}
		for _, d := range done {
			<-d
		}
	}
	for i := 0; i < ITERATIONS; i++ {
		test()
	}
	stop = true
}
