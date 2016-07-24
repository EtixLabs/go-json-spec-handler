package jsh

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Validator provides validation features for resource modeling.
type Validator struct {
	object *Object
	action string
}

// NewValidator returns a new instance of a JSH validator.
func NewValidator(obj *Object, action string) *Validator {
	return &Validator{
		object: obj,
		action: action,
	}
}

// stringValues is a slice of reflect.Value holding *reflect.StringValue.
// It implements sort.Interface to sort by string.
type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

/*
Validate validates that the given struct has no missing/forbidden
fields and relationships for the jsh action (i.e. create, update) according to JSH rules.

Customers must use JSH tags "create" and "update" to allow each field to be provided for create and update requests.
An optional "/required" can be added to the tag to require the field to be provided.

Additionally, relationships fields must fulfill the following requirements:
	- The field must be tagged "one" or "many".
	- The JSON tag of the field should be "-" to prevent it from being included in attributes.
	- If the relationship is tagged "one", the type of the field must be *jsh.IDObject.
	- If the relationship is tagged "many", the type of the field must be either:
		map[int]*jsh.IDObject or map[int]*jsh.IDObject. The map must be non-nil.

Example model:

	type User struct {
		Group *jsh.IDObject `json:"-"        jsh:"one,create,update"`
		Name  string        `json:"username" jsh:"create/required"`
		Email string        `json:"email"    jsh:"create,update"`
	}

The given model should be the result of the unmarshaling of the internal object attributes.
The validator will automatically update the relationship fields during validation.
*/
func (v *Validator) Validate(model interface{}) ([]string, ErrorList) {
	// Check argument is a non-nil pointer
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, ErrorList{ISE(fmt.Sprintf("The argument to %s must be a non-nil pointer", v.action))}
	}
	// Get pointer element
	rv = rv.Elem()
	// Check pointer element is a struct
	if rv.Kind() != reflect.Struct {
		return nil, ErrorList{ISE(fmt.Sprintf("The argument to %s must be a pointer to a struct", v.action))}
	}
	// Unmarshal to map to retrieve all provided attributes
	return v.validateStruct("", rv, v.object.Attributes)
}

// validateStruct validates all fields of the given struct according to JSH rules.
func (v *Validator) validateStruct(path string, rv reflect.Value, j json.RawMessage) ([]string, ErrorList) {
	// Decode struct provided attributes
	attrs, err := v.decodeKeys(j)
	if err != nil {
		return nil, ErrorList{err}
	}
	// Validate fields
	var fields []string
	var errors ErrorList
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		f := rt.Field(i)
		// Ignore zero reflect.Value and private fields
		if !fv.IsValid() || f.PkgPath != "" {
			continue
		}
		// Decode JSH tags
		tags := decodeFieldTags(f.Tag.Get(tagNameJSH))
		p := toLowerFirstRune(f.Name)
		_, one := tags[tagToOne]
		_, many := tags[tagToMany]
		if one || many {
			// Remove existing field from the relationships map
			var rel *Relationship
			for name, r := range v.object.Relationships {
				if strings.EqualFold(name, f.Name) {
					rel = r
					delete(v.object.Relationships, name)
					break
				}
			}
			// Validate relationship
			hasValue, err := validateModelRelationship(p, many, rel, tags[v.action])
			if err != nil {
				errors = append(errors, err)
			} else if hasValue {
				// Set relationship in model
				if err := setModelRelationship(p, many, rel, fv); err != nil {
					errors = append(errors, err)
				} else {
					fields = append(fields, p)
				}
			}
			continue
		}
		// Ignore fields ignored by json.Unmarshal
		p = decodeJSONTag(f)
		if p == "-" {
			continue
		}
		// Remove existing field from the provided attributes map
		var jValue json.RawMessage
		for name, value := range attrs {
			if strings.EqualFold(name, p) {
				jValue = value
				delete(attrs, name)
				break
			}
		}
		// Validate field
		if path != "" {
			p = path + fieldSep + p
		}
		hasValue, err := validateModelField(p, fv, tags[v.action])
		if err != nil {
			errors = append(errors, err)
		} else if hasValue {
			result, errlist := v.nestedResult(p, fv, jValue)
			if errlist != nil {
				errors = append(errors, errlist...)
			} else {
				fields = append(fields, result...)
			}
		}
	}
	// Add errors for non-existent attributes
	for name := range attrs {
		if path != "" {
			name = path + fieldSep + name
		}
		errors = append(errors, InputError("Attribute does not exist", name))
	}
	// Add errors for non-existent relationships
	for name := range v.object.Relationships {
		if path != "" {
			name = path + fieldSep + name
		}
		errors = append(errors, RelationshipError("Relationship does not exist", name))
	}
	if errors != nil {
		return nil, errors
	}
	return fields, nil
}

// nestedResult recurses until the field type is not a map, slice, array, interface, pointer or struct.
// It calls validateStruct recursively if it encounters a struct type.
func (v *Validator) nestedResult(path string, fv reflect.Value, jValue json.RawMessage) ([]string, ErrorList) {
	fields := []string{path}
	switch fv.Kind() {
	case reflect.Map:
		// Reject unsupported key types
		if fv.Type().Key().Kind() != reflect.String {
			return nil, ErrorList{ISE(fmt.Sprintf("Type %v is not supported", fv.Type()))}
		}
		// Decode map JSON values
		jsonValues, err := v.decodeKeys(jValue)
		if err != nil {
			return nil, ErrorList{err}
		}
		// Sort map by key
		var sv stringValues
		sv = fv.MapKeys()
		sort.Sort(sv)
		// Validate embedded struct values recursively and append field names
		for _, k := range sv {
			key := k.String()
			p := path + fieldSep + key
			result, errlist := v.nestedResult(p, fv.MapIndex(k), jsonValues[key])
			if errlist != nil {
				return nil, errlist
			}
			fields = append(fields, result...)
		}
	case reflect.Slice:
		if fv.Type() == reflect.TypeOf(json.RawMessage{}) {
			break
		}
		fallthrough
	case reflect.Array:
		// Decode JSON array to slice
		jArray, err := v.decodeSlice(jValue)
		if err != nil {
			return nil, ErrorList{err}
		}
		// Validate embedded struct values recusively and append field names
		for i := 0; i < fv.Len(); i++ {
			p := path + fieldSep + strconv.Itoa(i)
			result, errlist := v.nestedResult(p, fv.Index(i), jArray[i])
			if errlist != nil {
				return nil, errlist
			}
			fields = append(fields, result...)
		}
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		// If the value is an pointer or interface then encode its element
		if fv.IsNil() {
			break
		}
		return v.nestedResult(path, fv.Elem(), jValue)
	case reflect.Struct:
		// Validate embedded struct values recusively and append field names
		result, errlist := v.validateStruct(path, fv, jValue)
		if errlist != nil {
			return nil, errlist
		}
		fields = append(fields, result...)
	default:
	}
	return fields, nil
}

// decodeKeys decodes the keys of the given JSON message.
func (v *Validator) decodeKeys(j json.RawMessage) (map[string]json.RawMessage, *Error) {
	attrs := make(map[string]json.RawMessage)
	if len(j) > 0 {
		err := json.Unmarshal(j, &attrs)
		if err != nil {
			return nil, ISE(err.Error())
		}
	}
	return attrs, nil
}

// decodeSlice decodes the given JSON array into a slice of JSON messages.
func (v *Validator) decodeSlice(j json.RawMessage) ([]json.RawMessage, *Error) {
	var attrs []json.RawMessage
	if len(j) > 0 {
		err := json.Unmarshal(j, &attrs)
		if err != nil {
			return nil, ISE(err.Error())
		}
	}
	return attrs, nil
}

// setModelRelationship sets the given field (v) of the model to the given relationship value.
func setModelRelationship(name string, many bool, rel *Relationship, v reflect.Value) *Error {
	if many {
		return setModelRelationshipMany(name, v, rel)
	} else {
		return setModelRelationshipOne(v, rel)
	}
}

// setModelRelationshipOne sets the given field (v) of the model to the given to-one relationship value.
// The struct field must be of type *IDObject.
func setModelRelationshipOne(v reflect.Value, rel *Relationship) *Error {
	one := reflect.ValueOf(rel.Data[0])
	if !one.Type().AssignableTo(v.Type()) {
		return ISE("Invalid field type for to-one relation, must be *IDObject")
	}
	v.Set(one)
	return nil
}

// setModelRelationshipMany sets the given field (v) of the model to the given to-many relationship value.
// The struct field must be of type map[string]*IDObject or map[int]*IDObject.
func setModelRelationshipMany(name string, v reflect.Value, rel *Relationship) *Error {
	kind := v.Type().Kind()
	if kind != reflect.Map {
		return ISE("Invalid field type for to-many relation, must be map")
	}
	keyKind := v.Type().Key().Kind()
	if keyKind == reflect.String || keyKind == reflect.Int {
		for _, data := range rel.Data {
			dv := reflect.ValueOf(data)
			if !dv.Type().AssignableTo(v.Type().Elem()) {
				return ISE("Invalid map value type for to-many relation, must be *IDObject")
			}
			if keyKind == reflect.String {
				v.SetMapIndex(reflect.ValueOf(data.ID), dv)
			} else {
				id, err := strconv.Atoi(data.ID)
				if err != nil {
					return RelationshipError("Invalid resource ID", toLowerFirstRune(name))
				}
				v.SetMapIndex(reflect.ValueOf(id), dv)
			}
		}
	} else {
		return ISE("Invalid map key type for to-many relation, must be string or int")
	}
	return nil
}

// validateModelRelationship validates that the given struct has no forbidden or invalid
// relationships for the jsh action (i.e. create, update).
func validateModelRelationship(name string, many bool, rel *Relationship, opts *tagOptions) (bool, *Error) {
	// Check if relationship was not provided
	if rel == nil {
		if opts != nil && opts.required {
			return false, RelationshipError("Required relationship", toLowerFirstRune(name))
		}
		return false, nil
	}
	// Check if relationship has data
	if len(rel.Data) == 0 {
		return false, RelationshipError("Missing relationship data", toLowerFirstRune(name))
	}
	if !many && len(rel.Data) > 1 {
		return false, RelationshipError("Multiple objects for to-one relation", toLowerFirstRune(name))
	}
	// The relationship was provided: it must have jsh tag
	if opts == nil {
		err := ForbiddenError("Operation not allowed")
		err.Source = &ErrorSource{
			Pointer: RelationshipPointer(toLowerFirstRune(name)),
		}
		return false, err
	}
	return true, nil
}

// validateModelField validates that the value for the given field
// is neither missing or forbidden according to jsh tags.
func validateModelField(path string, v reflect.Value, opts *tagOptions) (bool, *Error) {
	// Check if attribute was not provided
	if isZero(v) {
		if opts != nil && opts.required {
			return false, InputError("Required attribute", toLowerFirstRune(path))
		}
		return false, nil
	}
	// The attribute was provided: it must have jsh tag
	if opts == nil {
		err := ForbiddenError("Operation not allowed")
		err.Source = &ErrorSource{
			Pointer: AttributePointer(toLowerFirstRune(path)),
		}
		return false, err
	}
	return true, nil
}

// isZero checks if the given value is the zero value of its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}
