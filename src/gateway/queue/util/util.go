package util

// Drain expects c to be closed.  It will drain all contents from c.  If c is
// not closed, this will block indefinitely.
func Drain(c chan []byte) {
	for _ = range c {
	}
}
