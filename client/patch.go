package jsc

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/EtixLabs/go-json-spec-handler"
)

// Patch allows a consumer to perform a PATCH /resources/:id request
// Example:
//
//  obj, _ := jsh.NewObject("123", "user", payload)
//	// does PATCH /http://postap.com/api/user/123
//  json, resp, err := jsc.Patch("http://postap.com/api/", obj)
//	updatedObj := json.First()
//
func Patch(baseURL string, object *jsh.Object) (*jsh.Document, *http.Response, error) {
	request, err := PatchRequest(baseURL, object)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// PatchRequest returns a fully formatted request with JSON body for performing
// a JSONAPI PATCH. This is useful for if you need to set custom headers on the
// request. Otherwise just use "jsc.Patch".
func PatchRequest(baseURL string, object *jsh.Object) (*http.Request, error) {
	u, err := fetchURL(baseURL, object.Type, object.ID, "")
	if err != nil {
		return nil, err
	}
	return patchRequest(u, object)
}

// PatchOne allows a consumer to perform a PATCH /resources/:id/relationships/relationship to-one request.
func PatchOne(baseURL, resourceType, id, relationship string, object *jsh.IDObject) (*jsh.Document, *http.Response, error) {
	request, err := PatchOneRequest(baseURL, resourceType, id, relationship, object)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// PatchOneRequest returns a fully formatted request with JSON body for performing
// a JSONAPI PATCH on a to-one relationship. This is useful for if you need to set custom headers on the
// request. Otherwise just use "jsc.PatchOne".
func PatchOneRequest(baseURL, resourceType, id, relationship string, object *jsh.IDObject) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, "relationships/"+relationship)
	if err != nil {
		return nil, err
	}
	return patchRequest(u, object)
}

// PatchMany allows a consumer to perform a PATCH /resources/:id/relationships/relationship to-many request.
func PatchMany(baseURL, resourceType, id, relationship string, list jsh.IDList) (*jsh.Document, *http.Response, error) {
	request, err := PatchManyRequest(baseURL, resourceType, id, relationship, list)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ListMode)
}

// PatchManyRequest returns a fully formatted request with JSON body for performing
// a JSONAPI PATCH on a to-many relationship. This is useful for if you need to set custom headers on the
// request. Otherwise just use "jsc.PatchMany".
func PatchManyRequest(baseURL, resourceType, id, relationship string, list jsh.IDList) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, "relationships/"+relationship)
	if err != nil {
		return nil, err
	}
	return patchRequest(u, list)
}

// patchRequest creates a fully formatted PATCH request with JSON body from the given URL and object.
func patchRequest(u *url.URL, payload jsh.Sendable) (*http.Request, error) {
	request, err := NewRequest("PATCH", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating PATCH request: %v", err)
	}
	err = prepareBody(request, payload)
	if err != nil {
		return nil, err
	}
	return request, nil
}
