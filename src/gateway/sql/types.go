package sql

import (
	"database/sql"
	"fmt"
)

// NullString represents a string that is nullable.  Custom serialization methods have been added
type NullString struct {
	sql.NullString
}

// MakeNullStringNull creates a NullString that represents a null value
func MakeNullStringNull() NullString {
	return NullString{NullString: sql.NullString{String: "", Valid: false}}
}

// MakeNullString creates a NullString that represents a non-null string value
func MakeNullString(str string) NullString {
	return NullString{NullString: sql.NullString{String: str, Valid: true}}
}

// MarshalJSON marshalls a NullString into JSON
func (nullString *NullString) MarshalJSON() ([]byte, error) {
	if !nullString.NullString.Valid {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf(`"%s"`, nullString.NullString.String)), nil
}
