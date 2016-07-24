package jsc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EtixLabs/go-json-spec-handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDelete(t *testing.T) {

	Convey("Delete Tests", t, func() {

		api := testAPI()
		server := httptest.NewServer(api)
		defer server.Close()

		baseURL := server.URL

		Convey("->Delete()", func() {
			resp, err := Delete(baseURL, "tests", "1")

			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("->DeleteMany()", func() {
			object := jsh.NewIDObject("foos", "1")
			json, resp, postErr := DeleteMany(baseURL, "tests", "1", "foos", jsh.IDList{object})
			So(postErr, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			So(json, ShouldBeNil)
		})
	})
}
