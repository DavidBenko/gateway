package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// NullString represents a string that is nullable.  Custom serialization methods have been added
type NullString struct {
	sql.NullString
}

// NullInt64 represents an int64 that is nullable.  Custom serialization methods have been added
type NullInt64 struct {
	sql.NullInt64
}

// MakeNullStringNull creates a NullString that represents a null value
func MakeNullStringNull() NullString {
	return NullString{NullString: sql.NullString{String: "", Valid: false}}
}

// MakeNullString creates a NullString that represents a non-null string value
func MakeNullString(str string) NullString {
	return NullString{NullString: sql.NullString{String: str, Valid: true}}
}

// MakeNullInt64Null creates a NullInt64 that represents a null value
func MakeNullInt64Null() NullInt64 {
	return NullInt64{NullInt64: sql.NullInt64{Int64: int64(0), Valid: false}}
}

// MakeNullString creates a NullString that represents a non-null string value
func MakeNullInt64(value int64) NullInt64 {
	return NullInt64{NullInt64: sql.NullInt64{Int64: value, Valid: true}}
}

// MarshalJSON marshalls a NullString into JSON
func (nullString *NullString) MarshalJSON() ([]byte, error) {
	if !nullString.NullString.Valid {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf(`"%s"`, nullString.NullString.String)), nil
}

func (nullString *NullString) UnmarshalJSON(data []byte) error {
	str := string(data)

	if str == "null" {
		nullString.NullString.Valid = false
		return nil
	}

	if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
		nullString.NullString.String = str[1 : len(str)-1]
		nullString.NullString.Valid = true
		return nil
	}

	return fmt.Errorf("String was not in expected format: %v", string(data))
}

// MarshalJSON marshalls a NullInt64 into JSON
func (nullInt64 *NullInt64) MarshalJSON() ([]byte, error) {
	if !nullInt64.NullInt64.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.Itoa(int(nullInt64.NullInt64.Int64))), nil
}

func (nullInt64 *NullInt64) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		nullInt64.NullInt64.Valid = false
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("Invalid NullINt64: %v", s)
	}
	nullInt64.NullInt64.Int64 = int64(v)
	nullInt64.NullInt64.Valid = true
	return nil
}
