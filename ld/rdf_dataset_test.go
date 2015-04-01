package ld_test

import (
	. "github.com/kazarena/json-gold/ld"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCanonicalDouble(t *testing.T) {
	assert.Equal(t, "5.3E0", GetCanonicalDouble(5.3))
}
