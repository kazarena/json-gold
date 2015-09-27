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

	doc := make(map[string]interface{})
	doc["@context"] = "http://schema.org/"
	doc["@type"] = "Person"
	doc["name"] = "Jane Doe"
	doc["jobTitle"] = "Professor"
	doc["telephone"] = "(425) 123-4567"
	doc["url"] = "http://www.janedoe.com"

	expanded, err = proc.Expand(doc, options)
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("JSON-LD expansion succeeded", expanded)
}
