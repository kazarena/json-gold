// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
	"log"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	// expanding remote document

	expanded, err := proc.Expand("http://json-ld.org/test-suite/tests/expand-0002-in.jsonld", options)
	if err != nil {
		log.Println("Error when expanding JSON-LD document:", err)
		return
	}

	ld.PrintDocument("JSON-LD expansion succeeded", expanded)

	// expanding in-memory document

	doc := map[string]interface{}{
		"@context":  "http://schema.org/",
		"@type":     "Person",
		"name":      "Jane Doe",
		"jobTitle":  "Professor",
		"telephone": "(425) 123-4567",
		"url":       "http://www.janedoe.com",
	}

	expanded, err = proc.Expand(doc, options)
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("JSON-LD expansion succeeded", expanded)
}
