# JSON-goLD

This library is an implementation of the [JSON-LD](http://json-ld.org/) specification in Go.

### Testing & Compliance ###

As of April 1, 2015 all tests from the [official JSON-LD test suite](https://github.com/json-ld/json-ld.org/tree/master/test-suite) pass. Thanks to httptest Go package, it takes around 300ms to run the whole suite while making live connections to a mock HTTP server.

### Inspiration ###

This implementation was heavily influenced by [JSONLD-Java](https://github.com/jsonld-java/jsonld-java) with some techniques borrowed from [PyLD](https://github.com/digitalbazaar/pyld) and [gojsonld](https://github.com/linkeddata/gojsonld). Big thank you to the contributors of the forementioned libraries for figuring out implementation details of the core algorithms.
