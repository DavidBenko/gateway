package util

func Drain(c chan []byte) {
	for _ = range c {
	}
}
