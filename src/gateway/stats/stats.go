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

// Logger is an interface defining the necessary methods for a stats logger.  It
// must be concurrency-safe.
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

// Sampler defines the methods a stats sampler must implement.  It must be
// concurrency-safe.
type Sampler interface {
	// Sample gets the logged values using the given constraints, given a
	// slice of measurements to return for result values.  Sample may be
	// terminated (since sampling a great many points may take a long time)
	// at any point by closing the given channel.
	Sample(
		[]Constraint,
		<-chan struct{},
		...string,
	) (Result, error)
}
