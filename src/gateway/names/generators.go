package names

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
)

type Generator interface {
	Generate() string
}

type Integer struct {
	lowerBound int // inclusive
	upperBound int // non-inclusive
}

func NewInteger(lb int, ub int) *Integer {
	if lb > ub {
		panic("Illegal bounds provided - lower bound cannot be greater than upper bound")
	}
	return &Integer{lowerBound: lb, upperBound: ub}
}

func (i *Integer) Generate() string {
	random := rand.Intn(i.upperBound-i.lowerBound) + i.lowerBound
	return fmt.Sprintf("%d", random)
}

type Dictionary struct {
	terms []string
}

func NewDictionary(dictionary string) *Dictionary {
	bytes, err := Asset(fmt.Sprintf("%s.txt", dictionary))
	if err != nil {
		panic(fmt.Sprintf("Missing asset for dictionary %s.  Received error: %v", dictionary, err))
	}
	words := strings.Split(string(bytes), "\n")
	terms := []string{}
	for _, word := range words {
		trimmed := strings.TrimSpace(word)
		if trimmed == "" {
			continue
		}
		terms = append(terms, trimmed)
	}
	return &Dictionary{terms: terms}
}

func (d *Dictionary) Generate() string {
	random := rand.Intn(len(d.terms))
	return d.terms[random]
}

type Composite struct {
	generators []Generator
	separator  string
}

func (c *Composite) Generate() string {
	buf := bytes.NewBufferString("")
	for idx, gen := range c.generators {
		buf.WriteString(gen.Generate())
		if idx < len(c.generators)-1 {
			buf.WriteString(c.separator)
		}
	}
	return buf.String()
}
