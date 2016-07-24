// Copyright (C) 2016 Etix Labs - All Rights Reserved.
// All information contained herein is, and remains the property of Etix Labs and its suppliers,
// if any. The intellectual and technical concepts contained herein are proprietary to Etix Labs
// Dissemination of this information or reproduction of this material is strictly forbidden unless
// prior written permission is obtained from Etix Labs.

package jsh

import (
	"reflect"
	"strings"
	"unicode"
)

// tagOptions represents the options that can be passed to JSH tags.
type tagOptions struct {
	required bool
}

// tags represents the tag options by tag name of a struct field
type tags map[string]*tagOptions

// decodeJSONTag returns the first JSON tag of the given struct field. If there is none, the field name is returned.
func decodeJSONTag(f reflect.StructField) string {
	rawTags := f.Tag.Get(tagNameJSON)
	tags := strings.SplitN(rawTags, tagSep, -1)
	if len(tags) == 0 {
		return f.Name
	}
	return tags[0]
}

// decodeFieldTags decodes all JSH tags from the struct field to a tag struct.
func decodeFieldTags(rawTags string) tags {
	var result = make(tags)
	options := strings.SplitN(rawTags, tagSep, -1)
	for _, option := range options {
		jshTag := strings.SplitN(option, optionSep, -1)
		if !isValidTag(jshTag[0]) {
			continue
		}
		options := &tagOptions{}
		if len(jshTag) == 2 {
			options.required = jshTag[1] == optionRequired
		}
		result[jshTag[0]] = options
	}
	return result
}

// decodeFieldTag decodes the JSH tag from the struct field to a tag struct.
// It returns nil if the tag was not found.
func decodeFieldTag(tags, tagName string) *tagOptions {
	result := decodeFieldTags(tags)
	return result[tagName]
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
