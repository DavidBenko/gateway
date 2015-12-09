package http

type AuthType int

const (
	AuthTypeUnknown AuthType = iota
	AuthTypeSite
	AuthTypeAdmin
	AuthTypeUser
)
