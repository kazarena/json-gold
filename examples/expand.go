// +build ignore

package main

import (
	"github.com/kazarena/json-gold/ld"
	"log"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	expanded, err := proc.Expand("http://json-ld.org/test-suite/tests/expand-0002-in.jsonld", options)
	if err != nil {
		log.Println("Error when expanding JSON-LD document:", err)
		return
	}

	ld.PrintDocument("JSON-LD expansion succeeded", expanded)
}
