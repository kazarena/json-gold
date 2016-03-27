// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	doc := map[string]interface{}{
		"@context": map[string]interface{}{
			"ex": "http://example.org/vocab#",
		},
		"@id":   "http://example.org/test#example",
		"@type": "ex:Foo",
		"ex:embed": map[string]interface{}{
			"@type": "ex:Bar",
		},
	}

	normalizedTriples, err := proc.Normalize(doc, options)
	if err != nil {
		panic(err)
	}

	print(normalizedTriples.(string))
}
