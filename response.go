package jsh

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// JSONAPIVersion is version of JSON API Spec that is currently compatible:
// http://jsonapi.org/format/1.1/
const JSONAPIVersion = "1.1"

// Sendable implements functions that allows different response types
// to produce a sendable JSON Response format
type Sendable interface {
	prepare(r *http.Request, response bool) (*Response, SendableError)
}

// Response represents the top level json format of incoming requests
// and outgoing responses
type Response struct {
	HTTPStatus int         `json:"-"`
	Validated  bool        `json:"-"`
	Data       interface{} `json:"data,omitempty"`
	Errors     interface{} `json:"errors,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
	Links      *Link       `json:"links,omitempty"`
	Included   *List       `json:"included,omitempty"`
	JSONAPI    struct {
		Version string `json:"version"`
	} `json:"jsonapi"`
}

// Validate checks JSON Spec for the top level JSON document
func (r *Response) Validate() SendableError {

	if r.Errors == nil && r.Data == nil {
		return ISE("Both `errors` and `data` cannot be blank for a JSON response")
	}
	if r.Errors != nil && r.Data != nil {
		return ISE("Both `errors` and `data` cannot be set for a JSON response")
	}
	if r.Data == nil && r.Included != nil {
		return ISE("'included' should only be set for a response if 'data' is as well")
	}
	if r.HTTPStatus < 100 || r.HTTPStatus > 600 {
		return ISE("Response HTTP Status must be of a valid range")
	}

	// probably not the best place for this, but...
	r.JSONAPI.Version = JSONAPIVersion

	return nil
}

// Send fires a JSON response if the payload is prepared successfully, otherwise it
// returns an Error which can also be sent.
func Send(w http.ResponseWriter, r *http.Request, payload Sendable) {
	response, err := payload.prepare(r, true)
	if err != nil {
		response, err = err.prepare(r, true)

		// If we ever hit this, something seriously wrong has happened
		if err != nil {
			log.Printf("Error preparing JSH error: %s", err.Error())
			http.Error(w, DefaultErrorTitle, http.StatusInternalServerError)
			return
		}
	}

	SendResponse(w, r, response)
}

// SendResponse handles sending a fully packaged JSON Response allows API consumers
// to more manually build their Responses in case they want to send Meta, Links, etc
func SendResponse(w http.ResponseWriter, r *http.Request, response *Response) {

	err := response.Validate()
	if err != nil {
		response, err = err.prepare(r, true)

		// If we ever hit this, something seriously wrong has happened
		if err != nil {
			log.Printf("Error preparing JSH error: %s", err.Error())
			http.Error(w, DefaultErrorTitle, http.StatusInternalServerError)
		}
	}

	content, jsonErr := json.MarshalIndent(response, "", "  ")
	if jsonErr != nil {
		log.Printf("Unable to prepare JSON content: %s", jsonErr)
		http.Error(w, DefaultErrorTitle, http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(response.HTTPStatus)
	w.Write(content)
}