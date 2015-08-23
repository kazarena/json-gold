package ld

import (
	"crypto/sha1"
	"sort"
	"strings"
)

// Normalize performs RDF normalization on the given JSON-LD input.
// dataset: the expanded JSON-LD object to normalize.
// Returns the normalized JSON-LD object.
func (api *JsonLdApi) Normalize(dataset *RDFDataset, opts *JsonLdOptions) (interface{}, error) {
	// create quads and map bnodes to their associated quads
	quads := make([]*Quad, 0)
	bnodes := make(map[string]interface{})
	for graphName, triples := range dataset.Graphs {
		if graphName == "@default" {
			graphName = ""
		}
		for _, quad := range triples {
			if graphName != "" {
				if strings.Index(graphName, "_:") == 0 {
					quad.Graph = NewBlankNode(graphName)
				} else {
					quad.Graph = NewIRI(graphName)
				}
			}

			quads = append(quads, quad)

			for _, attrNode := range []Node{quad.Subject, quad.Object, quad.Graph} {
				if attrNode != nil {
					if IsBlankNode(attrNode) {
						id := attrNode.GetValue()
						if _, hasID := bnodes[id]; !hasID {
							bnodes[id] = map[string]interface{}{
								"quads": make([]*Quad, 0),
							}
						}
						quadsList := bnodes[id].(map[string]interface{})["quads"].([]*Quad)
						bnodes[id].(map[string]interface{})["quads"] = append(quadsList, quad)
					}
				}
			}
		}
	}

	// mapping complete, start canonical naming
	normalizeUtils := NewNormalizeUtils(quads, bnodes, NewUniqueNamer("_:c14n"), opts.Format)

	return normalizeUtils.HashBlankNodes(GetKeys(bnodes))
}

// NormalizeUtils keeps the state of the Normalisation process
type NormalizeUtils struct {
	namer  *UniqueNamer
	bnodes map[string]interface{}
	quads  []*Quad
	format string
}

// NewNormalizeUtils creates a new instance of NormalizeUtils
func NewNormalizeUtils(quads []*Quad, bnodes map[string]interface{}, namer *UniqueNamer,
	format string) *NormalizeUtils {
	nu := &NormalizeUtils{
		format: format,
		quads:  quads,
		bnodes: bnodes,
		namer:  namer,
	}
	return nu
}

// HashBlankNodes generates unique and duplicate hashes for bnodes
func (nu *NormalizeUtils) HashBlankNodes(unnamed []string) (interface{}, error) {
	nextUnnamed := make([]string, 0)
	duplicates := make(map[string][]string)
	unique := make(map[string]string)

	// NOTE: not using the same structure as javascript here to avoid
	// possible stack overflows
	// hash quads for each unnamed bnode
	for hui := 0; ; hui++ {

		if hui == len(unnamed) {
			// done, name blank nodes
			named := false
			hashes := GetKeysString(unique)
			sort.Strings(hashes)

			for _, hash := range hashes {
				bnode := unique[hash]
				nu.namer.GetName(bnode)
				named = true
			}

			// continue to hash bnodes if a bnode was assigned a name
			if named {
				// this resets the initial variables, so it seems like it
				// has to go on the stack
				// but since this is the end of the function either way, it
				// might not have to
				// hashBlankNodes(unnamed);
				hui = -1
				unnamed = nextUnnamed
				nextUnnamed = make([]string, 0)
				duplicates = make(map[string][]string)
				unique = make(map[string]string)
				continue
			} else { // name the duplicate hash bnods
				// names duplicate hash bnodes
				// enumerate duplicate hash groups in sorted order
				hashes := make([]string, len(duplicates))
				i := 0
				for key := range duplicates {
					hashes[i] = key
					i++
				}
				sort.Strings(hashes)

				// process each group
				for pgi := 0; ; pgi++ {
					if pgi == len(hashes) {
						// done, create JSON-LD array
						// return createArray();
						normalized := make([]string, 0)

						// Note: At this point all bnodes in the set of RDF
						// quads have been
						// assigned canonical names, which have been stored
						// in the 'namer' object.
						// Here each quad is updated by assigning each of
						// its bnodes its new name
						// via the 'namer' object

						// update bnode names in each quad and serialize
						for _, quad := range nu.quads {
							for _, attrNode := range []Node{quad.Subject, quad.Object, quad.Graph} {
								if attrNode != nil {
									attrValue := attrNode.GetValue()
									if IsBlankNode(attrNode) && strings.Index(attrValue, "_:c14n") != 0 {
										bn := attrNode.(*BlankNode)
										bn.Attribute = nu.namer.GetName(attrValue)
									}
								}
							}

							var name string
							nameVal := quad.Graph
							if nameVal != nil {
								name = nameVal.(Node).GetValue()
							}
							normalized = append(normalized, toNQuad(quad, name, ""))
						}

						// sort normalized output
						sort.Strings(normalized)

						// handle output format
						if nu.format != "" {
							// TODO kazarena: review this condition
							if nu.format == "application/nquads" {
								rval := ""
								for _, n := range normalized {
									rval += n
								}
								return rval, nil
							} else {
								return nil, NewJsonLdError(UnknownFormat, nu.format)
							}
						}
						rval := ""
						for _, n := range normalized {
							rval += n
						}
						return ParseNQuads(rval)
					}

					// name each group member
					group := duplicates[hashes[pgi]]
					results := make([]*HashResult, 0)
					for n := 0; ; n++ {
						if n == len(group) {
							// name bnodes in hash order
							sort.Sort(ByHash(results))

							for _, r := range results {
								// name all bnodes in path namer in
								// key-entry order
								// Note: key-order is preserved in
								// javascript
								for _, key := range r.pathNamer.existingOrder {
									nu.namer.GetName(key)
								}
							}
							// processGroup(i+1);
							break
						} else {
							// skip already-named bnodes
							bnode := group[n]
							if nu.namer.IsNamed(bnode) {
								continue
							}

							// hash bnode paths
							pathNamer := NewUniqueNamer("_:b")
							pathNamer.GetName(bnode)

							result := hashPaths(bnode, nu.bnodes, nu.namer, pathNamer)
							results = append(results, result)
						}
					}
				}
			}
		}

		// hash unnamed bnode
		bnode := unnamed[hui]
		hash := hashQuads(bnode, nu.bnodes)

		// store hash as unique or a duplicate
		if dupVal, hasHash := duplicates[hash]; hasHash {
			duplicates[hash] = append(dupVal, bnode)
			nextUnnamed = append(nextUnnamed, bnode)
		} else if uniqueVal, hasHash := unique[hash]; hasHash {
			duplicates[hash] = []string{
				uniqueVal,
				bnode,
			}
			nextUnnamed = append(nextUnnamed, uniqueVal, bnode)
			delete(unique, hash)
		} else {
			unique[hash] = bnode
		}
	}
}

// HashResult
type HashResult struct {
	hash      string
	pathNamer *UniqueNamer
}

// ByHash helps sorting HashResult by hash
type ByHash []*HashResult

func (bh ByHash) Len() int {
	return len(bh)
}
func (bh ByHash) Swap(i, j int) {
	bh[i], bh[j] = bh[j], bh[i]
}
func (bh ByHash) Less(i, j int) bool {
	return bh[i].hash < bh[j].hash
}

// hashPaths produces a hash for the paths of adjacent bnodes for a bnode,
// incorporating all information about its subgraph of bnodes. This method
// will recursively pick adjacent bnode permutations that produce the
// lexicographically-least 'path' serializations.
func hashPaths(id string, bnodes map[string]interface{}, namer *UniqueNamer, pathNamer *UniqueNamer) *HashResult {
	//return nil

	// create SHA-1 digest
	md := sha1.New()

	groups := make(map[string][]string)
	groupHashes := make([]string, 0)
	quads := bnodes[id].(map[string]interface{})["quads"].([]*Quad)

	hpi := 0
	for ; ; hpi++ {
		if hpi == len(quads) {

			// done , hash groups
			groupHashes = make([]string, len(groups))
			i := 0
			for key := range groups {
				groupHashes[i] = key
				i++
			}
			sort.Strings(groupHashes)

			for hgi := 0; ; hgi++ {
				if hgi == len(groupHashes) {
					res := &HashResult{}
					res.hash = encodeHex(md.Sum(nil))
					res.pathNamer = pathNamer
					return res
				}

				// digest group hash
				groupHash := groupHashes[hgi]
				md.Write([]byte(groupHash))

				// choose a path and namer from the permutations
				chosenPath := ""
				var chosenNamer *UniqueNamer

				permutator := NewPermutator(groups[groupHash])
				for {
					contPermutation := false
					breakOut := false
					permutation := permutator.Next()
					pathNamerCopy := pathNamer.Clone()

					// build adjacent path
					path := ""
					recurse := make([]string, 0)
					for _, bnode := range permutation {
						// use canonical name if available
						if namer.IsNamed(bnode) {
							path += namer.GetName(bnode)
						} else {
							// recurse if bnode isn't named in the path
							// yet
							if !pathNamerCopy.IsNamed(bnode) {
								recurse = append(recurse, bnode)
							}
							path += pathNamerCopy.GetName(bnode)
						}

						// skip permutation if path is already >= chosen
						// path
						if chosenPath != "" && len(path) >= len(chosenPath) && path > chosenPath {
							if permutator.HasNext() {
								contPermutation = true
							} else {
								// digest chosen path and update namer
								md.Write([]byte(chosenPath))
								pathNamer = chosenNamer
								// hash the nextGroup
								breakOut = true
							}
							break
						}
					}

					// if we should do the next permutation
					if contPermutation {
						continue
					}
					// if we should stop processing this group
					if breakOut {
						break
					}

					// does the next recursion
					for nrn := 0; ; nrn++ {
						if nrn == len(recurse) {
							// return nextPermutation(false);
							if chosenPath == "" || path < chosenPath {
								chosenPath = path
								chosenNamer = pathNamerCopy
							}
							if !permutator.HasNext() {
								// digest chosen path and update namer
								md.Write([]byte(chosenPath))
								pathNamer = chosenNamer
								// hash the nextGroup
								breakOut = true
							}
							break
						}

						// do recursion
						bnode := recurse[nrn]
						result := hashPaths(bnode, bnodes, namer, pathNamerCopy)
						path += pathNamerCopy.GetName(bnode) + "<" + result.hash + ">"
						pathNamerCopy = result.pathNamer

						// skip permutation if path is already >= chosen
						// path
						if chosenPath != "" && len(path) >= len(chosenPath) && path > chosenPath {
							if !permutator.HasNext() {
								// digest chosen path and update namer
								md.Write([]byte(chosenPath))
								pathNamer = chosenNamer
								// hash the nextGroup
								breakOut = true
							}
							break
						}
						// do next recursion
					}

					// if we should stop processing this group
					if breakOut {
						break
					}
				}
			}
		}

		// get adjacent bnode
		quad := quads[hpi]
		bnode := getAdjacentBlankNodeName(quad.Subject, id)
		direction := ""
		if bnode != "" {
			// normal property
			direction = "p"
		} else {
			bnode = getAdjacentBlankNodeName(quad.Object, id)
			if bnode != "" {
				// reverse property
				direction = "r"
			}
		}

		if bnode != "" {
			// get bnode name (try canonical, path, then hash)
			name := ""
			if namer.IsNamed(bnode) {
				name = namer.GetName(bnode)
			} else if pathNamer.IsNamed(bnode) {
				name = pathNamer.GetName(bnode)
			} else {
				name = hashQuads(bnode, bnodes)
			}

			// hash direction, property, end bnode name/hash
			md1 := sha1.New()
			md1.Write([]byte(direction))
			md1.Write([]byte(quad.Predicate.GetValue()))
			md1.Write([]byte(name))
			groupHash := encodeHex(md1.Sum(nil))
			if groupVal, present := groups[groupHash]; present {
				groups[groupHash] = append(groupVal, bnode)
			} else {
				groups[groupHash] = []string{bnode}
			}
		}
	}
}

// hashQuads hashes all of the quads about a blank node.
// id: the ID of the bnode to hash quads for
// bnodes: the mapping of bnodes to quads
func hashQuads(id string, bnodes map[string]interface{}) string {
	// return cached hash
	v, _ := bnodes[id]
	idMap, _ := v.(map[string]interface{})
	if hashVal, hasHash := idMap["hash"]; hasHash {
		return hashVal.(string)
	}

	// serialize all of bnode's quads
	quads := idMap["quads"].([]*Quad)
	nquads := make([]string, 0)
	for _, quad := range quads {
		var name string
		graphVal := quad.Graph
		if graphVal != nil {
			name = graphVal.GetValue()
		}
		nquads = append(nquads, toNQuad(quad, name, id))
	}
	// sort serialized quads
	sort.Strings(nquads)
	// return hashed quads
	hash := sha1hash(nquads)
	idMap["hash"] = hash

	return hash
}

func sha1hash(nquads []string) string {
	h := sha1.New()
	for _, nquad := range nquads {
		h.Write([]byte(nquad))
	}
	return encodeHex(h.Sum(nil))
}

const hexDigit = "0123456789abcdef"

func encodeHex(data []byte) string {
	var buf = make([]byte, 0, len(data)*2)
	for _, b := range data {
		buf = append(buf, hexDigit[b>>4], hexDigit[b&0xf])
	}
	return string(buf)
}

// getAdjacentBlankNodeName is a helper function that gets the blank node name
// from an RDF quad node (subject or object). If the node is a blank node and
// its value does not match the given blank node ID, it will be returned.
func getAdjacentBlankNodeName(node Node, id string) string {
	nodeValue := node.GetValue()
	if IsBlankNode(node) && (nodeValue == "" || nodeValue != id) {
		return nodeValue
	}

	return ""
}

// Permutator
type Permutator struct {
	list []string
	done bool
	left map[string]bool
}

// NewPermutator creates a new instance of Permutator.
func NewPermutator(list []string) *Permutator {
	p := &Permutator{}
	p.list = make([]string, len(list))
	for i, elem := range list {
		p.list[i] = elem
	}
	sort.Strings(p.list)
	p.done = false
	p.left = make(map[string]bool, len(list))
	for _, i := range p.list {
		p.left[i] = true
	}

	return p
}

// HasNext returns true if there is another permutation.
func (p *Permutator) HasNext() bool {
	return !p.done
}

// Next gets the next permutation. Call HasNext() to ensure there is another one first.
func (p *Permutator) Next() []string {
	rval := make([]string, len(p.list))
	for i, elem := range p.list {
		rval[i] = elem
	}

	// Calculate the next permutation using Steinhaus-Johnson-Trotter
	// permutation algorithm

	// get largest mobile element k
	// (mobile: element is greater than the one it is looking at)
	k := ""
	pos := 0
	length := len(p.list)
	for i := 0; i < length; i++ {
		element := p.list[i]
		left := p.left[element]
		if (k == "" || element > k) &&
			((left && i > 0 && element > p.list[i-1]) || (!left && i < (length-1) && element > p.list[i+1])) {
			k = element
			pos = i
		}
	}

	// no more permutations
	if k == "" {
		p.done = true
	} else {
		// swap k and the element it is looking at
		var swap int
		if p.left[k] {
			swap = pos - 1
		} else {
			swap = pos + 1
		}
		p.list[pos] = p.list[swap]
		p.list[swap] = k

		// reverse the direction of all element larger than k
		for i := 0; i < length; i++ {
			if p.list[i] > k {
				p.left[p.list[i]] = !p.left[p.list[i]]
			}
		}
	}

	return rval
}
