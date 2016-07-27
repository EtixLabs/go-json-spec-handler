package jsc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {

	api := testAPI()
	server := httptest.NewServer(api)
	defer server.Close()

	baseURL := server.URL

	Convey("Get Tests", t, func() {

		Convey("->List()", func() {

			Convey("should handle an object listing request", func() {
				json, resp, err := List(baseURL, "tests")

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(json.HasErrors(), ShouldBeFalse)
				So(json.HasData(), ShouldBeTrue)
			})
		})

		Convey("->Fetch()", func() {

			Convey("should handle a specific object request", func() {
				json, resp, err := Fetch(baseURL, "tests", "1")

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(json.HasErrors(), ShouldBeFalse)
				So(json.HasData(), ShouldBeTrue)
				So(json.First().ID, ShouldEqual, "1")
			})
		})

		Convey("->FetchRelated()", func() {

			Convey("should handle a related object request", func() {
				json, resp, err := FetchRelated(baseURL, "tests", "1", "foo")

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(json.HasErrors(), ShouldBeFalse)
				So(json.HasData(), ShouldBeTrue)
				So(json.First().ID, ShouldEqual, "1")
				So(json.First().Type, ShouldEqual, "foos")
				So(json.First().Attributes, ShouldNotBeEmpty)
			})

			Convey("should handle a related list request", func() {
				json, resp, err := ListRelated(baseURL, "tests", "1", "foos")

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(json.HasErrors(), ShouldBeFalse)
				So(json.HasData(), ShouldBeTrue)
				So(json.First().ID, ShouldEqual, "1")
				So(json.First().Type, ShouldEqual, "foos")
				So(json.First().Attributes, ShouldNotBeEmpty)
			})
		})

		Convey("->FetchRelationship()", func() {

			Convey("should handle a relationship object request", func() {
				json, resp, err := FetchRelationship(baseURL, "tests", "1", "foo")

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(json.HasErrors(), ShouldBeFalse)
				So(json.HasData(), ShouldBeTrue)
				So(json.First().ID, ShouldEqual, "1")
				So(json.First().Type, ShouldEqual, "foos")
				So(json.First().Attributes, ShouldBeEmpty)
			})

			Convey("should handle a relationship list request", func() {
				json, resp, err := ListRelationship(baseURL, "tests", "1", "foos")

				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(json.HasErrors(), ShouldBeFalse)
				So(json.HasData(), ShouldBeTrue)
				So(json.First().ID, ShouldEqual, "1")
				So(json.First().Type, ShouldEqual, "foos")
				So(json.First().Attributes, ShouldBeEmpty)
			})
		})
	})
}
