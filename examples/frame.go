// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	doc := map[string]interface{}{
		"@context": map[string]interface{}{
			"dc":          "http://purl.org/dc/elements/1.1/",
			"ex":          "http://example.org/vocab#",
			"ex:contains": map[string]interface{}{"@type": "@id"},
		},
		"@graph": []interface{}{
			map[string]interface{}{
				"@id":         "http://example.org/test/#library",
				"@type":       "ex:Library",
				"ex:contains": "http://example.org/test#book",
			},
			map[string]interface{}{
				"@id":            "http://example.org/test#book",
				"@type":          "ex:Book",
				"dc:contributor": "Writer",
				"dc:title":       "My Book",
				"ex:contains":    "http://example.org/test#chapter",
			},
			map[string]interface{}{
				"@id":            "http://example.org/test#chapter",
				"@type":          "ex:Chapter",
				"dc:description": "Fun",
				"dc:title":       "Chapter One",
			},
		},
	}

	frame := map[string]interface{}{
		"@context": map[string]interface{}{
			"dc": "http://purl.org/dc/elements/1.1/",
			"ex": "http://example.org/vocab#",
		},
		"@type": "ex:Library",
		"ex:contains": map[string]interface{}{
			"@type": "ex:Book",
			"ex:contains": map[string]interface{}{
				"@type": "ex:Chapter",
			},
		},
	}

	framedDoc, err := proc.Frame(doc, frame, options)
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("JSON-LD framing succeeded", framedDoc)
}
