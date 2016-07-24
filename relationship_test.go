package jsh

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRelationship(t *testing.T) {
	Convey("Relationship Tests", t, func() {

		testObject := NewIDObject("tests", "1")

		testList := IDList{testObject}
		req := &http.Request{Method: "GET"}

		Convey("->Sort(IDList)", func() {
			secondObject := NewIDObject("tests", "2")
			thirdObject := NewIDObject("tests", "3")

			testList = append(testList, thirdObject)
			testList = append(testList, secondObject)
			sort.Sort(testList)
			So(testList[0].ID, ShouldEqual, testObject.ID)
			So(testList[1].ID, ShouldEqual, secondObject.ID)
			So(testList[2].ID, ShouldEqual, thirdObject.ID)
		})

		Convey("->Send(IDObject)", func() {

			Convey("should send a properly formatted IDObject response", func() {
				req.Method = "PATCH"
				writer := httptest.NewRecorder()
				err := Send(writer, req, testObject)
				So(err, ShouldBeNil)
				So(writer.Code, ShouldEqual, http.StatusOK)

				contentLength, convErr := strconv.Atoi(writer.HeaderMap.Get("Content-Length"))
				So(convErr, ShouldBeNil)
				So(contentLength, ShouldBeGreaterThan, 0)
				So(writer.HeaderMap.Get("Content-Type"), ShouldEqual, ContentType)

				req, reqErr := testRequest(writer.Body.Bytes())
				So(reqErr, ShouldBeNil)

				response, parseErr := ParseRelationship(req)
				So(parseErr, ShouldBeNil)
				So(response.ID, ShouldEqual, testObject.ID)
				So(response.Type, ShouldEqual, testObject.Type)
			})

			Convey("should accept a nil IDObject", func() {
				writer := httptest.NewRecorder()
				err := Send(writer, req, (*IDObject)(nil))
				So(err, ShouldBeNil)
				So(writer.Code, ShouldEqual, http.StatusOK)

				contentLength, convErr := strconv.Atoi(writer.HeaderMap.Get("Content-Length"))
				So(convErr, ShouldBeNil)
				So(contentLength, ShouldBeGreaterThan, 0)
				So(writer.HeaderMap.Get("Content-Type"), ShouldEqual, ContentType)

				req, reqErr := testRequest(writer.Body.Bytes())
				So(reqErr, ShouldBeNil)

				response, parseErr := ParseRelationship(req)
				So(parseErr, ShouldBeNil)
				So(response, ShouldBeNil)
			})
		})

		Convey("->Send(IDList)", func() {

			Convey("should send a properly formatted IDList response", func() {
				writer := httptest.NewRecorder()
				err := Send(writer, req, testList)
				So(err, ShouldBeNil)
				So(writer.Code, ShouldEqual, http.StatusOK)

				contentLength, convErr := strconv.Atoi(writer.HeaderMap.Get("Content-Length"))
				So(convErr, ShouldBeNil)
				So(contentLength, ShouldBeGreaterThan, 0)
				So(writer.HeaderMap.Get("Content-Type"), ShouldEqual, ContentType)

				req, reqErr := testRequest(writer.Body.Bytes())
				So(reqErr, ShouldBeNil)

				responseList, parseErr := ParseRelationshipList(req)
				So(parseErr, ShouldBeNil)
				So(len(responseList), ShouldEqual, 1)
			})

			Convey("should send a properly formatted empty IDList response", func() {
				writer := httptest.NewRecorder()
				err := Send(writer, req, IDList{})
				So(err, ShouldBeNil)
				So(writer.Code, ShouldEqual, http.StatusOK)

				contentLength, convErr := strconv.Atoi(writer.HeaderMap.Get("Content-Length"))
				So(convErr, ShouldBeNil)
				So(contentLength, ShouldBeGreaterThan, 0)
				So(writer.HeaderMap.Get("Content-Type"), ShouldEqual, ContentType)

				req, reqErr := testRequest(writer.Body.Bytes())
				So(reqErr, ShouldBeNil)

				responseList, parseErr := ParseRelationshipList(req)
				So(parseErr, ShouldBeNil)
				So(len(responseList), ShouldEqual, 0)
			})

			Convey("should accept empty IDList", func() {
				writer := httptest.NewRecorder()
				err := Send(writer, req, IDList{})
				So(err, ShouldBeNil)
			})
		})

		Convey("->UnmarshalJSON()", func() {

			Convey("should handle a linkage object", func() {
				jObj := `{"type": "testRelationship", "id": "ID456"}`

				rl := IDList{}
				err := rl.UnmarshalJSON([]byte(jObj))
				So(err, ShouldBeNil)
				So(len(rl), ShouldEqual, 1)
			})

			Convey("should handle a linkage list", func() {
				jList := `[
					{"type": "testRelationship", "id": "ID456"},
					{"type": "testRelationship", "id": "ID789"}
				]`

				rl := IDList{}
				err := rl.UnmarshalJSON([]byte(jList))
				So(err, ShouldBeNil)
				So(len(rl), ShouldEqual, 2)
			})
		})
	})
}
