package stats

import (
	"fmt"
	"sort"

	aperrors "gateway/errors"
)

// Operator defines operators that can be used by Constraints.
type Operator string

// Operator constants.
const (
	LT  Operator = "<"
	LTE Operator = "<="
	GT  Operator = ">"
	GTE Operator = ">="
	EQ  Operator = "="
	IN  Operator = "IN"
)

var (
	validSamples = map[string]bool{
		`node`:      true,
		`timestamp`: true,
	}

	validMeasurements = map[string]bool{
		`api.id`:          true,
		`request.size`:    true,
		`request.id`:      true,
		`response.time`:   true,
		`response.size`:   true,
		`response.status`: true,
		`response.error`:  true,
	}

	validOperators = map[Operator]bool{
		LT:  true,
		LTE: true,
		GT:  true,
		GTE: true,
		EQ:  true,
		IN:  true,
	}
)

// The package user may sample on all measurements plus node, timestamp, api.id,
// etc.
func init() {
	for k := range validMeasurements {
		validSamples[k] = true
	}
}

// Valid determines whether the given Operator is valid.
func (o Operator) Valid() bool {
	return validOperators[o]
}

// AllMeasurements returns a sorted slice of all valid measurement names.
func AllMeasurements() []string {
	toReturn := make([]string, len(validMeasurements))
	i := 0

	for k := range validMeasurements {
		toReturn[i] = k
		i++
	}

	sort.Strings(toReturn)
	return toReturn
}

// AllSamples returns a sorted slice of all valid sample names.  These are the
// names of all the variables logged besides internal variables.
func AllSamples() []string {
	toReturn := make([]string, len(validSamples))
	i := 0

	for k := range validSamples {
		toReturn[i] = k
		i++
	}

	sort.Strings(toReturn)
	return toReturn
}

// ValidSample returns whether the given sample variable is valid.
func ValidSample(s string) bool {
	return validSamples[s]
}

// Constraint implements stats.Constrainer for sql.
type Constraint struct {
	Key      string
	Operator Operator
	Value    interface{}
}

// Validate returns an API-serializable set of validation errors for the given
// Constraint.
func (c *Constraint) Validate() aperrors.Errors {
	errs := make(aperrors.Errors)

	if k := c.Key; !validSamples[k] {
		errs.Add("key", fmt.Sprintf("invalid measurement %q", k))
	}

	if o := c.Operator; !validOperators[o] {
		errs.Add("operator", fmt.Sprintf("invalid operator %q", o))
	}

	return errs
}
