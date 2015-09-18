package testing

import (
	"fmt"
	"time"
)

var ShortWait = time.Duration(10 * time.Millisecond)
var LongWait = time.Duration(100 * time.Millisecond)

func TrySend(ch chan []byte, messages [][]byte) (e error) {
	for _, m := range messages {
		// try to send a few times
		err := func(msg []byte) error {
			for {
			Send:
				select {
				case ch <- msg:
					return nil
				case <-time.After(ShortWait):
					break Send
				case <-time.After(LongWait):
					return fmt.Errorf("error sending message %q", msg)
				}
			}
		}(m)

		if err != nil {
			return err
		}
	}

	close(ch)

	return nil
}

func MakeBytes(ms ...string) [][]byte {
	bs := make([][]byte, len(ms))
	for i, m := range ms {
		bs[i] = []byte(m)
	}
	return bs
}
