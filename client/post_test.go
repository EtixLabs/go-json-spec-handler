package jsc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EtixLabs/go-json-spec-handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPost(t *testing.T) {

	attrs := map[string]string{
		"foo": "bar",
	}

	api := testAPI()
	server := httptest.NewServer(api)
	defer server.Close()

	baseURL := server.URL

	Convey("Post Tests", t, func() {
		testObject, err := jsh.NewObject("", "tests", attrs)
		So(err, ShouldBeNil)

		_, resp, postErr := Post(baseURL, testObject)
		So(postErr, ShouldBeNil)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
	})
}
