package jsc

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/EtixLabs/go-json-spec-handler"
)

// Fetch performs an outbound GET /resources/:id request
func Fetch(baseURL string, resourceType string, id string) (*jsh.Document, *http.Response, error) {
	request, err := FetchRequest(baseURL, resourceType, id)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

/*
FetchRequest returns a fully formatted JSONAPI Fetch request. Useful if you need to
set custom headers before proceeding. Otherwise just use "jsh.Fetch".
*/
func FetchRequest(baseURL, resourceType, id string) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, "")
	if err != nil {
		return nil, err
	}
	return NewRequest("GET", u.String(), nil)
}

// FetchRelated performs an outbound GET /resources/:id/relationship request
func FetchRelated(baseURL, resourceType, id, relationship string) (*jsh.Document, *http.Response, error) {
	request, err := FetchRelatedRequest(baseURL, resourceType, id, relationship)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// ListRelated performs an outbound GET /resources/:id/relationship request
func ListRelated(baseURL, resourceType, id, relationship string) (*jsh.Document, *http.Response, error) {
	request, err := FetchRelatedRequest(baseURL, resourceType, id, relationship)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ListMode)
}

/*
FetchRelatedRequest returns a fully formatted JSONAPI Fetch request for a to-one relationship resource.
Useful if you need to set custom headers before proceeding. Otherwise just use "jsh.FetchRelated".
*/
func FetchRelatedRequest(baseURL, resourceType, id, relationship string) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, relationship)
	if err != nil {
		return nil, err
	}
	return NewRequest("GET", u.String(), nil)
}

// FetchRelationship performs an outbound GET /resources/:id/relationships/relationship request
func FetchRelationship(baseURL, resourceType, id, relationship string) (*jsh.Document, *http.Response, error) {
	request, err := FetchRelationshipRequest(baseURL, resourceType, id, relationship)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ObjectMode)
}

// ListRelationship performs an outbound GET /resources/:id/relationships/relationship request
func ListRelationship(baseURL, resourceType, id, relationship string) (*jsh.Document, *http.Response, error) {
	request, err := FetchRelationshipRequest(baseURL, resourceType, id, relationship)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ListMode)
}

/*
FetchRelationshipRequest returns a fully formatted JSONAPI Fetch request for a to-one relationship.
Useful if you need to set custom headers before proceeding. Otherwise just use "jsh.FetchRelationship".
*/
func FetchRelationshipRequest(baseURL, resourceType, id, relationship string) (*http.Request, error) {
	u, err := fetchURL(baseURL, resourceType, id, "relationships/"+relationship)
	if err != nil {
		return nil, err
	}
	return NewRequest("GET", u.String(), nil)
}

// List performs an outbound GET /resourceTypes request
func List(baseURL, resourceType string) (*jsh.Document, *http.Response, error) {
	request, err := ListRequest(baseURL, resourceType)
	if err != nil {
		return nil, nil, err
	}
	return Do(request, jsh.ListMode)
}

/*
ListRequest returns a fully formatted JSONAPI List request. Useful if you need to
set custom headers before proceeding. Otherwise just use "jsh.List".
*/
func ListRequest(baseURL, resourceType string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %v", err)
	}
	setPath(u, resourceType)

	return NewRequest("GET", u.String(), nil)
}

// fetchURL creates a fully formatted URL representing a resource or relationship from the given parts.
func fetchURL(baseURL, resourceType, id, relationship string) (*url.URL, error) {
	if id == "" {
		return nil, errors.New("ID cannot be empty for an Object request type")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %v", err)
	}

	setIDPath(u, resourceType, id)
	if relationship != "" {
		// concat the relationship the end of url.Path, ensure no "/" prefix
		if strings.HasPrefix(relationship, "/") {
			relationship = relationship[1:]
		}
		u.Path = strings.Join([]string{u.Path, relationship}, "/")
	}

	return u, nil
}
