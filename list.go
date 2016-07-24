package jsh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// List is a wrapper around an object slice that implements Sendable.
// List implements sort.Interface for []*Object based on the ID field.
type List []*Object

func (l List) Len() int           { return len(l) }
func (l List) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l List) Less(i, j int) bool { return l[i].ID < l[j].ID }

/*
Validate ensures that List is JSON API compatible.
*/
func (list List) Validate(r *http.Request, response bool) *Error {
	for _, object := range list {
		err := object.Validate(r, response)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
UnmarshalJSON allows us to manually decode a list via the json.Unmarshaler
interface.
*/
func (list *List) UnmarshalJSON(rawData []byte) error {
	// Create a sub-type here so when we call Unmarshal below, we don't recursively
	// call this function over and over
	type UnmarshalList List

	// if our "List" is a single object, modify the JSON to make it into a list
	// by wrapping with "[ ]"
	if rawData[0] == '{' {
		rawData = []byte(fmt.Sprintf("[%s]", rawData))
	}

	newList := UnmarshalList{}

	err := json.Unmarshal(rawData, &newList)
	if err != nil {
		return err
	}

	convertedList := List(newList)
	*list = convertedList

	return nil
}
