package jsc

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/EtixLabs/go-json-spec-handler"
)

// Post allows a user to make an outbound POST /resources request:
//
//	obj, _ := jsh.NewObject("123", "user", payload)
//	// does POST http://apiserver/user/123
//	json, resp, err := jsh.Post("http://apiserver", obj)
func Post(baseURL string, object *jsh.Object) (*jsh.Document, *http.Response, error) {
	request, err := PostRequest(baseURL, object)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// PostRequest returns a fully formatted request with JSON body for performing
// a JSONAPI POST. This is useful for if you need to set custom headers on the
// request. Otherwise just use "jsc.Post".
func PostRequest(baseURL string, object *jsh.Object) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %v", err)
	}
	setPath(u, object.Type)
	return postRequest(u, object)
}

// PostMany allows a consumer to perform a POST /resources/:id/relationships/relationship request.
func PostMany(baseURL, resourceType, id, relationship string, list jsh.IDList) (*jsh.Document, *http.Response, error) {
	request, err := PostManyRequest(baseURL, resourceType, id, relationship, list)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// PostManyRequest returns a fully formatted request with JSON body for performing
// a JSONAPI POST on a relationship. This is useful for if you need to set custom headers on the
// request. Otherwise just use "jsc.PostMany".
func PostManyRequest(baseURL, resourceType, id, relationship string, list jsh.IDList) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, "relationships/"+relationship)
	if err != nil {
		return nil, err
	}
	return postRequest(u, list)
}

// postRequest creates a fully formatted POST request with JSON body from the given URL and object.
func postRequest(u *url.URL, payload jsh.Sendable) (*http.Request, error) {
	request, err := NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error building POST request: %v", err)
	}
	if payload != nil {
		err = prepareBody(request, payload)
		if err != nil {
			return nil, err
		}
	}
	return request, nil
}
