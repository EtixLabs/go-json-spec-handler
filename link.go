package jsh

import (
	"encoding/json"
	"fmt"
)

// Links is a top-level document field
type Links struct {
	Self    *Link `json:"self,omitempty"`
	Related *Link `json:"related,omitempty"`
}

// NewRelationshipLinks creates a new pair of relationship links encoded as a string.
func NewRelationshipLinks(id interface{}, resource, name string) *Links {
	return &Links{
		Self:    NewRelationshipLink(id, resource, name, true),
		Related: NewRelationshipLink(id, resource, name, false),
	}
}

// Link is a resource link that can encode as a string or as an object
// as per the JSON API specification.
type Link struct {
	HREF string                 `json:"href,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// NewLink creates a new link encoded as a string.
func NewLink(href string) *Link {
	return &Link{
		HREF: href,
	}
}

// NewSelfLink creates a new self link encoded as a string.
func NewSelfLink(id interface{}, resource string) *Link {
	return NewLink(fmt.Sprintf("/%s/%v", resource, id))
}

// NewRelationshipLink creates a new relationship link encoded as a string.
func NewRelationshipLink(id interface{}, resource, name string, relationship bool) *Link {
	if relationship {
		return NewLink(fmt.Sprintf("/%s/%v/relationships/%s", resource, id, name))
	} else {
		return NewLink(fmt.Sprintf("/%s/%v/%s", resource, id, name))
	}
}

// NewMetaLink creates a new link with metadata encoded as an object.
func NewMetaLink(href string, meta map[string]interface{}) *Link {
	return &Link{
		HREF: href,
		Meta: meta,
	}
}

// MarshalJSON implements the Marshaler interface for Link.
func (l *Link) MarshalJSON() ([]byte, error) {
	if l.Meta == nil {
		return json.Marshal(l.HREF)
	}
	// Create a sub-type here so when we call Marshal below, we don't recursively
	// call this function over and over
	type MarshalLink Link
	return json.Marshal(MarshalLink(*l))
}

// UnmarshalJSON implements the Unmarshaler interface for Link.
func (l *Link) UnmarshalJSON(data []byte) error {
	var href string
	err := json.Unmarshal(data, &href)
	if err == nil {
		l.HREF = href
		return nil
	}
	// Create a sub-type here so when we call Unmarshal below, we don't recursively
	// call this function over and over
	type UnmarshalLink Link
	link := UnmarshalLink{}

	err = json.Unmarshal(data, &link)
	if err != nil {
		return err
	}
	*l = Link(link)
	return nil
}
