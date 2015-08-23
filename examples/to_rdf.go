// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
	"log"
	"os"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/nquads"

	// this JSON-LD document was taken from http://json-ld.org/test-suite/tests/toRdf-0028-in.jsonld
	doc := map[string]interface{}{
		"@context": map[string]interface{}{
			"sec":        "http://purl.org/security#",
			"xsd":        "http://www.w3.org/2001/XMLSchema#",
			"rdf":        "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"dc":         "http://purl.org/dc/terms/",
			"sec:signer": map[string]interface{}{"@type": "@id"},
			"dc:created": map[string]interface{}{"@type": "xsd:dateTime"},
		},
		"@id":                "http://example.org/sig1",
		"@type":              []interface{}{"rdf:Graph", "sec:SignedGraph"},
		"dc:created":         "2011-09-23T20:21:34Z",
		"sec:signer":         "http://payswarm.example.com/i/john/keys/5",
		"sec:signatureValue": "OGQzNGVkMzVm4NTIyZTkZDYMmMzQzNmExMgoYzI43Q3ODIyOWM32NjI=",
		"@graph": map[string]interface{}{
			"@id":      "http://example.org/fact1",
			"dc:title": "Hello World!",
		},
	}
	triples, err := proc.ToRDF(doc, options)
	if err != nil {
		log.Println("Error when transforming JSON-LD document to RDF:", err)
		return
	}

	os.Stdout.WriteString(triples.(string))
}
