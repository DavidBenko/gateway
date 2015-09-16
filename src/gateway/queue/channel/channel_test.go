package channel

import (
	"fmt"
	"gateway/queue"
	"testing"
)

func TestChannel(t *testing.T) {
	test := func(path string) []chan bool {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if s, ok := r.(string); ok && s == "send on closed channel" {
						return
					}
					panic(r)
				}
			}()

			send, err := queue.Publish(path, Publish)
			if err != nil {
				t.Fatal(err)
			}
			for {
				send.C <- []byte("hello world")
			}
		}()

		done := make([]chan bool, 8)
		for i := 0; i < 8; i++ {
			done[i] = make(chan bool, 1)
			go func(done chan bool) {
				rec, err := queue.Subscribe(path, Subscribe)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					go func(c <-chan []byte) {
						for _ = range c {
							//noop
						}
					}(rec.C)
				}()

				i := 0
				for _ = range rec.C {
					i++
					if i == 8 {
						break
					}
				}
				if i != 8 {
					t.Fatalf("%v didn't get 8 messages", path)
				}
				done <- true
			}(done[i])
		}

		return done
	}

	var done []chan bool
	for i := 0; i < 8; i++ {
		path := fmt.Sprintf("test %v", i)
		done = append(done, test(path)...)
	}
	for i, d := range done {
		t.Logf("done %v\n", i)
		<-d
	}
}
