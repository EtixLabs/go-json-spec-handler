package jsc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EtixLabs/go-json-spec-handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPost(t *testing.T) {

	Convey("Post Tests", t, func() {
		api := testAPI()
		server := httptest.NewServer(api)
		defer server.Close()
		baseURL := server.URL

		attrs := map[string]string{
			"foo": "bar",
		}

		Convey("->Post()", func() {
			object, err := jsh.NewObject("", "tests", attrs)
			So(err, ShouldBeNil)

			json, resp, postErr := Post(baseURL, object)
			So(postErr, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(json.HasErrors(), ShouldBeFalse)
			So(json.HasData(), ShouldBeTrue)
		})

		Convey("->PostMany()", func() {
			object := jsh.NewIDObject("foos", "1")
			json, resp, postErr := PostMany(baseURL, "tests", "1", "foos", jsh.IDList{object})
			So(postErr, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			So(json, ShouldBeNil)
		})
	})
}
