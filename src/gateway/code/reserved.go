package code

import "strings"

var ReservedWords = []string{
	// JavaScript
	"break", "case", "catch", "continue", "debugger", "default", "delete",
	"do", "else", "finally", "for", "function", "if", "in", "instanceof", "new",
	"return", "switch", "this", "throw", "try", "typeof", "var", "void", "while",
	"with",

	// Reserved for future
	"class", "const", "enum", "export", "extends", "import", "super",
	"implements", "interface", "let", "package", "private", "protected", "public",
	"static", "yield",

	// Gateway Specific
	"request", "response", "session", "env",
}

func IsReserved(word string) bool {
	for _, reserved := range ReservedWords {
		if strings.ToLower(word) == reserved {
			return true
		}
	}
	return false
}
