package jsh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/asaskevich/govalidator"
)

// Object represents the default JSON spec for objects
type Object struct {
	Type          string                   `json:"type" valid:"required"`
	ID            string                   `json:"id"`
	Attributes    json.RawMessage          `json:"attributes,omitempty"`
	Links         map[string]*Link         `json:"links,omitempty"`
	Relationships map[string]*Relationship `json:"relationships,omitempty"`
	Meta          map[string]interface{}   `json:"meta,omitempty"`
	// Status is the HTTP Status Code that should be associated with the object
	// when it is sent.
	Status int `json:"-"`
}

// NewObject prepares a new JSON Object for an API response. Whatever is provided
// as attributes will be marshalled to JSON.
func NewObject(id string, resourceType string, attributes interface{}) (*Object, *Error) {
	object := &Object{
		ID:            id,
		Type:          resourceType,
		Links:         map[string]*Link{},
		Relationships: map[string]*Relationship{},
	}
	if err := object.Marshal(attributes); err != nil {
		return nil, err
	}
	return object, nil
}

/*
Unmarshal puts an Object's Attributes into a more useful target resourceType defined
by the user. A correct object resourceType specified must also be provided otherwise
an error is returned to prevent hard to track down situations.

Optionally, used https://github.com/go-validator/validator for request input validation.
Simply define your struct with valid input tags:

	struct {
		Username string `json:"username" valid:"required,alphanum"`
	}


As the final action, the Unmarshal function will run govalidator on the unmarshal
result. If the validator fails, a Sendable error response of HTTP Status 422 will
be returned containing each validation error with a populated Error.Source.Pointer
specifying each struct attribute that failed. In this case, all you need to do is:

	errors := obj.Unmarshal("mytype", &myType)
	if errors != nil {
		// log errors via error.ISE
		jsh.Send(w, r, errors)
	}
*/
func (o *Object) Unmarshal(resourceType string, target interface{}) ErrorList {

	if resourceType != o.Type {
		return ErrorList{ConflictError(o.Type, "")}
	}

	if len(o.Attributes) == 0 {
		return nil
	}

	jsonErr := json.Unmarshal(o.Attributes, target)
	if jsonErr != nil {
		return []*Error{BadRequestError(fmt.Sprintf(
			"For type '%s' unable to unmarshal",
			resourceType,
		), jsonErr.Error())}
	}

	return validateInput(target)
}

/*
Marshal allows you to load a modified payload back into an object to preserve
all of the data it has.
*/
func (o *Object) Marshal(attributes interface{}) *Error {
	if attributes == nil {
		o.Attributes = json.RawMessage{}
		return nil
	}
	raw, err := json.MarshalIndent(attributes, "", " ")
	if err != nil {
		return ISE(fmt.Sprintf("Error marshaling attrs while creating a new JSON Object: %s", err))
	}

	o.Attributes = raw
	return nil
}

/*
Validate ensures that an object is JSON API compatible. Has a side effect of also
setting the Object's Status attribute to be used as the Response HTTP Code if one
has not already been set.
*/
func (o *Object) Validate(r *http.Request, response bool) *Error {
	if o.ID == "" {
		// don't error if the client is attempting to performing a POST request, in
		// which case, ID shouldn't actually be set
		if !response && r.Method != "POST" {
			return SpecificationError("ID must be set for Object response")
		}
	}

	if o.Type == "" {
		return SpecificationError("Type must be set for Object response")
	}

	switch r.Method {
	case "POST":
		acceptable := map[int]bool{201: true, 202: true, 204: true}

		if o.Status != 0 {
			if _, validCode := acceptable[o.Status]; !validCode {
				return SpecificationError("POST Status must be one of 201, 202, or 204.")
			}
			break
		}

		o.Status = http.StatusCreated
	case "PATCH":
		acceptable := map[int]bool{200: true, 202: true, 204: true}

		if o.Status != 0 {
			if _, validCode := acceptable[o.Status]; !validCode {
				return SpecificationError("PATCH Status must be one of 200, 202, or 204.")
			}
			break
		}

		fallthrough
	case "HEAD":
		fallthrough
	case "GET":
		o.Status = http.StatusOK
	// If we hit this it means someone is attempting to use an unsupported HTTP
	// method. Return a 406 error instead
	default:
		return SpecificationError(fmt.Sprintf(
			"The JSON Specification does not accept '%s' requests.",
			r.Method,
		))
	}

	return nil
}

// AddSelfLink creates a new self link and adds it to the resource object links.
func (o *Object) AddSelfLink() {
	o.Links["self"] = NewSelfLink(o.ID, o.Type)
}

// AddRelationshipLinks creates a new relationship link and adds it to the resource object relationships.
func (o *Object) AddRelationshipLinks(name string) {
	o.Relationships[name] = &Relationship{
		Links: NewRelationshipLinks(o.ID, o.Type, name),
	}
}

// AddRelationshipOne sets the resource linkage of the resource object for the given to-one relationship.
func (o *Object) AddRelationshipOne(name string, linkage *IDObject) {
	o.Relationships[name] = &Relationship{
		Data: IDList{linkage},
	}
}

// AddRelationshipMany sets the resource linkage of the resource object for the given to-many relationship.
func (o *Object) AddRelationshipMany(name string, linkage IDList) {
	o.Relationships[name] = &Relationship{
		Data: linkage,
	}
}

/*
ProcessCreate unmarshals the object to the given struct (see Object.Unmarshal) and uses JSH tags
to validate that there is no missing attributes or forbidden ones.

Simply define your struct with jsh tags to allow for the model to be created with the tagged attributes.

	struct {
		Username string `json:"username" jsh:"create"`
	}

You can also add a required option to ensure a specific attribute is non-zero.

	struct {
		Username string `json:"username" jsh:"create/required"`
	}

The model must be a non-nil pointer to a struct.
If valid, the model contains the valid request attributes after the call (even on validation error).
Relationship fields, if any, are set to the IDObject values in object.Relationships.

See the documentation of Validator.Validate for more detailed information.

The string slice returned contains the names of the attributes and relationships
that were unmarshaled to the model.
*/
func (o *Object) ProcessCreate(resourceType string, model interface{}) ([]string, ErrorList) {
	return o.process(tagCreate, resourceType, model)
}

// ProcessUpdate behaves just like ProcessCreate but uses the update tag for validation.
// It also adds the constraint of requiring at least one field to be updated.
func (o *Object) ProcessUpdate(resourceType string, model interface{}) ([]string, ErrorList) {
	attrs, err := o.process(tagUpdate, resourceType, model)
	if err != nil {
		return nil, err
	}
	// Return 400 if no attributes were provided.
	if len(attrs) == 0 {
		return nil, ErrorList{BadRequestError("Invalid patch document", "Missing description of changes")}
	}
	return attrs, nil
}

// ToIDObject returns a resource identifier object created with the object type and ID.
func (o *Object) ToIDObject() *IDObject {
	return NewIDObject(o.Type, o.ID)
}

// String prints a formatted string representation of the object
func (o *Object) String() string {
	raw, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return err.Error()
	}

	return string(raw)
}

// process validates that the object's attributes are valid for the given action.
// It unmarshals the attributes to the model's fields that are tagged with the action.
func (o *Object) process(action, resourceType string, model interface{}) ([]string, ErrorList) {
	// Unmarshal to model and validates input against govalidator rules
	err := o.Unmarshal(resourceType, model)
	if err != nil {
		return nil, err
	}
	// Look for missing/forbidden attributes and relationships for action
	return NewValidator(o, action).Validate(model)
}


// validateInput runs go-validator on each attribute of the struct and returns all errors.
func validateInput(target interface{}) ErrorList {
	adapter := func(err govalidator.Error) *Error {
		return InputError(err.Err.Error(), toLowerFirstRune(err.Name))
	}
	return validator(target, adapter)
}

// validator runs go-validator on each attribute of the struct and
// converts the errors to jsh errors by using the provided adapter.
// It returns all errors that it picks up converted.
func validator(target interface{}, adapter func(govalidator.Error) *Error) ErrorList {
	_, errors := govalidator.ValidateStruct(target)
	if errors != nil {
		errorList, ok := errors.(govalidator.Errors)
		if ok {
			var errors ErrorList
			for _, err := range errorList.Errors() {
				// parse out validation error
				validatorErr, _ := err.(govalidator.Error)
				errors = append(errors, adapter(validatorErr))
			}
			return errors
		}
	}
	return nil
}
