package model

// Errors represents API-serializable validation errors.
type Errors map[string][]string

func (e Errors) add(name, message string) {
	e[name] = append(e[name], message)
}

// Empty reports if there are no errors.
func (e Errors) Empty() bool {
	return len(e) == 0
}
