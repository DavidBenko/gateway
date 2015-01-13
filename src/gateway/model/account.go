package model

// Account represents a single tenant in multi-tenant deployment.
type Account struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Validate validates the model.
func (a *Account) Validate() Errors {
	errors := make(Errors)
	if a.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}
