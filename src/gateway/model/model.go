package model

type Typed interface {
	GetType() string
	SetType(t string)
}
