package jsh

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

/*
DefaultError can be customized in order to provide a more customized error
Detail message when an Internal Server Error occurs. Optionally, you can modify
a returned jsh.Error before sending it as a response as well.
*/
var DefaultErrorDetail = "Request failed, something went wrong"

// DefaultTitle can be customized to provide a more customized ISE Title
var DefaultErrorTitle = "Internal Server Error"

/*
ErrorType represents the common interface requirements that libraries may
specify if they would like to accept either a single error or a list.
*/
type ErrorType interface {
	// Error returns a formatted error and allows it to conform to the stdErr
	// interface.
	Error() string
	// Validate checks that the error is valid in the context of JSONAPI
	Validate(r *http.Request, response bool) *Error
	// StatusCode returns the first encountered HTTP Status Code for the error type.
	// Returns 0 if none is set.
	StatusCode() int
}

// ErrorList is wraps an Error Array so that it can implement Sendable
type ErrorList []*Error

// Validate checks all errors within the list to ensure that they are valid
func (e ErrorList) Validate(r *http.Request, response bool) *Error {
	for _, err := range e {
		validationErr := err.Validate(r, response)
		if validationErr != nil {
			return validationErr
		}
	}

	return nil
}

// Fulfills the default error interface
func (e ErrorList) Error() string {
	var msg string

	for _, err := range e {
		msg += fmt.Sprintf("%s\n", err.Error())
	}

	return msg
}

/*
StatusCode (HTTP) of the first error in the list. Defaults to 0 if the list is
empty or one has not yet been set for the first error.
*/
func (e ErrorList) StatusCode() int {
	if len(e) == 0 {
		return 0
	}

	return e[0].Status
}

// ErrorSource represents the source of a JSONAPI error, either by a pointer or a query parameter name.
type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

/*
Error consists of a number of contextual attributes to make conveying
certain error type simpler as per the JSON API specification:
http://jsonapi.org/format/#error-objects

	error := &jsh.Error{
		Title: "Authentication Failure",
		Detail: "Category 4 Username Failure",
		Status: 401
	}

	jsh.Send(w, r, error)
*/
type Error struct {
	Status int          `json:"status,string"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
	ISE    string       `json:"-"`
}

/*
Error will print an internal server error if set, or default back to the SafeError()
format if not. As usual, err.Error() should not be considered safe for presentation
to the end user, use err.SafeError() instead.
*/
func (e *Error) Error() string {
	msg := fmt.Sprintf("%d: %s - %s", e.Status, e.Title, e.Detail)
	if e.Source != nil && e.Source.Pointer != "" {
		msg += fmt.Sprintf(" (Source.Pointer: %s)", e.Source.Pointer)
	}

	if e.ISE != "" {
		msg += fmt.Sprintf(": %s", e.ISE)
	}

	return msg
}

/*
Validate ensures that the error meets all JSON API criteria.
*/
func (e *Error) Validate(r *http.Request, response bool) *Error {

	switch {
	case e.Status == 0:
		return ISE(fmt.Sprintf("No HTTP Status set for error %+v\n", e))
	case e.Status < 400 || e.Status > 600:
		return ISE(fmt.Sprintf("HTTP Status out of valid range for error %+v\n", e))
	case e.Status == 422 && (e.Source == nil || e.Source.Pointer == ""):
		return ISE(fmt.Sprintf("Source Pointer must be set for 422 Status error"))
	}

	return nil
}

/*
StatusCode (HTTP) for the error. Defaults to 0.
*/
func (e *Error) StatusCode() int {
	return e.Status
}

// BadRequestError is a convenience function to return a 400 Bad Request response.
func BadRequestError(msg string, detail string) *Error {
	return &Error{
		Title:  msg,
		Detail: detail,
		Status: http.StatusBadRequest,
	}
}

/*
ParameterError creates a properly formatted HTTP Status 400 error with an appropriate
user safe message. The err.Source.Parameter field will be set to the parameter "param".
*/
func ParameterError(msg string, param string) *Error {
	return &Error{
		Title:  "Invalid Query Parameter",
		Detail: msg,
		Status: http.StatusBadRequest,
		Source: &ErrorSource{
			Parameter: strings.ToLower(param),
		},
	}
}

// ForbiddenError is used whenever an attempt to do a forbidden operation is made.
func ForbiddenError(msg string) *Error {
	return &Error{
		Title:  msg,
		Status: http.StatusForbidden,
	}
}

// NotFound returns a 404 formatted error.
func NotFound(resourceType string, id string) *Error {
	return &Error{
		Title:  "Not Found",
		Detail: fmt.Sprintf("No resource of type '%s' exists for ID: %s", resourceType, id),
		Status: http.StatusNotFound,
	}
}

// SpecificationError returnss a 406 Not Acceptable.
// It is used whenever the Client violates the JSON API Spec.
func SpecificationError(detail string) *Error {
	return &Error{
		Title:  "JSON API Specification Error",
		Detail: detail,
		Status: http.StatusNotAcceptable,
	}
}

// ConflictError returns a 409 Conflict error.
func ConflictError(resourceType string, id string) *Error {
	var detail string
	if id == "" {
		detail = fmt.Sprintf("Resource type '%s' does not match URL's", resourceType)
	} else {
		detail = fmt.Sprintf("ID '%s' does not match URL's", id)
	}
	return &Error{
		Title:  "Resource conflict",
		Detail: detail,
		Status: http.StatusConflict,
	}
}

// TopLevelError is used whenever the client sends a JSON payload with a missing top-level field.
func TopLevelError(field string) *Error {
	// NOTE: Here we should point to the top-level of the document (""),
	// but as it is also the empty string value it would be ignored by marshalling.
	// Instead we point to “/” even if it is an appropriate reference to
	// the string `"some value"` in the request document `{"": "some value"}`.
	// The detail message however eliminates the misunderstanding by specifying
	// the name of the missing field.
	err := &Error{
		Detail: fmt.Sprintf("Missing `%s` at document's top level", strings.ToLower(field)),
		Status: 422,
		Source: &ErrorSource{Pointer: "/"},
	}
	return err
}

/*
InputError creates a properly formatted HTTP Status 422 error with an appropriate
user safe message. The parameter "attribute" will format err.Source.Pointer to be
"/data/attributes/<attribute>".
*/
func InputError(msg string, attribute string) *Error {
	return &Error{
		Title:  "Invalid Attribute",
		Detail: msg,
		Status: 422,
		Source: &ErrorSource{
			Pointer: AttributePointer(attribute),
		},
	}
}

/*
RelationshipError creates a properly formatted HTTP Status 422 error with an appropriate
user safe message. The parameter "relationship" will format err.Source.Pointer to be
"/data/relationship/<attribute>".
*/
func RelationshipError(msg string, relationship string) *Error {
	return &Error{
		Title:  "Invalid Relationship",
		Detail: msg,
		Status: 422,
		Source: &ErrorSource{
			Pointer: RelationshipPointer(relationship),
		},
	}
}

/*
ISE is a convenience function for creating a ready-to-go Internal Service Error
response. The message you pass in is set to the ErrorObject.ISE attribute so you
can gracefully log ISE's internally before sending them.
*/
func ISE(internalMessage string) *Error {
	return &Error{
		Title:  DefaultErrorTitle,
		Detail: DefaultErrorDetail,
		Status: http.StatusInternalServerError,
		ISE:    internalMessage,
	}
}

// NotImplemented is a convenience function similar to ISE except if generates a 501 response.
func NotImplemented(internalMessage string) *Error {
	return &Error{
		Title:  "Not implemented",
		Status: http.StatusNotImplemented,
		ISE:    internalMessage,
	}
}

// AttributePointer returns a JSON pointer to the given attribute in a JSON API document.
func AttributePointer(attribute string) string {
	return fmt.Sprintf("/data/attributes/%s", attribute)
}

// RelationshipPointer returns a JSON pointer to the given primary resource relationship in a JSON API document.
func RelationshipPointer(relationship string) string {
	return fmt.Sprintf("/data/relationships/%s", relationship)
}

// toLowerFirstRune changes the first rune of the given string to lower case.
func toLowerFirstRune(s string) string {
	if len(s) == 0 {
		return s
	}
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}
