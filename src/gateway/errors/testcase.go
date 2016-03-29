package errors

// TestCase is a test-case for a validation error.
type TestCase struct {
	// Valid should be true if the case is not an error.
	Valid bool
	// FailField is the component which will be reported to have an error.
	FailField string
	// FailMessage is the error message for the FailField.
	FailMessage string
}

// ValidateCases returns a slice of Errors, one for each failing TestCase.
func ValidateCases(cases ...TestCase) Errors {
	errors := make(Errors)

	for _, c := range cases {
		if !c.Valid {
			errors.Add(c.FailField, c.FailMessage)
		}
	}

	return errors
}
