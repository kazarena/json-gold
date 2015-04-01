package ld_test

import (
	. "github.com/kazarena/json-gold/ld"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadDocument(t *testing.T) {
	dl := NewDefaultDocumentLoader(nil)

	rd, _ := dl.LoadDocument("testdata/expand-0002-in.jsonld")

	assert.Equal(t, "t1", rd.Document.(map[string]interface{})["@type"])
}

func TestParseLinkHeader(t *testing.T) {
	rval := ParseLinkHeader("<remote-doc-0010-context.jsonld>; rel=\"http://www.w3.org/ns/json-ld#context\"")

	assert.Equal(
		t,
		map[string][]map[string]string{
			"http://www.w3.org/ns/json-ld#context": []map[string]string{map[string]string{
				"target": "remote-doc-0010-context.jsonld",
				"rel":    "http://www.w3.org/ns/json-ld#context",
			}},
		},
		rval,
	)
}
