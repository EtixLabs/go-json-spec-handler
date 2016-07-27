package jsc

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/EtixLabs/go-json-spec-handler"
)

/*
Delete allows a user to make an outbound "DELETE /resource/:id" request.

	resp, err := jsh.Delete("http://apiserver", "user", "2")
*/
func Delete(baseURL, resourceType, id string) (*http.Response, error) {
	request, err := DeleteRequest(baseURL, resourceType, id)
	if err != nil {
		return nil, err
	}
	_, response, err := Do(request, jsh.ObjectMode)
	if err != nil {
		return nil, err
	}
	return response, nil
}

/*
DeleteRequest returns a fully formatted request for performing a JSON API DELETE.
This is useful for if you need to set custom headers on the request. Otherwise
just use "jsc.Delete".
*/
func DeleteRequest(baseURL, resourceType, id string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %v", err)
	}
	setIDPath(u, resourceType, id)
	return deleteRequest(u)
}

// DeleteMany allows a consumer to perform a DELETE /resources/:id/relationships/relationship request.
func DeleteMany(baseURL, resourceType, id, relationship string, list jsh.IDList) (*jsh.Document, *http.Response, error) {
	request, err := DeleteManyRequest(baseURL, resourceType, id, relationship, list)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// DeleteManyRequest returns a fully formatted request with JSON body for performing
// a JSONAPI DELETE on a relationship. This is useful for if you need to set custom headers on the
// request. Otherwise just use "jsc.DeleteMany".
func DeleteManyRequest(baseURL, resourceType, id, relationship string, list jsh.IDList) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, "relationships/"+relationship)
	if err != nil {
		return nil, err
	}
	request, err := deleteRequest(u)
	if err != nil {
		return nil, err
	}
	err = prepareBody(request, list)
	if err != nil {
		return nil, err
	}
	return request, nil
}

// deleteRequest creates a fully formatted DELETE request with JSON body from the given URL and object.
func deleteRequest(u *url.URL) (*http.Request, error) {
	request, err := NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating DELETE request: %v", err)
	}
	return request, nil
}
