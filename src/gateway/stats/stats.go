package stats

import "time"

const (
	// Resolution is the resolution of the stats tables.
	Resolution = time.Millisecond
)

// Point is a timestamp and map of measurement name to value.
type Point struct {
	Timestamp time.Time
	Values    map[string]interface{}
}

// Logger is an interface defining the necessary methods for a stats logger.
type Logger interface {
	// Log logs a set of Points to a logging target.
	Log(...Point) error
}

// Row is a row of the results of a Sampler query.
type Row struct {
	Node      string
	Timestamp time.Time
	Err       error
	Values    map[string]interface{}
}

// Result is the result of a Sampler query.  It is an alias for []Row.
type Result []Row

// Sampler defines the methods a stats sampler must implement.
type Sampler interface {
	// Sample gets the values matching the given tags over the given time
	// interval, optionally given a slice of measurements to restrict
	// results by.
	Sample(
		map[string]interface{},
		time.Time,
		time.Time,
		...string,
	) (Result, error)
}
