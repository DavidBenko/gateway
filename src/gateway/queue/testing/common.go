package testing

import (
	"fmt"
	"gateway/queue"
	"runtime"
	"time"
)

var ShortWait = time.Duration(10 * time.Millisecond)
var LongWait = time.Duration(100 * time.Millisecond)

func IsReplyClosed(reply chan struct{}) bool {
	SyncGC()

	select {
	case _ = <-reply:
		return true
	case _ = <-time.After(ShortWait):
		return false
	}
}

func IsSubChanClosed(ch <-chan []byte) bool {
	SyncGC()

	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func IsPubChanClosed(ch chan<- []byte) (isClosed bool) {
	SyncGC()

	defer func() {
		if r := recover(); r != nil {
			isClosed = true
		}
	}()

	ch <- []byte("test")

	return false
}

func TrySend(ch *queue.PubChannel, messages [][]byte) (e error) {
	for _, m := range messages {
		// try to send a few times
		err := func(msg []byte) error {
			for {
			Send:
				select {
				case ch.C <- msg:
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

	return nil
}

func SyncGC() {
	for i := 0; i < 3; i++ {
		runtime.GC()
		runtime.Gosched()
	}
}

func MakeMessages(ms ...string) [][]byte {
	bs := make([][]byte, len(ms))
	for i, m := range ms {
		bs[i] = []byte(m)
	}
	return bs
}
