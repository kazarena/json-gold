// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	doc := map[string]interface{}{
		"@id": "http://example.org/test#book",
		"http://example.org/vocab#contains": map[string]interface{}{
			"@id": "http://example.org/test#chapter",
		},
		"http://purl.org/dc/elements/1.1/title": "Title",
	}

	context := map[string]interface{}{
		"@context": map[string]interface{}{
			"dc": "http://purl.org/dc/elements/1.1/",
			"ex": "http://example.org/vocab#",
			"ex:contains": map[string]interface{}{
				"@type": "@id",
			},
		},
	}

	compactedDoc, err := proc.Compact(doc, context, options)
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("JSON-LD compation succeeded", compactedDoc)
}
