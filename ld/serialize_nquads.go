package ld

import (
	"regexp"
	"sort"
	"strings"
)

// NQuadRDFSerializer parses and serializes N-Quads.
type NQuadRDFSerializer struct {
}

// Parse N-Quads from string into an RDFDataset.
func (s *NQuadRDFSerializer) Parse(input interface{}) (*RDFDataset, error) {
	if inputStr, isString := input.(string); isString {
		return ParseNQuads(inputStr)
	}

	return nil, NewJsonLdError(InvalidInput, "NQuad Parser expected string input")
}

// Serialize an RDFDataset into N-Quad string.
func (s *NQuadRDFSerializer) Serialize(dataset *RDFDataset) (interface{}, error) {
	quads := make([]string, 0)
	for graphName, triples := range dataset.Graphs {
		if graphName == "@default" {
			graphName = ""
		}
		for _, triple := range triples {
			quads = append(quads, toNQuad(triple, graphName, ""))
		}
	}
	sort.Strings(quads)
	rval := ""
	for _, quad := range quads {
		rval += quad
	}
	return rval, nil
}

func toNQuad(triple *Quad, graphName string, bnode string) string {

	s := triple.Subject
	p := triple.Predicate
	o := triple.Object

	quad := ""

	// subject is an IRI or bnode
	if IsIRI(s) {
		quad += "<" + escape(s.GetValue()) + ">"
	} else if bnode != "" {
		// normalization mode
		if bnode == s.GetValue() {
			quad += "_:a"
		} else {
			quad += "_:z"
		}
	} else {
		// normal mode
		quad += s.GetValue()
	}

	if IsIRI(p) {
		quad += " <" + escape(p.GetValue()) + "> "
	} else {
		// otherwise it must be a bnode (TODO: can we only allow this if the
		// flag is set in options?)
		quad += " " + escape(p.GetValue()) + " "
	}

	// object is IRI, bnode or literal
	if IsIRI(o) {
		quad += "<" + escape(o.GetValue()) + ">"
	} else if IsBlankNode(o) {
		// normalization mode
		if bnode != "" {
			if bnode == o.GetValue() {
				quad += "_:a"
			} else {
				quad += "_:z"
			}
		} else {
			// normal mode
			quad += o.GetValue()
		}
	} else {
		literal := o.(*Literal)
		escaped := escape(literal.GetValue())
		quad += "\"" + escaped + "\""
		if literal.Datatype == RDFLangString {
			quad += "@" + literal.Language
		} else if literal.Datatype != XSDString {
			quad += "^^<" + escape(literal.Datatype) + ">"
		}
	}

	// graph
	if graphName != "" {
		if strings.Index(graphName, "_:") != 0 {
			quad += " <" + escape(graphName) + ">"
		} else if bnode != "" {
			quad += " _:g"
		} else {
			quad += " " + graphName
		}
	}

	quad += " .\n"

	return quad
}

func unescape(str string) string {
	str = strings.Replace(str, "\\\\", "\\", -1)
	str = strings.Replace(str, "\\\"", "\"", -1)
	str = strings.Replace(str, "\\n", "\n", -1)
	str = strings.Replace(str, "\\r", "\r", -1)
	str = strings.Replace(str, "\\t", "\t", -1)
	return str
}

func escape(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	str = strings.Replace(str, "\n", "\\n", -1)
	str = strings.Replace(str, "\r", "\\r", -1)
	str = strings.Replace(str, "\t", "\\t", -1)
	return str
}

const (
	wso      = "[ \\t]*"
	iri      = "(?:<([^>]*)>)"
	bnode    = "(_:(?:[A-Za-z][A-Za-z0-9]*))"
	plain    = "\"([^\"\\\\]*(?:\\\\.[^\"\\\\]*)*)\""
	datatype = "(?:\\^\\^" + iri + ")"
	language = "(?:@([a-z]+(?:-[a-zA-Z0-9]+)*))"
	literal  = "(?:" + plain + "(?:" + datatype + "|" + language + ")?)"
	ws       = "[ \\t]+"

	subject  = "(?:" + iri + "|" + bnode + ")" + ws
	property = iri + ws
	object   = "(?:" + iri + "|" + bnode + "|" + literal + ")" + wso
	graph    = "(?:\\.|(?:(?:" + iri + "|" + bnode + ")" + wso + "\\.))"
)

var regexWSO = regexp.MustCompile(wso)

var regexEOLN = regexp.MustCompile("(?:\\r\\n)|(?:\\n)|(?:\\r)")

var regexEmpty = regexp.MustCompile("^" + wso + "$")

// define quad part regexes

var regexSubject = regexp.MustCompile("(?:" + iri + "|" + bnode + ")" + ws)
var regexProperty = regexp.MustCompile(iri + ws)
var regexObject = regexp.MustCompile("(?:" + iri + "|" + bnode + "|" + literal + ")" + wso)
var regexGraph = regexp.MustCompile("(?:\\.|(?:(?:" + iri + "|" + bnode + ")" + wso + "\\.))")

// full quad regex

var regexQuad = regexp.MustCompile("^" + wso + subject + property + object + graph + wso + "$")

// ParseNQuads parses RDF in the form of N-Quads.
func ParseNQuads(input string) (*RDFDataset, error) {

	// build RDF dataset
	dataset := NewRDFDataset()

	// split N-Quad input into lines
	lines := regexEOLN.Split(input, -1)
	lineNumber := 0
	for _, line := range lines {
		lineNumber++

		// skip empty lines
		if regexEmpty.MatchString(line) {
			continue
		}

		// parse quad
		var match []string
		if regexQuad.MatchString(line) {
			match = regexQuad.FindStringSubmatch(line)
		} else {
			return nil, NewJsonLdError(SyntaxError, "Error while parsing N-Quads; invalid quad. line:"+string(lineNumber))
		}

		// get subject
		var subject Node
		if match[1] != "" {
			subject = NewIRI(unescape(match[1]))
		} else {
			subject = NewBlankNode(unescape(match[2]))
		}

		// get predicate
		predicate := NewIRI(unescape(match[3]))

		// get object
		var object Node
		if match[4] != "" {
			object = NewIRI(unescape(match[4]))
		} else if match[5] != "" {
			object = NewBlankNode(unescape(match[5]))
		} else {
			language := unescape(match[8])
			var datatype string
			if match[7] != "" {
				datatype = unescape(match[7])
			} else if match[8] != "" {
				datatype = RDFLangString
			} else {
				datatype = XSDString
			}
			unescaped := unescape(match[6])
			object = NewLiteral(unescaped, datatype, language)
		}

		// get graph name ('@default' is used for the default graph)
		name := "@default"
		if match[9] != "" {
			name = unescape(match[9])
		} else if match[10] != "" {
			name = unescape(match[10])
		}

		triple := NewQuad(subject, predicate, object, name)

		// initialise graph in dataset
		triples, present := dataset.Graphs[name]
		if !present {
			dataset.Graphs[name] = []*Quad{triple}
		} else {
			// add triple if unique to its graph
			containsTriple := false
			for _, elem := range triples {
				if triple.Equal(elem) {
					containsTriple = true
					break
				}
			}
			if !containsTriple {
				dataset.Graphs[name] = append(triples, triple)
			}
		}
	}

	return dataset, nil
}
