package jsh

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

const (
	tagNameJSON    = "json"
	tagIgnore      = "-"
	tagNameJSH     = "jsh"
	tagSep         = ","
	tagCreate      = "create"
	tagUpdate      = "update"
	optionSep      = "/"
	optionRequired = "required"
)

// tagOptions represents the options that can be passed to JSH tags
type tagOptions struct {
	required bool
}

// validateStruct validates that the given struct has no missing or forbidden
// fields for the jsh action (i.e. create, update).
func validateStruct(rv reflect.Value, tag string) ([]string, ErrorList) {
	var fields []string
	var errors ErrorList
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		v := rv.Field(i)
		f := rt.Field(i)
		// Ignore fields ignored by json.Unmarshal, zero reflect.Value and private fields
		if hasIgnoreTag(f.Tag) || !v.IsValid() || !v.CanSet() {
			continue
		}
		hasValue, err := validateField(f, v, tag)
		if err != nil {
			errors = append(errors, err)
		} else if hasValue {
			fields = append(fields, f.Name)
		}
	}
	if errors != nil {
		return nil, errors
	}
	return fields, nil
}

// validateField validates that the value for the given field
// is neither missing or forbidden according to jsh tags.
func validateField(f reflect.StructField, v reflect.Value, tag string) (bool, *Error) {
	opts := decodeTag(f.Tag.Get(tagNameJSH), tag)
	// Check if attribute was not provided
	if isZero(v) {
		if opts != nil && opts.required {
			return false, InputError("Required attribute", f.Name)
		}
		return false, nil
	}
	// The attribute was provided: it must have jsh tag
	if opts == nil {
		err := ForbiddenError("Attribute not allowed")
		err.Source = &ErrorSource{
			Pointer: fmt.Sprintf("/data/attributes/%s", strings.ToLower(f.Name)),
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

// hasIgnoreTag returns true if the given struct tag has the JSON ignore tag "-".
func hasIgnoreTag(tag reflect.StructTag) bool {
	return tag.Get(tagNameJSON) == tagIgnore
}

// decodeTag decodes the tag from the struct field to a tagOptions struct.
// It returns nil if the tag was not found.
func decodeTag(tags, tag string) *tagOptions {
	options := strings.SplitN(tags, tagSep, -1)
	for _, option := range options {
		jshTag := strings.SplitN(option, optionSep, -1)
		if !isValidTag(jshTag[0]) {
			continue
		}
		if jshTag[0] == tag {
			if len(jshTag) == 2 {
				return &tagOptions{jshTag[1] == optionRequired}
			} else {
				return &tagOptions{false}
			}
		}
	}
	return nil
}

// isValidTag returns false if the tag is empty or contains invalid characters.
func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}
