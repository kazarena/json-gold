package ld

import (
	"fmt"
)

// UniqueNamer issues unique names, keeping track of any previously issued names.
type UniqueNamer struct {
	prefix        string
	counter       int
	existing      map[string]string
	existingOrder []string
}

// NewUniqueNamer creates and returns a new UniqueNamer.
func NewUniqueNamer(prefix string) *UniqueNamer {
	return &UniqueNamer{
		prefix:        prefix,
		counter:       0,
		existing:      make(map[string]string),
		existingOrder: make([]string, 0),
	}
}

// Clone copies this UniqueNamer.
func (un *UniqueNamer) Clone() *UniqueNamer {
	copy := &UniqueNamer{
		prefix:        un.prefix,
		counter:       un.counter,
		existing:      make(map[string]string, len(un.existing)),
		existingOrder: make([]string, len(un.existingOrder)),
	}
	i := 0
	for k, v := range un.existing {
		copy.existing[k] = v
		copy.existingOrder[i] = un.existingOrder[i]
		i++
	}

	return copy
}

// GetName gets the new name for the given old name, where if no old name
// is given a new name will be generated.
func (un *UniqueNamer) GetName(oldName string) string {
	if oldName != "" {
		if ex, present := un.existing[oldName]; present {
			return ex
		}
	}

	name := un.prefix + fmt.Sprintf("%d", un.counter)
	un.counter++

	if oldName != "" {
		un.existing[oldName] = name
		un.existingOrder = append(un.existingOrder, oldName)
	}

	return name
}

// IsNamed returns true if there was already a name created for the given old name.
func (un *UniqueNamer) IsNamed(oldName string) bool {
	_, hasKey := un.existing[oldName]
	return hasKey
}
