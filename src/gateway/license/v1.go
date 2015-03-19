package license

import "time"

// V1 holds basic license data and is time-expirable.
type V1 struct {
	Name       string
	Company    string
	Id         string
	Expiration *time.Time
}

func (l *V1) version() int {
	return 1
}

func (l *V1) valid() bool {
	if l.Expiration == nil {
		return true
	}

	return (*l.Expiration).After(time.Now())
}
