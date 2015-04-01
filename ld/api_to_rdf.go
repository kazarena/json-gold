package ld

// ToRDF adds RDF triples for each graph in the current node map to an RDF dataset.
func (api *JsonLdApi) ToRDF(input interface{}, opts *JsonLdOptions) (*RDFDataset, error) {
	idGen := NewBlankNodeIDGenerator()

	nodeMap := make(map[string]interface{})
	nodeMap["@default"] = make(map[string]interface{})
	api.GenerateNodeMap(input, nodeMap, "@default", nil, "", nil, idGen)

	dataset := NewRDFDataset()

	for graphName, graphVal := range nodeMap {
		// 4.1)
		if IsRelativeIri(graphName) {
			continue
		}
		graph := graphVal.(map[string]interface{})
		dataset.GraphToRDF(graphName, graph, idGen, opts.ProduceGeneralizedRdf)
	}

	return dataset, nil
}
