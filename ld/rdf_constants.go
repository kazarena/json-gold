package ld

const (
	RDFSyntaxNS string = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	RDFSchemaNS string = "http://www.w3.org/2000/01/rdf-schema#"
	XSDNS       string = "http://www.w3.org/2001/XMLSchema#"

	XSDAnyType string = XSDNS + "anyType"
	XSDBoolean string = XSDNS + "boolean"
	XSDDouble  string = XSDNS + "double"
	XSDInteger string = XSDNS + "integer"
	XSDFloat   string = XSDNS + "float"
	XSDDecimal string = XSDNS + "decimal"
	XSDAnyURI  string = XSDNS + "anyURI"
	XSDString  string = XSDNS + "string"

	RDFType         string = RDFSyntaxNS + "type"
	RDFFirst        string = RDFSyntaxNS + "first"
	RDFRest         string = RDFSyntaxNS + "rest"
	RDFNil          string = RDFSyntaxNS + "nil"
	RDFPlainLiteral string = RDFSyntaxNS + "PlainLiteral"
	RDFXMLLiteral   string = RDFSyntaxNS + "XMLLiteral"
	RDFObject       string = RDFSyntaxNS + "object"
	RDFLangString   string = RDFSyntaxNS + "langString"
	RDFList         string = RDFSyntaxNS + "List"
)
