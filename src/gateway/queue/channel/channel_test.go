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

			send, err := queue.PubChannel(path, Publish())
			if err != nil {
				t.Fatal(err)
			}
			for {
				send <- []byte("hello world")
			}
		}()

		done := make([]chan bool, 8)
		for i := 0; i < 8; i++ {
			done[i] = make(chan bool, 1)
			go func(done chan bool) {
				rec, err := queue.SubChannel(path, Subscribe())
				if err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 8; i++ {
					<-rec
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
