package ld

const (
	JsonLd_1_0       = "json-ld-1.0"
	JsonLd_1_1       = "json-ld-1.1"
	JsonLd_1_1_Frame = "json-ld-1.1-expand-frame"
)

// JsonLdOptions type as specified in the JSON-LD-API specification:
// http://www.w3.org/TR/json-ld-api/#the-jsonldoptions-type
type JsonLdOptions struct {

	// Base options: http://www.w3.org/TR/json-ld-api/#idl-def-JsonLdOptions

	// http://www.w3.org/TR/json-ld-api/#widl-JsonLdOptions-base
	Base string
	// http://www.w3.org/TR/json-ld-api/#widl-JsonLdOptions-compactArrays
	CompactArrays bool
	// http://www.w3.org/TR/json-ld-api/#widl-JsonLdOptions-expandContext
	ExpandContext interface{}
	// http://www.w3.org/TR/json-ld-api/#widl-JsonLdOptions-processingMode
	ProcessingMode string
	// http://www.w3.org/TR/json-ld-api/#widl-JsonLdOptions-documentLoader
	DocumentLoader DocumentLoader

	// Frame options: http://json-ld.org/spec/latest/json-ld-framing/

	Embed       bool
	Explicit    bool
	OmitDefault bool

	// RDF conversion options: http://www.w3.org/TR/json-ld-api/#serialize-rdf-as-json-ld-algorithm

	UseRdfType            bool
	UseNativeTypes        bool
	ProduceGeneralizedRdf bool

	// The following properties aren't in the spec

	InputFormat   string
	Format        string
	Algorithm     string
	UseNamespaces bool
	OutputForm    string
}

// NewJsonLdOptions creates and returns new instance of JsonLdOptions with the given base.
func NewJsonLdOptions(base string) *JsonLdOptions {
	return &JsonLdOptions{
		Base:                  base,
		CompactArrays:         true,
		ProcessingMode:        JsonLd_1_0,
		DocumentLoader:        NewDefaultDocumentLoader(nil),
		Embed:                 true,
		Explicit:              false,
		OmitDefault:           false,
		UseRdfType:            false,
		UseNativeTypes:        false,
		ProduceGeneralizedRdf: false,
		InputFormat:           "",
		Format:                "",
		Algorithm:             "URGNA2012",
		UseNamespaces:         false,
		OutputForm:            "",
	}
}
