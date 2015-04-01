package ld_test

import (
	. "github.com/kazarena/json-gold/ld"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonLdUrl(t *testing.T) {
	parsedUrl := ParseURL("http://www.example.com")

	assert.Equal(t, "http:", parsedUrl.Protocol)
	assert.Equal(t, "www.example.com", parsedUrl.Host)
}

func TestRemoveBase(t *testing.T) {
	result := RemoveBase(
		"http://json-ld.org/test-suite/tests/compact-0045-in.jsonld",
		"http://json-ld.org/test-suite/parent-node",
	)
	assert.Equal(t, "../parent-node", result)

	result = RemoveBase(
		"http://example.com/",
		"http://example.com/relative-url",
	)
	assert.Equal(t, "relative-url", result)

	result = RemoveBase(
		"http://json-ld.org/test-suite/tests/compact-0066-in.jsonld",
		"http://json-ld.org/test-suite/",
	)
	assert.Equal(t, "../", result)
}
