package jsh

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"

	"encoding/json"
)

// Relationship represents a reference from the resource object in which it's
// defined to other resource objects.
type Relationship struct {
	Links *Links                 `json:"links,omitempty"`
	Data  IDList                 `json:"data,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}

// IDObject identifies an individual resource.
type IDObject struct {
	Type string `json:"type" valid:"required"`
	ID   string `json:"id" valid:"required"`
}

// NewIDObject creates a new resource identifier object instance.
func NewIDObject(tp, id string) *IDObject {
	return &IDObject{
		Type: tp,
		ID:   id,
	}
}

// ToObject returns a new object instance from the given resource object.
func (obj *IDObject) ToObject() *Object {
	// We can safely ignore the error when attributes are nil
	result, _ := NewObject(obj.ID, obj.Type, nil)
	return result
}

// Validate ensures that the relationship is JSON API compatible.
func (obj *IDObject) Validate(r *http.Request, response bool) *Error {
	adapter := func(err govalidator.Error) *Error {
		return SpecificationError(err.Err.Error())
	}
	errlist := validator(obj, adapter)
	if len(errlist) > 0 {
		return errlist[0]
	}
	return nil
}

// IDList is a wrapper around a resource identifier slice that implements Sendable and Unmarshaler.
// IDList also implements sort.Interface for []*IDObject based on the ID field.
type IDList []*IDObject

// ToList returns a list with a new object instance for each resource object.
func (list IDList) ToList() List {
	result := make(List, 0, len(list))
	for _, obj := range list {
		result = append(result, obj.ToObject())
	}
	return result
}

// Validate ensures that the relationship list is JSON API compatible.
func (list IDList) Validate(r *http.Request, response bool) *Error {
	for _, relationship := range list {
		if err := relationship.Validate(r, response); err != nil {
			return err
		}
	}
	return nil
}

/*
UnmarshalJSON allows us to manually decode a the resource linkage via the
json.Unmarshaler interface.
*/
func (l *IDList) UnmarshalJSON(data []byte) error {
	// Create a sub-type here so when we call Unmarshal below, we don't recursively
	// call this function over and over
	type UnmarshalLinkage IDList

	// if our "List" is a single object, modify the JSON to make it into a list
	// by wrapping with "[ ]"
	if data[0] == '{' {
		data = []byte(fmt.Sprintf("[%s]", data))
	}

	newLinkage := UnmarshalLinkage{}
	err := json.Unmarshal(data, &newLinkage)
	if err != nil {
		return err
	}

	convertedLinkage := IDList(newLinkage)
	*l = convertedLinkage
	return nil
}

func (l IDList) Len() int           { return len(l) }
func (l IDList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l IDList) Less(i, j int) bool { return l[i].ID < l[j].ID }
