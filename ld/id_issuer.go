package ld

import (
	"fmt"
)

// IdentifierIssuer issues unique identifiers, keeping track of any previously issued identifiers.
type IdentifierIssuer struct {
	prefix        string
	counter       int
	existing      map[string]string
	existingOrder []string
}

// NewIdentifierIssuer creates and returns a new IdentifierIssuer.
func NewIdentifierIssuer(prefix string) *IdentifierIssuer {
	return &IdentifierIssuer{
		prefix:        prefix,
		counter:       0,
		existing:      make(map[string]string),
		existingOrder: make([]string, 0),
	}
}

// Clone copies this IdentifierIssuer.
func (ii *IdentifierIssuer) Clone() *IdentifierIssuer {
	copy := &IdentifierIssuer{
		prefix:        ii.prefix,
		counter:       ii.counter,
		existing:      make(map[string]string, len(ii.existing)),
		existingOrder: make([]string, len(ii.existingOrder)),
	}
	i := 0
	for k, v := range ii.existing {
		copy.existing[k] = v
		copy.existingOrder[i] = ii.existingOrder[i]
		i++
	}

	return copy
}

// GetId Gets the new identifier for the given old identifier, where if no old
// identifier is given a new identifier will be generated.
func (ii *IdentifierIssuer) GetId(oldId string) string {
	if oldId != "" {
		// return existing old identifier
		if ex, present := ii.existing[oldId]; present {
			return ex
		}
	}

	id := ii.prefix + fmt.Sprintf("%d", ii.counter)
	ii.counter++

	if oldId != "" {
		ii.existing[oldId] = id
		ii.existingOrder = append(ii.existingOrder, oldId)
	}

	return id
}

// HasId returns True if the given old identifier has already been assigned a new identifier.
func (ii *IdentifierIssuer) HasId(oldId string) bool {
	_, hasKey := ii.existing[oldId]
	return hasKey
}
