package license

import (
	"testing"
	"time"
)

func TestV1Version(t *testing.T) {
	l := V1{}
	if l.version() != 1 {
		t.Error("Expected version to be 1")
	}
}

func TestV1ExpirationNever(t *testing.T) {
	l := V1{}
	if !l.valid() {
		t.Error("Expected non-expiring license to be valid")
	}
}

func TestV1ExpirationFuture(t *testing.T) {
	future := time.Now().Add(time.Minute)
	l := V1{Expiration: &future}
	if !l.valid() {
		t.Error("Expected future-expiring license to be valid")
	}
}

func TestV1ExpirationPast(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	l := V1{Expiration: &past}
	if l.valid() {
		t.Error("Expected past-expiring license not to be valid")
	}
}
