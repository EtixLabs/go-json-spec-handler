package jsc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EtixLabs/go-json-spec-handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPatch(t *testing.T) {

	Convey("Patch Tests", t, func() {
		api := testAPI()
		server := httptest.NewServer(api)
		defer server.Close()
		baseURL := server.URL

		Convey("->Patch()", func() {
			object, err := jsh.NewObject("2", "tests", nil)
			So(err, ShouldBeNil)

			json, resp, patchErr := Patch(baseURL, object)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(patchErr, ShouldBeNil)
			So(json.HasErrors(), ShouldBeFalse)
			So(json.HasData(), ShouldBeTrue)
		})

		Convey("->PatchOne()", func() {

			Convey("should accept an ID object", func() {
				object := jsh.NewIDObject("foos", "1")
				json, resp, patchErr := PatchOne(baseURL, "tests", "1", "foo", object)
				So(patchErr, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
				So(json, ShouldBeNil)
			})

			Convey("should accept a nil object", func() {
				json, resp, patchErr := PatchOne(baseURL, "tests", "1", "foo", nil)
				So(patchErr, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
				So(json, ShouldBeNil)
			})
		})

		Convey("->PatchMany()", func() {

			Convey("should accept an ID list", func() {
				object := jsh.NewIDObject("foos", "1")
				json, resp, patchErr := PatchMany(baseURL, "tests", "1", "foos", jsh.IDList{object})
				So(patchErr, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
				So(json, ShouldBeNil)
			})

			Convey("should accept an empty list", func() {
				json, resp, patchErr := PatchMany(baseURL, "tests", "1", "foo", nil)
				So(patchErr, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
				So(json, ShouldBeNil)
			})
		})
	})
}
