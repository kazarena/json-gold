// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	doc := map[string]interface{}{
		"@context": []interface{}{
			map[string]interface{}{
				"name": "http://xmlns.com/foaf/0.1/name",
				"homepage": map[string]interface{}{
					"@id":   "http://xmlns.com/foaf/0.1/homepage",
					"@type": "@id",
				},
			},
			map[string]interface{}{
				"ical": "http://www.w3.org/2002/12/cal/ical#",
			},
		},
		"@id":           "http://example.com/speakers#Alice",
		"name":          "Alice",
		"homepage":      "http://xkcd.com/177/",
		"ical:summary":  "Alice Talk",
		"ical:location": "Lyon Convention Centre, Lyon, France",
	}

	flattenedDoc, err := proc.Flatten(doc, nil, options)
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("JSON-LD flattening succeeded", flattenedDoc)
}
