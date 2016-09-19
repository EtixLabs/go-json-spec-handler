package jsh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// JSONAPIVersion is version of JSON API Spec that is currently compatible:
// http://jsonapi.org/format/1.1/
const JSONAPIVersion = "1.1"

// Sendable implements functions that allows different response types
// to produce a sendable JSON Response format
type Sendable interface {
	Validate(r *http.Request, response bool) *Error
}

// Send will respond with the given JSON payload to the client. If the payload response validation
// fails, it will respond with the validation error and will return it.
// Send is designed to always send a response, but will also return the last
// error it encountered to help with debugging in the event of an Internal Server
// Error.
func Send(w http.ResponseWriter, r *http.Request, payload Sendable) *Error {
	// Validate payload
	var doc *Document
	validationErr := payload.Validate(r, true)
	if validationErr == nil {
		// Build and validate document
		doc = Build(payload)
		validationErr = doc.Validate(r, true)
	}
	if validationErr != nil {
		// Make the validation error the new response
		doc = Build(validationErr)
		if err := doc.Validate(r, true); err != nil {
			// If we ever hit this, something seriously wrong has happened
			http.Error(w, DefaultErrorTitle, http.StatusInternalServerError)
			return err
		}
	}

	err := sendDocument(w, doc)
	if err != nil {
		return err
	}
	return validationErr
}

// Ok makes it simple to return a 200 OK response via jsh:
//
//	jsh.Send(w, r, jsh.Ok())
func Ok() *Document {
	doc := New()
	doc.Status = http.StatusOK
	doc.empty = true

	return doc
}

// sendDocument marshals the document, sets the header and writes the result to the given writer.
func sendDocument(w http.ResponseWriter, document *Document) *Error {
	content, err := json.MarshalIndent(document, "", " ")
	if err != nil {
		http.Error(w, DefaultErrorTitle, http.StatusInternalServerError)
		return ISE(fmt.Sprintf("Unable to marshal JSON payload: %v", err))
	}

	w.Header().Add("Content-Type", ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(document.Status)
	w.Write(content)
	return nil
}
