package jsh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestObject(t *testing.T) {

	Convey("Object Tests", t, func() {

		testObject := &Object{
			ID:         "ID123",
			Type:       "testObject",
			Attributes: json.RawMessage(`{"foo":"bar"}`),
		}

		request := &http.Request{}

		Convey("->NewObject()", func() {

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
				f, err := testObject.ProcessCreate("testObject", nil)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should reject non-pointer type", func() {
				f, err := testObject.ProcessCreate("testObject", "invalid type")
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should reject pointers to non-struct types", func() {
				testConversion := "non-struct type"

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should reject attributes with no jsh create tag", func() {
				testConversion := struct {
					Foo string `json:"foo"`
				}{}

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(err[0].StatusCode(), ShouldEqual, http.StatusForbidden)
				So(err[0].Source, ShouldNotBeNil)
				So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
				So(f, ShouldBeNil)
				// XXX: Foo == bar
			})

			Convey("Should reject attributes with invalid jsh tag", func() {
				testConversion := struct {
					Foo string `json:"foo" jsh:"invalid"`
				}{}

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(err[0].StatusCode(), ShouldEqual, http.StatusForbidden)
				So(err[0].Source, ShouldNotBeNil)
				So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
				So(f, ShouldBeNil)
				// XXX: Foo == bar
			})

			Convey("Should ignore attributes with json '-' tag", func() {
				testConversion := struct {
					Foo string `json:"-"`
				}{
					Foo: "shouldNotBeOverriden",
				}

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldBeNil)
				So(f, ShouldBeEmpty)
				So(testConversion.Foo, ShouldEqual, "shouldNotBeOverriden")
			})

			Convey("Should accept attributes with jsh 'create' tag", func() {
				testConversion := struct {
					Foo string `json:"foo" jsh:"create"`
				}{
					Foo: "shouldBeOverriden",
				}

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldBeNil)
				So(f, ShouldContain, "Foo")
				So(testConversion.Foo, ShouldEqual, "bar")
			})

			Convey("Should ignore private attributes", func() {
				testConversion := struct {
					foo string `json:"foo" jsh:"create"`
				}{
					foo: "shouldNotBeOverriden",
				}

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldBeNil)
				So(f, ShouldBeEmpty)
				So(testConversion.foo, ShouldEqual, "shouldNotBeOverriden")
			})

			Convey("Should reject attributes with missing required field", func() {
				testConversion := struct {
					Foo string `json:"foo" jsh:"create/required"`
					Bar string `json:"bar" jsh:"create/required"`
				}{}

				f, err := testObject.ProcessCreate("testObject", &testConversion)
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(err[0].StatusCode(), ShouldEqual, 422)
				So(err[0].Source, ShouldNotBeNil)
				So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/bar")
				So(f, ShouldBeNil)
			})

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
					f, err := testObject.ProcessCreate("testObject", &testConversion)
					So(err, ShouldBeNil)
					So(testConversion.Foo, ShouldEqual, "bar")
					So(f, ShouldContain, "Foo")
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

						_, errlist = newObj.ProcessCreate("testObject", &testConversion)
						attr = "ptr"
					})

					Convey("Should detect non-zero value for arrays", func() {
						array := [2]int{0, 1}
						attrs := struct {
							Array [2]int `json:"array"`
						}{array}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate("testObject", &testConversion)
						attr = "array"
					})

					Convey("Should detect non-zero value for slices", func() {
						slice := []int{0, 1}
						attrs := struct {
							Slice []int `json:"slice"`
						}{slice}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate("testObject", &testConversion)
						attr = "slice"
					})

					Convey("Should detect non-zero value for maps", func() {
						m := map[string]int{"a": 0, "b": 1}
						attrs := struct {
							Map map[string]int `json:"map"`
						}{m}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate("testObject", &testConversion)
						attr = "map"
					})

					Convey("Should detect non-zero value for interfaces", func() {
						i := struct{}{}
						attrs := struct {
							Interface interface{} `json:"interface"`
						}{i}
						newObj, err := NewObject(testObject.ID, testObject.Type, attrs)
						So(err, ShouldBeNil)

						_, errlist = newObj.ProcessCreate("testObject", &testConversion)
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

		// ProcessUpdate has the same requirements than ProcessCreate plus a few specific ones
		Convey("->ProcessUpdate()", func() {

			Convey("Should reject different types", func() {
				f, err := testObject.ProcessUpdate("testObject", &struct{}{}, &struct{ int }{})
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(f, ShouldBeNil)
			})

			Convey("Should accept nil model and simply validate fields", func() {
				f, err := testObject.ProcessUpdate("testObject", &struct{}{}, nil)
				So(err, ShouldBeNil)
				So(f, ShouldBeEmpty)
			})

			Convey("Should reject attributes with missing required field", func() {
				testConversion := [2]struct {
					Foo string `json:"foo" jsh:"update/required"`
					Bar string `json:"bar" jsh:"update/required"`
				}{}

				f, err := testObject.ProcessUpdate("testObject", &testConversion[0], &testConversion[1])
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveLength, 1)
				So(err[0].StatusCode(), ShouldEqual, 422)
				So(err[0].Source, ShouldNotBeNil)
				So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/bar")
				So(f, ShouldBeNil)
			})

			Convey("Should accept attribute with jsh 'update' tag", func() {
				testConversion := [2]struct {
					Foo string `json:"foo" jsh:"update"`
					Bar string `json:"bar" jsh:"update"`
				}{}
				f, err := testObject.ProcessUpdate("testObject", &testConversion[0], &testConversion[1])
				So(err, ShouldBeNil)
				So(f, ShouldHaveLength, 1)
				So(f, ShouldContain, "Foo")
				So(testConversion[1].Foo, ShouldEqual, "bar")
				So(testConversion[1].Bar, ShouldBeEmpty)
			})

			Convey("Should accept multiple attributes with jsh 'update' tag", func() {
				testConversion := [3]struct {
					Foo string  `json:"foo" jsh:"update"`
					Bar *string `json:"bar" jsh:"update"`
				}{}
				str := "foo"
				testConversion[0].Bar = &str
				testConversion[0].Foo = "bar"
				newObj, err := NewObject(testObject.ID, testObject.Type, &testConversion[0])
				So(err, ShouldBeNil)

				f, errlist := newObj.ProcessUpdate("testObject", &testConversion[1], &testConversion[2])
				So(errlist, ShouldBeNil)
				So(f, ShouldHaveLength, 2)
				So(f, ShouldContain, "Foo")
				So(f, ShouldContain, "Bar")
				So(testConversion[2].Foo, ShouldEqual, "bar")
				So(*testConversion[2].Bar, ShouldEqual, "foo")
			})
		})

		Convey("->Unmarshal()", func() {

			Convey("non-govalidator structs", func() {

				testConversion := struct {
					ID  string
					Foo string `json:"foo"`
				}{}

				Convey("Should successfully populate a valid struct", func() {
					err := testObject.Unmarshal("testObject", &testConversion)
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

					err := testObject.Unmarshal("testObject", &testValidation)
					So(err, ShouldBeNil)
					So(testValidation.Foo, ShouldEqual, "bar")
				})

				Convey("should return a 422 Error correctly for a validation failure", func() {
					testValidation := struct {
						Foo string `valid:"ipv4,required" json:"foo"`
					}{}

					err := testObject.Unmarshal("testObject", &testValidation)
					So(err, ShouldNotBeNil)
					So(err[0].Source, ShouldNotBeNil)
					So(err[0].Source.Pointer, ShouldEqual, "/data/attributes/foo")
				})

				Convey("should return a 422 Error correctly for multiple validation failures", func() {

					testManyObject := &Object{
						ID:         "ID123",
						Type:       "testObject",
						Attributes: json.RawMessage(`{"foo":"bar", "baz":"4567"}`),
					}

					testManyValidations := struct {
						Foo string `valid:"ipv4,required" json:"foo"`
						Baz string `valid:"alpha,required" json:"baz"`
					}{}

					err := testManyObject.Unmarshal("testObject", &testManyValidations)
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

		Convey("->JSON()", func() {

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
