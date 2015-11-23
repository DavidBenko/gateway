package names

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestGenerateInteger(t *testing.T) {
	lb := 1000
	ub := 10000
	generatorIterations := 10000

	intGen := NewInteger(lb, ub)
	for i := 0; i < generatorIterations; i++ {
		generated, _ := strconv.Atoi(intGen.Generate())
		if generated < lb || generated > ub {
			t.Errorf("Integer generator generated value greater than upper bound %d", ub)
		}
	}
}

func TestGenerateDictionaryWord(t *testing.T) {
	generatorIterations := 10000
	dictGen := NewDictionary("adjectives")

	for i := 0; i < generatorIterations; i++ {
		generated := dictGen.Generate()
		if strings.TrimSpace(generated) == "" {
			t.Errorf("Generator returned empty word %s", generated)
			t.FailNow()
		}
	}
}

func TestGenerateComposite(t *testing.T) {
	lb := 1000
	ub := 10000
	generatorIterations := 10000

	adjGen := NewDictionary("adjectives")
	nounGen := NewDictionary("nouns")
	intGen := NewInteger(lb, ub)

	gens := []Generator{adjGen, nounGen, intGen}
	compGen := Composite{generators: gens, separator: "-"}

	for i := 0; i < generatorIterations; i++ {
		generated := compGen.Generate()
		if strings.TrimSpace(generated) == "" {
			t.Errorf("Generator returned empty word %s", generated)
			t.FailNow()
		}
		fmt.Println(generated)
	}
}
