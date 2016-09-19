package jsc

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/EtixLabs/go-json-spec-handler"
)

// TopLevelAction performs an outbound POST /action request
func TopLevelAction(baseURL, action string, payload jsh.Sendable) (*jsh.Document, *http.Response, error) {
	request, err := TopLevelActionRequest(baseURL, action, payload)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

/*
TopLevelActionRequest returns a fully formatted JSONAPI Action (POST /action) request.
Useful if you need to set custom headers before proceeding. Otherwise just use "jsh.TopLevelAction".
*/
func TopLevelActionRequest(baseURL, action string, payload jsh.Sendable) (*http.Request, error) {
	if action == "" {
		return nil, errors.New("Action specifier cannot be empty for an Action request type")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %v", err)
	}
	setPath(u, action)
	return postRequest(u, payload)
}

// Action performs an outbound POST /resource/:id/action request
func Action(baseURL, resourceType, id, action string, payload jsh.Sendable) (*jsh.Document, *http.Response, error) {
	request, err := ActionRequest(baseURL, resourceType, id, action, payload)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

/*
ActionRequest returns a fully formatted JSONAPI Action (POST /resource/:id/action) request.
Useful if you need to set custom headers before proceeding. Otherwise just use "jsh.Action".
*/
func ActionRequest(baseURL, resourceType, id, action string, payload jsh.Sendable) (*http.Request, error) {
	if action == "" {
		return nil, errors.New("Action specifier cannot be empty for an Action request type")
	}
	u, err := fetchURL(baseURL, resourceType, id, action)
	if err != nil {
		return nil, err
	}
	return postRequest(u, payload)
}
