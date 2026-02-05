// This package validates JSON keys before delegating actual parsing to
// encoding/json, preserving full compatibility with custom UnmarshalJSON
// implementations
package strictjson

import "fmt"

const (
	errPrefixNonPointer = "strictjson: Unmarshal(non-pointer)"
)

type UnmarshalError struct {
	message string
}

func (e *UnmarshalError) Error() string {
	return e.message
}

func newNonPointerError() error {
	return &UnmarshalError{message: errPrefixNonPointer}
}

type unknownFieldError struct {
	fieldName  string
	suggestion string
}

func (e *unknownFieldError) Error() string {
	if e.suggestion != "" {
		return fmt.Sprintf(`strictjson: unknown field "%s" (did you mean "%s"?)`, e.fieldName, e.suggestion)
	}
	return fmt.Sprintf(`strictjson: unknown or mis-cased field "%s"`, e.fieldName)
}

func newUnknownFieldError(fieldName, suggestion string) error {
	return &unknownFieldError{
		fieldName:  fieldName,
		suggestion: suggestion,
	}
}

type fieldConflictError struct {
	fieldName string
}

func (e *fieldConflictError) Error() string {
	return fmt.Sprintf(`strictjson: field conflict: "%s" defined in multiple embedded structs`, e.fieldName)
}

func newFieldConflictError(fieldName string) error {
	return &fieldConflictError{fieldName: fieldName}
}
