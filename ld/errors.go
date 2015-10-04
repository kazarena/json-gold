package ld

import (
	"fmt"
)

// ErrorCode is a JSON-LD error code as per spec.
type ErrorCode string

// JsonLdError is a JSON-LD error as defined in the spec.
// See the allowed values and error messages below.
type JsonLdError struct {
	Code    ErrorCode
	Details interface{}
}

const (
	LoadingDocumentFailed       ErrorCode = "loading document failed"
	ListOfLists                 ErrorCode = "list of lists"
	InvalidIndexValue           ErrorCode = "invalid @index value"
	ConflictingIndexes          ErrorCode = "conflicting indexes"
	InvalidIDValue              ErrorCode = "invalid @id value"
	InvalidLocalContext         ErrorCode = "invalid local context"
	MultipleContextLinkHeaders  ErrorCode = "multiple context link headers"
	LoadingRemoteContextFailed  ErrorCode = "loading remote context failed"
	InvalidRemoteContext        ErrorCode = "invalid remote context"
	RecursiveContextInclusion   ErrorCode = "recursive context inclusion"
	InvalidBaseIRI              ErrorCode = "invalid base IRI"
	InvalidVocabMapping         ErrorCode = "invalid vocab mapping"
	InvalidDefaultLanguage      ErrorCode = "invalid default language"
	KeywordRedefinition         ErrorCode = "keyword redefinition"
	InvalidTermDefinition       ErrorCode = "invalid term definition"
	InvalidReverseProperty      ErrorCode = "invalid reverse property"
	InvalidIRIMapping           ErrorCode = "invalid IRI mapping"
	CyclicIRIMapping            ErrorCode = "cyclic IRI mapping"
	InvalidKeywordAlias         ErrorCode = "invalid keyword alias"
	InvalidTypeMapping          ErrorCode = "invalid type mapping"
	InvalidLanguageMapping      ErrorCode = "invalid language mapping"
	CollidingKeywords           ErrorCode = "colliding keywords"
	InvalidContainerMapping     ErrorCode = "invalid container mapping"
	InvalidTypeValue            ErrorCode = "invalid type value"
	InvalidValueObject          ErrorCode = "invalid value object"
	InvalidValueObjectValue     ErrorCode = "invalid value object value"
	InvalidLanguageTaggedString ErrorCode = "invalid language-tagged string"
	InvalidLanguageTaggedValue  ErrorCode = "invalid language-tagged value"
	InvalidTypedValue           ErrorCode = "invalid typed value"
	InvalidSetOrListObject      ErrorCode = "invalid set or list object"
	InvalidLanguageMapValue     ErrorCode = "invalid language map value"
	CompactionToListOfLists     ErrorCode = "compaction to list of lists"
	InvalidReversePropertyMap   ErrorCode = "invalid reverse property map"
	InvalidReverseValue         ErrorCode = "invalid @reverse value"
	InvalidReversePropertyValue ErrorCode = "invalid reverse property value"

	// non spec related errors
	SyntaxError    ErrorCode = "syntax error"
	NotImplemented ErrorCode = "not implemnted"
	UnknownFormat  ErrorCode = "unknown format"
	InvalidInput   ErrorCode = "invalid input"
	ParseError     ErrorCode = "parse error"
	IOError        ErrorCode = "io error"
	UnknownError   ErrorCode = "unknown error"
)

func (e JsonLdError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%v: %v", e.Code, e.Details)
	}
	return fmt.Sprintf("%v", e.Code)
}

// NewJsonLdError creates a new instance of JsonLdError.
func NewJsonLdError(code ErrorCode, details interface{}) *JsonLdError {
	return &JsonLdError{Code: code, Details: details}
}
