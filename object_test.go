package jsh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestObject(t *testing.T) {

	Convey("Object Tests", t, func() {

		testType := "tests"
		testObject := &Object{
			ID:            "ID123",
			Type:          testType,
			Attributes:    json.RawMessage(`{"foo":"bar"}`),
			Links:         map[string]*Link{},
			Relationships: map[string]*Relationship{},
		}

		request := &http.Request{}

		Convey("->NewObject()", func() {

			Convey("should create a new object with no attributes", func() {
				newObj, err := NewObject(testObject.ID, testObject.Type, nil)
				So(err, ShouldBeNil)
				So(newObj.Attributes, ShouldBeEmpty)
			})

			Convey("should create a new object with populated attrs", func() {
				attrs := struct {
					Foo string `json:"foo"`
				}{"bar"}

				newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
				So(err, ShouldBeNil)
				So(newObj.Attributes, ShouldNotBeEmpty)
			})
		})

		Convey("->ProcessCreate()", func() {

			Convey("Should reject nil interface", func() {
				f, err := testObject.ProcessCreate(testType, nil)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should reject non-pointer type", func() {
				f, err := testObject.ProcessCreate(testType, "invalid type")
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should reject pointers to non-struct types", func() {
				testConversion := "non-struct type"

				f, err := testObject.ProcessCreate(testType, &testConversion)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			// Relationship tests
			Convey("Relationships", func() {
				testObject.Attributes = json.RawMessage{}
				foo := testObject.ToIDObject()
				testObject.AddRelationshipOne("foo", foo)
				testObject.AddRelationshipMany("foos", IDList{foo})

				Convey("Should reject relationships with no relationship tag", func() {
					testConversion := struct {
						Foo  *IDObject            `json:"-" jsh:"create"`
						Foos map[string]*IDObject `json:"-" jsh:"create"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 2)
					for _, e := range err {
						So(e.StatusCode(), ShouldEqual, 422)
						So(e.Source, ShouldNotBeNil)
						So(e.Source.Pointer, ShouldStartWith, "/data/relationships/foo")
					}
					So(f, ShouldBeNil)
				})

				Convey("Should detect missing and required relationships", func() {
					testConversion := struct {
						Foo  *IDObject            `json:"-" jsh:"one"`
						Bars map[string]*IDObject `json:"-" jsh:"many,create/required"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 3)
					So(err[0].StatusCode(), ShouldEqual, 403)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/relationships/foo")
					So(err[1].StatusCode(), ShouldEqual, 422)
					So(err[1].Source, ShouldNotBeNil)
					So(err[1].Source.Pointer, ShouldEqual, "/data/relationships/bars")
					So(err[2].StatusCode(), ShouldEqual, 422)
					So(err[2].Source, ShouldNotBeNil)
					So(err[2].Source.Pointer, ShouldEqual, "/data/relationships/foos")
					So(f, ShouldBeNil)
				})

				Convey("Should detect invalid linkage data", func() {
					// Override relationships with invalid ones
					testObject.AddRelationshipMany("foo", IDList{foo, foo})
					testObject.AddRelationshipMany("foos", nil)
					testConversion := struct {
						Foo  *IDObject            `json:"-" jsh:"one,create"`
						Foos map[string]*IDObject `json:"-" jsh:"many,create"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 2)
					for _, e := range err {
						So(e.StatusCode(), ShouldEqual, 422)
						So(e.Source, ShouldNotBeNil)
						So(e.Source.Pointer, ShouldStartWith, "/data/relationships/foo")
					}
					So(f, ShouldBeNil)
				})

				Convey("Should reject relationships with invalid field type", func() {
					testObject.AddRelationshipOne("bar", foo)
					testObject.AddRelationshipMany("bars", IDList{foo})
					testObject.AddRelationshipMany("jons", IDList{foo})
					testConversion := struct {
						Foo  string            `json:"-" jsh:"one,create"`
						Bar  IDObject          `json:"-" jsh:"one,create"`
						Foos map[string]string `json:"-" jsh:"many,create"`
						Bars *IDObject         `json:"-" jsh:"many,create"`
						Jons map[int]*IDObject `json:"-" jsh:"many,create"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 5)
					for i, e := range err {
						if i == len(err)-1 {
							So(e.StatusCode(), ShouldEqual, 422)
						} else {
							So(e.StatusCode(), ShouldEqual, 500)
						}
					}
					So(f, ShouldBeNil)
				})

				Convey("Should accept and set relationships with a relationship tag", func() {
					other := NewIDObject(testType, "1")
					testObject.AddRelationshipMany("bars", IDList{other})
					testConversion := struct {
						Foo  *IDObject            `json:"-" jsh:"one,create"`
						Bar  *IDObject            `json:"-" jsh:"one,create"`
						Foos map[string]*IDObject `json:"-" jsh:"many,create"`
						Bars map[int]*IDObject    `json:"-" jsh:"many,create"`
					}{
						Foos: make(map[string]*IDObject),
						Bars: make(map[int]*IDObject),
					}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldHaveLength, 3)
					So(f, ShouldContain, "foo")
					So(f, ShouldContain, "foos")
					So(f, ShouldContain, "bars")
					So(testConversion.Foo.ID, ShouldEqual, foo.ID)
					So(testConversion.Foo.Type, ShouldEqual, foo.Type)
					So(testConversion.Foos, ShouldHaveLength, 1)
					So(testConversion.Foos, ShouldContainKey, foo.ID)
					So(testConversion.Foos[foo.ID].ID, ShouldEqual, foo.ID)
					So(testConversion.Foos[foo.ID].Type, ShouldEqual, foo.Type)
					id, _ := strconv.Atoi(other.ID)
					So(testConversion.Bars, ShouldHaveLength, 1)
					So(testConversion.Bars, ShouldContainKey, id)
					So(testConversion.Bars[id].ID, ShouldEqual, other.ID)
					So(testConversion.Bars[id].Type, ShouldEqual, other.Type)
				})

				Convey("Should accept and set nested relationships with a relationship tag", func() {
					testConversion := struct {
						Bars struct {
							Foo  *IDObject            `json:"-" jsh:"one,create"`
							Foos map[string]*IDObject `json:"-" jsh:"many,create"`
						} `json:"bars" jsh:"create"`
					}{}
					testConversion.Bars.Foos = make(map[string]*IDObject)

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldHaveLength, 3)
					So(f, ShouldContain, "foo")
					So(f, ShouldContain, "foos")
					So(testConversion.Bars.Foo.ID, ShouldEqual, foo.ID)
					So(testConversion.Bars.Foo.Type, ShouldEqual, foo.Type)
					So(testConversion.Bars.Foos, ShouldHaveLength, 1)
					So(testConversion.Bars.Foos, ShouldContainKey, foo.ID)
					So(testConversion.Bars.Foos[foo.ID].ID, ShouldEqual, foo.ID)
					So(testConversion.Bars.Foos[foo.ID].Type, ShouldEqual, foo.Type)
				})

			})

			// Attribute tests
			Convey("Attributes", func() {

				Convey("Should reject request with invalid attributes", func() {
					testObject.Attributes = json.RawMessage(`{"invalid":"bar"}`)
					testConversion := struct {
						Foo string `json:"foo"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 1)
					So(err[0].StatusCode(), ShouldEqual, 422)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/invalid")
					So(f, ShouldBeNil)
				})

				Convey("Should reject attributes with no jsh create tag", func() {
					testConversion := struct {
						Foo string `json:"foo"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 1)
					So(err[0].StatusCode(), ShouldEqual, http.StatusForbidden)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
					So(f, ShouldBeNil)
				})

				Convey("Should reject attributes with invalid jsh tag", func() {
					testConversion := struct {
						Foo string `json:"foo" jsh:"invalid"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 1)
					So(err[0].StatusCode(), ShouldEqual, http.StatusForbidden)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
					So(f, ShouldBeNil)
				})

				Convey("Should ignore attributes with json '-' tag", func() {
					testConversion := struct {
						Foo string `json:"-"`
					}{
						Foo: "shouldNotBeOverriden",
					}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 1)
					So(err[0].StatusCode(), ShouldEqual, 422)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
					So(f, ShouldBeNil)
					So(testConversion.Foo, ShouldEqual, "shouldNotBeOverriden")
				})

				Convey("Should accept attributes with jsh 'create' tag", func() {
					testConversion := struct {
						Foo string `json:"foo" jsh:"create"`
					}{
						Foo: "shouldBeOverriden",
					}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldContain, "foo")
					So(testConversion.Foo, ShouldEqual, "bar")
				})

				Convey("Should accept attributes with specific JSON tag", func() {
					testConversion := struct {
						Other string `json:"foo" jsh:"create"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldContain, "foo")
					So(testConversion.Other, ShouldEqual, "bar")
				})

				Convey("Should ignore private attributes", func() {
					testConversion := struct {
						foo string `json:"foo" jsh:"create"`
					}{
						foo: "shouldNotBeOverriden",
					}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 1)
					So(err[0].StatusCode(), ShouldEqual, 422)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
					So(f, ShouldBeNil)
					So(testConversion.foo, ShouldEqual, "shouldNotBeOverriden")
				})

				Convey("Should reject attributes with missing required field", func() {
					testConversion := struct {
						Foo string `json:"foo" jsh:"create/required"`
						Bar string `json:"bar" jsh:"create/required"`
					}{}

					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 1)
					So(err[0].StatusCode(), ShouldEqual, 422)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/bar")
					So(f, ShouldBeNil)
				})
			})

			// Nested attributes tests
			Convey("Nested attributes", func() {

				Convey("Should accept and validate nested struct", func() {
					testConversion := struct {
						Foo struct {
							Foo string `json:"foo" jsh:"create"`
							Bar string `json:"bar" jsh:"create"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":"bar", "bar":"foo"}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/foo", "foo/bar"})
					So(testConversion.Foo.Foo, ShouldEqual, "bar")
					So(testConversion.Foo.Bar, ShouldEqual, "foo")
				})

				Convey("Should accept and validate embedded struct", func() {
					type Embedded struct {
						Foo string `json:"foo" jsh:"create"`
						Bar string `json:"bar" jsh:"create"`
					}
					testConversion := struct {
						Embedded `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":"bar", "bar":"foo"}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/foo", "foo/bar"})
					So(testConversion.Foo, ShouldEqual, "bar")
					So(testConversion.Bar, ShouldEqual, "foo")
				})

				Convey("Should accept and validate nested struct pointer", func() {
					testConversion := struct {
						Foo *struct {
							Foo string `json:"foo" jsh:"create"`
							Bar string `json:"bar" jsh:"create"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":"bar", "bar":"foo"}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/foo", "foo/bar"})
					So(testConversion.Foo.Foo, ShouldEqual, "bar")
					So(testConversion.Foo.Bar, ShouldEqual, "foo")
				})

				Convey("Should reject nested struct with invalid or missing required fields", func() {
					testConversion := struct {
						Foo struct {
							Foo string `json:"foo"`
							Bar string `json:"bar" jsh:"create/required"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":"bar","invalid":"test"}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 3)
					So(err[0].StatusCode(), ShouldEqual, 403)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo/foo")
					So(err[1].StatusCode(), ShouldEqual, 422)
					So(err[1].Source, ShouldNotBeNil)
					So(err[1].Source.Pointer, ShouldEqual, "/data/attributes/foo/bar")
					So(err[2].StatusCode(), ShouldEqual, 422)
					So(err[2].Source, ShouldNotBeNil)
					So(err[2].Source.Pointer, ShouldEqual, "/data/attributes/foo/invalid")
					So(f, ShouldBeNil)
				})

				// Nested interface tests

				Convey("Should accept and validate nested interfaces", func() {
					type fooType struct {
						Foo string `json:"foo" jsh:"create"`
						Bar string `json:"bar" jsh:"create"`
					}
					nested := fooType{
						Foo: "shouldBeOverriden",
					}
					testConversion := struct {
						Foo interface{} `json:"foo" jsh:"create"`
					}{
						Foo: &nested,
					}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":"bar", "bar":"foo"}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/foo", "foo/bar"})
					foo, ok := testConversion.Foo.(*fooType)
					So(ok, ShouldBeTrue)
					So(foo.Foo, ShouldEqual, "bar")
					So(foo.Bar, ShouldEqual, "foo")
				})

				Convey("Should accept nil interface", func() {
					testConversion := struct {
						Foo interface{} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":"bar", "bar":"foo"}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldContain, "foo")
					So(f, ShouldContain, "foo/foo")
					So(f, ShouldContain, "foo/bar")
					So(testConversion.Foo, ShouldNotBeNil)
				})

				// Nested slice tests

				Convey("Should accept and validate nested slice", func() {
					testConversion := struct {
						Foo []struct {
							Foo string `json:"foo" jsh:"create"`
							Bar string `json:"bar" jsh:"create"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":[{"foo":"bar", "bar":"foo"}]}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/0", "foo/0/foo", "foo/0/bar"})
					So(testConversion.Foo, ShouldHaveLength, 1)
					So(testConversion.Foo[0].Foo, ShouldEqual, "bar")
					So(testConversion.Foo[0].Bar, ShouldEqual, "foo")
				})

				Convey("Should accept and validate nested slice of pointers", func() {
					testConversion := struct {
						Foo []*struct {
							Foo string `json:"foo" jsh:"create"`
							Bar string `json:"bar" jsh:"create"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":[{"foo":"bar", "bar":"foo"}]}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/0", "foo/0/foo", "foo/0/bar"})
					So(testConversion.Foo, ShouldHaveLength, 1)
					So(testConversion.Foo[0].Foo, ShouldEqual, "bar")
					So(testConversion.Foo[0].Bar, ShouldEqual, "foo")
				})

				Convey("Should accept nested slice of primitive type", func() {
					testConversion := struct {
						Foo []string `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":["foo","bar"]}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/0", "foo/1"})
					So(testConversion.Foo, ShouldHaveLength, 2)
					So(testConversion.Foo[0], ShouldEqual, "foo")
					So(testConversion.Foo[1], ShouldEqual, "bar")
				})

				Convey("Should reject nested slice with invalid or missing required fields", func() {
					testConversion := struct {
						Foo []struct {
							Foo string `json:"foo"`
							Bar string `json:"bar" jsh:"create/required"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":[{"foo":"bar","invalid":"test"}]}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 3)
					So(err[0].StatusCode(), ShouldEqual, 403)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo/0/foo")
					So(err[1].StatusCode(), ShouldEqual, 422)
					So(err[1].Source, ShouldNotBeNil)
					So(err[1].Source.Pointer, ShouldEqual, "/data/attributes/foo/0/bar")
					So(err[2].StatusCode(), ShouldEqual, 422)
					So(err[2].Source, ShouldNotBeNil)
					So(err[2].Source.Pointer, ShouldEqual, "/data/attributes/foo/0/invalid")
					So(f, ShouldBeNil)
				})

				// Nested map tests

				Convey("Should accept and validate nested map", func() {
					type fooType struct {
						Foo string `json:"foo" jsh:"create"`
						Bar string `json:"bar" jsh:"create"`
					}
					testConversion := struct {
						Foo map[string]fooType `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"key":{"foo":"bar","bar":"foo"}}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/key", "foo/key/foo", "foo/key/bar"})
					So(testConversion.Foo, ShouldHaveLength, 1)
					elem, ok := testConversion.Foo["key"]
					So(ok, ShouldBeTrue)
					So(elem.Foo, ShouldEqual, "bar")
					So(elem.Bar, ShouldEqual, "foo")
				})

				Convey("Should accept and validate nested map of pointers", func() {
					testConversion := struct {
						Foo map[string]*struct {
							Foo string `json:"foo" jsh:"create"`
							Bar string `json:"bar" jsh:"create"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"key":{"foo":"bar","bar":"foo"}}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldResemble, []string{"foo", "foo/key", "foo/key/foo", "foo/key/bar"})
					So(testConversion.Foo, ShouldHaveLength, 1)
					elem, ok := testConversion.Foo["key"]
					So(ok, ShouldBeTrue)
					So(elem.Foo, ShouldEqual, "bar")
					So(elem.Bar, ShouldEqual, "foo")
				})

				Convey("Should accept and validate nested map of primitive type", func() {
					testConversion := struct {
						Foo map[string]int `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":1,"bar":2}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(f, ShouldContain, "foo")
					So(f, ShouldContain, "foo/foo")
					So(f, ShouldContain, "foo/bar")
					So(testConversion.Foo, ShouldHaveLength, 2)
					So(testConversion.Foo["foo"], ShouldEqual, 1)
					So(testConversion.Foo["bar"], ShouldEqual, 2)
				})

				Convey("Should reject nested map with non-string key type", func() {
					testConversion := struct {
						Foo map[int]int `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"foo":1,"bar":2}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(f, ShouldBeEmpty)
				})

				Convey("Should reject nested map with invalid or missing required fields", func() {
					testConversion := struct {
						Foo map[string]*struct {
							Foo string `json:"foo"`
							Bar string `json:"bar" jsh:"create/required"`
						} `json:"foo" jsh:"create"`
					}{}

					testObject.Attributes = json.RawMessage(`{"foo":{"key":{"foo":"bar","invalid":"test"}}}`)
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveLength, 3)
					So(err[0].StatusCode(), ShouldEqual, 403)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo/key/foo")
					So(err[1].StatusCode(), ShouldEqual, 422)
					So(err[1].Source, ShouldNotBeNil)
					So(err[1].Source.Pointer, ShouldEqual, "/data/attributes/foo/key/bar")
					So(err[2].StatusCode(), ShouldEqual, 422)
					So(err[2].Source, ShouldNotBeNil)
					So(err[2].Source.Pointer, ShouldEqual, "/data/attributes/foo/key/invalid")
					So(f, ShouldBeNil)
				})
			})

			// Zero/non-zero value tests

			Convey("Should detect zero/non-zero value for all type kinds", func() {
				testConversion := struct {
					Foo       string         `json:"foo" jsh:"create"`
					Ptr       *string        `json:"ptr"`
					Array     [2]int         `json:"array"`
					Slice     []int          `json:"slice"`
					Map       map[string]int `json:"map"`
					Interface interface{}    `json:"interface"`
					Struct    struct {
						B bool `json:"b"`
						N int  `json:"n"`
					} `json:"struct"`
				}{}

				Convey("Should detect and ignore zero value for all type kinds", func() {
					f, err := testObject.ProcessCreate(testType, &testConversion)
					So(err, ShouldBeNil)
					So(testConversion.Foo, ShouldEqual, "bar")
					So(f, ShouldContain, "foo")
				})

				Convey("Should detect non-zero value for all type kinds", func() {
					var errlist ErrorList
					var attr string

					Convey("Should detect non-zero value for pointers", func() {
						str := "bar"
						attrs := struct {
							Ptr *string `json:"ptr"`
						}{&str}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate(testType, &testConversion)
						attr = "ptr"
					})

					Convey("Should detect non-zero value for arrays", func() {
						array := [2]int{0, 1}
						attrs := struct {
							Array [2]int `json:"array"`
						}{array}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate(testType, &testConversion)
						attr = "array"
					})

					Convey("Should detect non-zero value for slices", func() {
						slice := []int{0, 1}
						attrs := struct {
							Slice []int `json:"slice"`
						}{slice}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate(testType, &testConversion)
						attr = "slice"
					})

					Convey("Should detect non-zero value for maps", func() {
						m := map[string]int{"a": 0, "b": 1}
						attrs := struct {
							Map map[string]int `json:"map"`
						}{m}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate(testType, &testConversion)
						attr = "map"
					})

					Convey("Should detect non-zero value for interfaces", func() {
						i := struct{}{}
						attrs := struct {
							Interface interface{} `json:"interface"`
						}{i}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate(testType, &testConversion)
						attr = "interface"
					})

					So(errlist, ShouldNotBeNil)
					So(errlist, ShouldHaveLength, 1)
					So(errlist[0].StatusCode(), ShouldEqual, http.StatusForbidden)
					So(errlist[0].Source, ShouldNotBeNil)
					So(errlist[0].Source.Pointer, ShouldEqual, "/data/attributes/"+attr)
				})
			})
		})

		// ProcessUpdate has the same requirements than ProcessCreate and requires that at least one field be updated
		Convey("->ProcessUpdate()", func() {

			Convey("Should reject requests with no valid updated attributes", func() {
				f, err := testObject.ProcessUpdate(testType, &struct{}{})
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should accept multiple attributes with jsh 'update' tag", func() {
				testConversion := [2]struct {
					Foo string  `json:"foo" jsh:"update"`
					Bar *string `json:"bar" jsh:"update"`
				}{}
				str := "foo"
				testConversion[0].Bar = &str
				testConversion[0].Foo = "bar"
				newObj, err := NewObject(testObject.ID, testObject.Type, &testConversion[0])
				So(err, ShouldBeNil)

				f, errlist := newObj.ProcessUpdate(testType, &testConversion[1])
				So(errlist, ShouldBeNil)
				So(f, ShouldHaveLength, 2)
				So(f, ShouldContain, "foo")
				So(f, ShouldContain, "bar")
			})

			Convey("Should reject attributes with missing required field", func() {
				testConversion := struct {
					Foo string `json:"foo" jsh:"update/required"`
					Bar string `json:"bar" jsh:"update/required"`
				}{}

				f, err := testObject.ProcessUpdate(testType, &testConversion)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(err[0].StatusCode(), ShouldEqual, 422)
				So(err[0].Source, ShouldNotBeNil)
				So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/bar")
				So(f, ShouldBeNil)
			})
		})

		Convey("->Unmarshal()", func() {

			Convey("non-govalidator structs", func() {

				testConversion := struct {
					ID  string
					Foo string `json:"foo"`
				}{}

				Convey("Should accept empty attributes", func() {
					object := &Object{
						ID:         "ID123",
						Type:       testType,
						Attributes: json.RawMessage{},
					}
					err := object.Unmarshal(testType, &testConversion)
					So(err, ShouldBeNil)
					So(testConversion.Foo, ShouldBeEmpty)
				})

				Convey("Should successfully populate a valid struct", func() {
					err := testObject.Unmarshal(testType, &testConversion)
					So(err, ShouldBeNil)
					So(testConversion.Foo, ShouldEqual, "bar")
				})

				Convey("Should reject a non-matching type", func() {
					err := testObject.Unmarshal("badType", &testConversion)
					So(err, ShouldNotBeNil)
				})

			})

			Convey("govalidator struct unmarshals", func() {

				Convey("should not error if input validates properly", func() {
					testValidation := struct {
						Foo string `json:"foo" valid:"alphanum"`
					}{}

					err := testObject.Unmarshal(testType, &testValidation)
					So(err, ShouldBeNil)
					So(testValidation.Foo, ShouldEqual, "bar")
				})

				Convey("should return a 422 Error correctly for a validation failure", func() {
					testValidation := struct {
						Foo string `valid:"ipv4,required" json:"foo"`
					}{}

					err := testObject.Unmarshal(testType, &testValidation)
					So(err, ShouldNotBeNil)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
				})

				Convey("should return a 422 Error correctly for multiple validation failures", func() {

					testManyObject := &Object{
						ID:         "ID123",
						Type:       testType,
						Attributes: json.RawMessage(`{"foo":"bar", "baz":"4567"}`),
					}

					testManyValidations := struct {
						Foo string `valid:"ipv4,required" json:"foo"`
						Baz string `valid:"alpha,required" json:"baz"`
					}{}

					err := testManyObject.Unmarshal(testType, &testManyValidations)
					So(err, ShouldNotBeNil)

					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
					So(err[1].Source, ShouldNotBeNil)
					So(err[1].Source.Pointer, ShouldEqual, "/data/attributes/baz")
				})
			})
		})

		Convey("->Marshal()", func() {

			Convey("should properly update attributes", func() {
				attrs := map[string]string{"foo": "baz"}
				err := testObject.Marshal(attrs)
				So(err, ShouldBeNil)

				raw, jsonErr := json.MarshalIndent(attrs, "", " ")
				So(jsonErr, ShouldBeNil)
				So(string(testObject.Attributes), ShouldEqual, string(raw))
			})
		})

		Convey("->Validate()", func() {

			Convey("should handle a POST response correctly", func() {
				request.Method = "POST"
				err := testObject.Validate(request, true)
				So(err, ShouldBeNil)
				So(testObject.Status, ShouldEqual, http.StatusCreated)
			})

			Convey("should handle a GET response correctly", func() {
				request.Method = "GET"
				err := testObject.Validate(request, true)
				So(err, ShouldBeNil)
				So(testObject.Status, ShouldEqual, http.StatusOK)
			})

			Convey("should handle a PATCH response correctly", func() {
				request.Method = "PATCH"
				err := testObject.Validate(request, true)
				So(err, ShouldBeNil)
				So(testObject.Status, ShouldEqual, http.StatusOK)
			})

			Convey("should return a formatted Error for an unsupported method Type", func() {
				request.Method = "PUT"
				err := testObject.Validate(request, true)
				So(err, ShouldNotBeNil)
				So(err.Status, ShouldEqual, http.StatusNotAcceptable)
			})
		})

		Convey("->Send(Object)", func() {
			request.Method = "POST"
			writer := httptest.NewRecorder()
			err := Send(writer, request, testObject)
			So(err, ShouldBeNil)
			So(writer.Code, ShouldEqual, http.StatusCreated)
		})
	})
}
