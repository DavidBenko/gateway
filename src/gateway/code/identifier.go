package code

import "regexp"

var variableRx = regexp.MustCompile(`(?i)^[A-Z_][0-9A-Z_]*$`)

func IsValidVariableIdentifier(word string) bool {
	return variableRx.MatchString(word)
}
