package names

var adjGen = NewDictionary("adjectives")
var nounGen = NewDictionary("nouns")
var intGen = NewInteger(1000, 10000)

var gens = []Generator{adjGen, nounGen, intGen}
var compGen = Composite{generators: gens, separator: "-"}

func GenerateHostName() string {
	return compGen.Generate()
}
