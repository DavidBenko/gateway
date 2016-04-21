package sql

//go:generate go-bindata -o sql_gen.go -nocompress -pkg sql static/...

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	knownMeasurements *regexp.Regexp
	knownConstraints  *regexp.Regexp

	measurements = []string{
		`request_size`,
		`request_id`,
		`response_time`,
		`response_size`,
		`response_status`,
		`response_error`,
	}
	constraints = []string{
		`api_id`,
		`user_id`,
		`node`,
		`timestamp`,
	}
)

func init() {
	knownMeasurements = regexp.MustCompile(regexp.QuoteMeta(
		"^(" + strings.Join(measurements, "|") + ")$",
	))
	knownConstraints = regexp.MustCompile(regexp.QuoteMeta(
		"^(" + strings.Join(append(constraints, measurements...), "|") + ")$",
	))
}

// Driver is the driver to be used for the given stats logger / sampler.
// This must be one of the given constants.
type Driver string

const (
	// Version is the version of the binary's stats schema.
	Version = 1

	// SQLite3 is the SQLite3 Driver.
	SQLite3 Driver = "sqlite3"

	// Postgres is the Postgres Driver.
	Postgres Driver = "postgres"
)

func dayMillis(t time.Time) int64 {
	return int64(t.Hour()*1000*60*60 +
		t.Minute()*1000*60 +
		t.Second()*1000)
}

// SQL implements stats.Logger and stats.Sampler.
type SQL struct {
	ID string
	*sqlx.DB
}

func (s *SQL) quoteCol(str string) string {
	if Driver(s.DriverName()) == SQLite3 {
		return fmt.Sprintf("`%s`", str)
	}
	return fmt.Sprintf(`"%s"`, str)
}

// Parameters returns the correct number of ?'s or $n's as a slice, depending on
// driver.
func (s *SQL) Parameters(n int) []string {
	if n < 1 {
		return nil
	}

	result := make([]string, n)

	if Driver(s.DriverName()) == SQLite3 {
		for i := 0; i < n; i++ {
			result[i] = "?"
		}
	} else {
		for i := 1; i <= n; i++ {
			result[i-1] = fmt.Sprintf(`$%d`, i)
		}
	}

	return result
}
