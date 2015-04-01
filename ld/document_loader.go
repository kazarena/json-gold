package ld

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const (
	// An HTTP Accept header that prefers JSONLD.
	acceptHeader = "application/ld+json, application/json;q=0.9, application/javascript;q=0.5, text/javascript;q=0.5, text/plain;q=0.2, */*;q=0.1"

	// JSON-LD link header rel
	linkHeaderRel = "http://www.w3.org/ns/json-ld#context"
)

// RemoteDocument is a document retrieved from a remote source.
type RemoteDocument struct {
	DocumentURL string
	Document    interface{}
	ContextURL  string
}

// DocumentLoader knows how to load remote documents.
type DocumentLoader interface {
	LoadDocument(u string) (*RemoteDocument, error)
}

// DefaultDocumentLoader is a standard implementation of DocumentLoader
// which can retrieve documents via HTTP.
type DefaultDocumentLoader struct {
	httpClient *http.Client
}

// NewDefaultDocumentLoader creates a new instance of DefaultDocumentLoader
func NewDefaultDocumentLoader(httpClient *http.Client) *DefaultDocumentLoader {
	rval := &DefaultDocumentLoader{}

	if httpClient != nil {
		rval.httpClient = httpClient
	} else {
		rval.httpClient = &http.Client{}
	}
	return rval
}

// LoadDocument returns a RemoteDocument containing the contents of the JSON resource
// from the given URL.
func (dl *DefaultDocumentLoader) LoadDocument(u string) (*RemoteDocument, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, NewJsonLdError(LoadingDocumentFailed, err)
	}

	var documentBody []byte
	var finalURL, contextURL string

	protocol := parsedURL.Scheme
	if protocol != "http" && protocol != "https" {
		// Can't use the HTTP client for those!
		finalURL = u
		documentBody, err = ioutil.ReadFile(u)
	} else {

		req, err := http.NewRequest("GET", u, nil)
		// We prefer application/ld+json, but fallback to application/json
		// or whatever is available
		req.Header.Add("Accept", acceptHeader)

		res, err := dl.httpClient.Do(req)

		if err != nil {
			return nil, NewJsonLdError(LoadingDocumentFailed, err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, NewJsonLdError(LoadingDocumentFailed,
				fmt.Sprintf("Bad response status code: %d", res.StatusCode))
		}

		finalURL = res.Request.URL.String()

		contentType := res.Header.Get("Content-Type")
		linkHeader := res.Header.Get("Link")

		if len(linkHeader) > 0 && contentType != "application/ld+json" {
			header := ParseLinkHeader(linkHeader)[linkHeaderRel]
			if len(header) > 1 {
				return nil, NewJsonLdError(MultipleContextLinkHeaders, nil)
			} else if len(header) == 1 {
				contextURL = header[0]["target"]
			}
		}

		documentBody, err = ioutil.ReadAll(res.Body)
	}
	if err != nil {
		return nil, NewJsonLdError(LoadingDocumentFailed, err)
	}

	var document interface{}
	if len(documentBody) > 0 {
		err = json.Unmarshal(documentBody, &document)
	} else {
		document = ""
	}

	return &RemoteDocument{DocumentURL: finalURL, Document: document, ContextURL: contextURL}, nil
}

var rSplitOnComma = regexp.MustCompile("(?:<[^>]*?>|\"[^\"]*?\"|[^,])+")
var rLinkHeader = regexp.MustCompile("\\s*<([^>]*?)>\\s*(?:;\\s*(.*))?")
var rParams = regexp.MustCompile("(.*?)=(?:(?:\"([^\"]*?)\")|([^\"]*?))\\s*(?:(?:;\\s*)|$)")

// ParseLinkHeader parses a link header. The results will be keyed by the value of "rel".
//
// Link: <http://json-ld.org/contexts/person.jsonld>; \
//   rel="http://www.w3.org/ns/json-ld#context"; type="application/ld+json"
//
// Parses as: {
//   'http://www.w3.org/ns/json-ld#context': {
//     target: http://json-ld.org/contexts/person.jsonld,
//     rel:    http://www.w3.org/ns/json-ld#context
//   }
// }
//
// If there is more than one "rel" with the same IRI, then entries in the
// resulting map for that "rel" will be lists.
func ParseLinkHeader(header string) map[string][]map[string]string {

	rval := make(map[string][]map[string]string)

	// split on unbracketed/unquoted commas
	entries := rSplitOnComma.FindAllString(header, -1)
	if len(entries) == 0 {
		return rval
	}

	for _, entry := range entries {
		if !rLinkHeader.MatchString(entry) {
			continue
		}
		match := rLinkHeader.FindStringSubmatch(entry)

		result := map[string]string{
			"target": match[1],
		}
		params := match[2]
		matches := rParams.FindAllStringSubmatch(params, -1)
		for _, match := range matches {
			if match[2] == "" {
				result[match[1]] = match[3]
			} else {
				result[match[1]] = match[2]
			}
		}
		rel := result["rel"]
		relVal, hasRel := rval[rel]
		if hasRel {
			rval[rel] = append(relVal, result)
		} else {
			rval[rel] = []map[string]string{result}
		}
	}
	return rval
}
