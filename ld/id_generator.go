package ld

import (
	"fmt"
)

// BlankNodeIDGenerator generates IDs for blank nodes
// as specified in http://www.w3.org/TR/json-ld-api/#generate-blank-node-identifier
type BlankNodeIDGenerator struct {
	blankNodeIdentifierMap map[string]string
	blankNodeCounter       int

	nodeMap map[string]interface{}
}

// NewBlankNodeIDGenerator creates and returns a new instance of BlankNodeIDGenerator
func NewBlankNodeIDGenerator() *BlankNodeIDGenerator {
	return &BlankNodeIDGenerator{
		blankNodeIdentifierMap: make(map[string]string),
		blankNodeCounter:       0,
	}
}

// GenerateBlankNodeIdentifier generates a blank node identifier for the given key
// using the algorithm specified in: http://www.w3.org/TR/json-ld-api/#generate-blank-node-identifier
//
// id: The id, or an empty string to generate a fresh, unused, blank node identifier.
//
// Returns a blank node identifier based on id if it was not an empty string,
// or a fresh, unused, blank node identifier if it was an empty string.
func (bnig *BlankNodeIDGenerator) GenerateBlankNodeIdentifier(id string) string {
	if id != "" {
		if val, hasID := bnig.blankNodeIdentifierMap[id]; hasID {
			return val
		}
	}
	bnid := fmt.Sprintf("_:b%v", bnig.blankNodeCounter)
	bnig.blankNodeCounter++
	if id != "" {
		bnig.blankNodeIdentifierMap[id] = bnid
	}
	return bnid
}
