package ld

// TurtleRDFSerializer parses and serializes Turtle data.
type TurtleRDFSerializer struct {
}

// Parse Turtle from string into an RDFDataset
func (s *TurtleRDFSerializer) Parse(input interface{}) (*RDFDataset, error) {
	return nil, NewJsonLdError(NotImplemented, "Turtle not supported")
}

// Serialize an RDFDataset into a Turtle string.
func (s *TurtleRDFSerializer) Serialize(dataset *RDFDataset) (interface{}, error) {
	return nil, NewJsonLdError(NotImplemented, "Turtle not supported")
}
