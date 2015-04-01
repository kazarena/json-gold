package ld_test

import (
	"encoding/json"
	. "github.com/kazarena/json-gold/ld"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockServer uses httptest package to mock live HTTP calls.
type MockServer struct {
	Base       string
	TestFolder string

	ContentType string
	HttpLink    []string
	HttpStatus  int
	RedirectTo  string

	server *httptest.Server

	DocumentLoader DocumentLoader
}

// NewMockServer creates a new instance of MockServer.
func NewMockServer(base string, testFolder string) *MockServer {

	mockServer := &MockServer{
		Base:       base,
		TestFolder: testFolder,
	}

	mockServer.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mockServer.HttpStatus != 0 {
			// must be a redirect
			w.Header().Set("Location", mockServer.Base+mockServer.RedirectTo)
			w.WriteHeader(mockServer.HttpStatus)
		} else {
			u := r.URL.String()

			if strings.HasPrefix(u, mockServer.Base) {
				contentType := mockServer.ContentType
				if contentType == "" {
					if strings.HasSuffix(u, ".jsonld") {
						contentType = "application/ld+json"
					} else {
						contentType = "application/json"
					}
				}

				fileName := filepath.Join(mockServer.TestFolder, u[len(mockServer.Base):len(u)])
				inputBytes, err := ioutil.ReadFile(fileName)
				if err == nil {
					w.Header().Set("Content-Type", contentType)
					if mockServer.HttpLink != nil {
						w.Header().Set("Link", strings.Join(mockServer.HttpLink, ", "))
					}
					w.WriteHeader(http.StatusOK)
					w.Write(inputBytes)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}

		}

		// reset the context for the second call so that it succeeds.
		// currently there are no tests where it needs to work in a different way
		mockServer.HttpStatus = 0
		mockServer.HttpLink = nil
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(mockServer.server.URL)
		},
	}

	mockServer.DocumentLoader = NewDefaultDocumentLoader(&http.Client{Transport: transport})

	return mockServer
}

func (ms *MockServer) SetExpectedBehaviour(contentType string, httpLink []string, httpStatus int, redirectTo string) {
	ms.ContentType = contentType
	ms.HttpLink = httpLink
	ms.HttpStatus = httpStatus
	ms.RedirectTo = redirectTo
}

func (ms *MockServer) Close() {
	if ms.server != nil {
		ms.server.Close()
	}
}

func TestSuite(t *testing.T) {
	testDir := "testdata"
	fileInfoList, err := ioutil.ReadDir(testDir)
	assert.NoError(t, err)

	// read all manifests
	manifestMap := make(map[string]map[string]interface{})
	for _, fileInfo := range fileInfoList {
		if strings.HasSuffix(fileInfo.Name(), "-manifest.jsonld") {
			inputBytes, err := ioutil.ReadFile(filepath.Join(testDir, fileInfo.Name()))
			assert.NoError(t, err)

			var manifest map[string]interface{}
			err = json.Unmarshal(inputBytes, &manifest)
			assert.NoError(t, err)

			manifestMap[fileInfo.Name()] = manifest
		}
	}

	dl := NewDefaultDocumentLoader(nil)
	proc := NewJsonLdProcessor()
	earlReport := NewEarlReport()

	for manifestName, manifest := range manifestMap {
		manifestURI := manifest["baseIri"].(string) + manifestName

		// start a mock HTTP server
		mockServer := NewMockServer(manifest["baseIri"].(string), testDir)
		defer mockServer.Close()

	SequenceLoop:
		for _, testData := range manifest["sequence"].([]interface{}) {
			testMap := testData.(map[string]interface{})
			testName := manifestURI + testMap["@id"].(string)

			inputURL := manifest["baseIri"].(string) + testMap["input"].(string)

			// read 'option' section and initialise JsonLdOptions and expected HTTP server responses

			options := NewJsonLdOptions("")

			var returnContentType string
			var returnHttpStatus int
			var returnRedirectTo string
			var returnHttpLink []string

			if optionVal, optionsPresent := testMap["option"]; optionsPresent {
				testOpts := optionVal.(map[string]interface{})

				if value, hasValue := testOpts["base"]; hasValue {
					options.Base = value.(string)
				}
				if value, hasValue := testOpts["expandContext"]; hasValue {
					contextDoc, err := dl.LoadDocument(filepath.Join(testDir, value.(string)))
					assert.NoError(t, err)
					options.ExpandContext = contextDoc.Document
				}
				if value, hasValue := testOpts["compactArrays"]; hasValue {
					options.CompactArrays = value.(bool)
				}
				if value, hasValue := testOpts["useNativeTypes"]; hasValue {
					options.UseNativeTypes = value.(bool)
				}
				if value, hasValue := testOpts["useRdfType"]; hasValue {
					options.UseRdfType = value.(bool)
				}
				if value, hasValue := testOpts["produceGeneralizedRdf"]; hasValue {
					options.ProduceGeneralizedRdf = value.(bool)
				}

				if value, hasValue := testOpts["contentType"]; hasValue {
					returnContentType = value.(string)
				}
				if value, hasValue := testOpts["httpStatus"]; hasValue {
					returnHttpStatus = int(value.(float64))
				}
				if value, hasValue := testOpts["redirectTo"]; hasValue {
					returnRedirectTo = value.(string)
				}
				if value, hasValue := testOpts["httpLink"]; hasValue {
					returnHttpLink = make([]string, 0)
					if valueList, isList := value.([]interface{}); isList {
						for _, link := range valueList {
							returnHttpLink = append(returnHttpLink, link.(string))
						}
					} else {
						returnHttpLink = append(returnHttpLink, value.(string))
					}
				}
			}

			mockServer.SetExpectedBehaviour(returnContentType, returnHttpLink, returnHttpStatus, returnRedirectTo)

			options.DocumentLoader = mockServer.DocumentLoader

			var result interface{}
			var opError error

			testType := testMap["@type"].([]interface{})

			switch {
			case testType[1] == "jld:ExpandTest":
				log.Println("Running Expand test", testMap["@id"], ":", testMap["name"])
				result, opError = proc.Expand(inputURL, options)
			case testType[1] == "jld:CompactTest":
				log.Println("Running Compact test", testMap["@id"], ":", testMap["name"])

				contextFilename := testMap["context"].(string)
				contextDoc, err := dl.LoadDocument(filepath.Join(testDir, contextFilename))
				assert.NoError(t, err)

				result, opError = proc.Compact(inputURL, contextDoc.Document, options)
			case testType[1] == "jld:FlattenTest":
				log.Println("Running Flatten test", testMap["@id"], ":", testMap["name"])

				var ctxDoc interface{}
				if ctxVal, hasContext := testMap["context"]; hasContext {
					contextFilename := ctxVal.(string)
					contextDoc, err := dl.LoadDocument(filepath.Join(testDir, contextFilename))
					assert.NoError(t, err)
					ctxDoc = contextDoc.Document
				}

				result, opError = proc.Flatten(inputURL, ctxDoc, options)
			case testType[1] == "jld:FrameTest":
				log.Println("Running Frame test", testMap["@id"], ":", testMap["name"])

				frameFilename := testMap["frame"].(string)
				frameDoc, err := dl.LoadDocument(filepath.Join(testDir, frameFilename))
				assert.NoError(t, err)

				result, opError = proc.Frame(inputURL, frameDoc.Document, options)
			case testType[1] == "jld:FromRDFTest":
				log.Println("Running FromRDF test", testMap["@id"], ":", testMap["name"])

				inputFilename := filepath.Join(testDir, testMap["input"].(string))
				inputBytes, err := ioutil.ReadFile(inputFilename)
				assert.NoError(t, err)
				input := string(inputBytes)

				result, err = proc.FromRDF(input, options)
			case testType[1] == "jld:ToRDFTest":
				log.Println("Running ToRDF test", testMap["@id"], ":", testMap["name"])

				options.Format = "application/nquads"
				result, opError = proc.ToRDF(inputURL, options)
			case testType[1] == "jld:NormalizeTest":
				log.Println("Running Normalize test", testMap["@id"], ":", testMap["name"])

				inputFilename := filepath.Join(testDir, testMap["input"].(string))
				rdIn, err := dl.LoadDocument(inputFilename)
				assert.NoError(t, err)
				input := rdIn.Document

				options.Format = "application/nquads"
				result, opError = proc.Normalize(input, options)
			default:
				break SequenceLoop
			}

			var expected interface{}
			var expectedType string
			if testType[0] == "jld:PositiveEvaluationTest" {
				// we don't expect any errors here
				if !assert.NoError(t, opError) {
					earlReport.addAssertion(testName, false)
				}

				// load expected document
				expectedFilename := filepath.Join(testDir, testMap["expect"].(string))
				expectedType = filepath.Ext(expectedFilename)
				if expectedType == ".jsonld" || expectedType == ".json" {
					// load as JSON-LD/JSON
					rdOut, err := dl.LoadDocument(filepath.Join(testDir, testMap["expect"].(string)))
					assert.NoError(t, err)
					expected = rdOut.Document
				} else if expectedType == ".nq" {
					// load as N-Quads
					expectedBytes, err := ioutil.ReadFile(expectedFilename)
					assert.NoError(t, err)
					expected = string(expectedBytes)
				}

				// marshal/unmarshal the result to avoid any differences due to formatting & key sequences
				resultBytes, _ := json.MarshalIndent(result, "", "  ")
				err = json.Unmarshal(resultBytes, &result)
			} else if testType[0] == "jld:NegativeEvaluationTest" {
				expected = testMap["expect"].(string)

				if opError != nil {
					result = string(opError.(*JsonLdError).Code)
				} else {
					result = ""
				}
			}

			if !assert.True(t, DeepCompare(expected, result, true)) {
				// print out expected vs. actual results in a human readable form
				if expectedType == ".jsonld" || expectedType == ".json" {
					log.Println("==== ACTUAL ====")
					b, _ := json.MarshalIndent(result, "", "  ")
					os.Stdout.Write(b)
					os.Stdout.WriteString("\n")
					log.Println("==== EXPECTED ====")
					b, _ = json.MarshalIndent(expected, "", "  ")
					os.Stdout.Write(b)

				} else if expectedType == ".nq" {
					log.Println("==== ACTUAL ====")
					os.Stdout.WriteString(result.(string))
					log.Println("==== EXPECTED ====")
					os.Stdout.WriteString(expected.(string))
				} else {
					log.Println("==== ACTUAL ====")
					os.Stdout.WriteString(result.(string))
					os.Stdout.WriteString("\n")
					log.Println("==== EXPECTED ====")
					os.Stdout.WriteString(expected.(string))
					os.Stdout.WriteString("\n")
				}
				log.Println("Error when running", testMap["@id"], "for", testType[1])
				earlReport.addAssertion(testName, false)
				return
			} else {
				earlReport.addAssertion(testName, true)
			}
		}
	}
	earlReport.write("earl.jsonld")
}

const (
	assertor     = "https://github.com/kazarena"
	assertorName = "Stan Nazarenko"
)

// EarlReport generates an EARL report.
type EarlReport struct {
	report map[string]interface{}
}

func NewEarlReport() *EarlReport {
	rval := &EarlReport{
		report: map[string]interface{}{
			"@context": map[string]interface{}{
				"doap":            "http://usefulinc.com/ns/doap#",
				"foaf":            "http://xmlns.com/foaf/0.1/",
				"dc":              "http://purl.org/dc/terms/",
				"earl":            "http://www.w3.org/ns/earl#",
				"xsd":             "http://www.w3.org/2001/XMLSchema#",
				"doap:homepage":   map[string]interface{}{"@type": "@id"},
				"doap:license":    map[string]interface{}{"@type": "@id"},
				"dc:creator":      map[string]interface{}{"@type": "@id"},
				"foaf:homepage":   map[string]interface{}{"@type": "@id"},
				"subjectOf":       map[string]interface{}{"@reverse": "earl:subject"},
				"earl:assertedBy": map[string]interface{}{"@type": "@id"},
				"earl:mode":       map[string]interface{}{"@type": "@id"},
				"earl:test":       map[string]interface{}{"@type": "@id"},
				"earl:outcome":    map[string]interface{}{"@type": "@id"},
				"dc:date":         map[string]interface{}{"@type": "xsd:date"},
			},
			"@id": "https://github.com/kazarena/json-gold",
			"@type": []interface{}{
				"doap:Project",
				"earl:TestSubject",
				"earl:Software",
			},
			"doap:name":                 "JSON-goLD",
			"dc:title":                  "JSON-goLD",
			"doap:homepage":             "https://github.com/kazarena/json-gold",
			"doap:license":              "https://github.com/kazarena/json-gold/blob/master/LICENSE",
			"doap:description":          "A JSON-LD processor for Go",
			"doap:programming-language": "Go",
			"dc:creator":                assertor,
			"doap:developer": map[string]interface{}{
				"@id": assertor,
				"@type": []interface{}{
					"foaf:Person",
					"earl:Assertor",
				},
				"foaf:name":     assertorName,
				"foaf:homepage": assertor,
			},
			"dc:date": map[string]interface{}{
				"@value": time.Now().Format("2006-01-02"),
				"@type":  "xsd:date",
			},
			"subjectOf": make([]interface{}, 0),
		},
	}

	return rval
}

func (er *EarlReport) addAssertion(testName string, success bool) {
	var outcome string
	if success {
		outcome = "earl:passed"
	} else {
		outcome = "earl:failed"
	}
	er.report["subjectOf"] = append(
		er.report["subjectOf"].([]interface{}),
		map[string]interface{}{
			"@type":           "earl:Assertion",
			"earl:assertedBy": assertor,
			"earl:mode":       "earl:automatic",
			"earl:test":       testName,
			"earl:result": map[string]interface{}{
				"@type":        "earl:TestResult",
				"dc:date":      time.Now().Format("2006-01-02T15:04:05.999999"),
				"earl:outcome": outcome,
			},
		},
	)
}

func (er *EarlReport) write(filename string) {
	b, _ := json.MarshalIndent(er.report, "", "  ")

	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(b)
	f.WriteString("\n")
}
