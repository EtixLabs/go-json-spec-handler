package jsh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestList(t *testing.T) {

	Convey("List Tests", t, func() {

		testObject := &Object{
			ID:         "ID123",
			Type:       "testConversion",
			Attributes: json.RawMessage(`{"foo":"bar"}`),
		}

		testList := List{testObject}
		req := &http.Request{Method: "GET"}

		Convey("->Validate()", func() {
			err := testList.Validate(req, true)
			So(err, ShouldBeNil)
		})

		Convey("->Sort(List)", func() {
			firstObject := &Object{
				ID:         "ID1",
				Type:       testObject.Type,
				Attributes: testObject.Attributes,
			}
			thirdObject := &Object{
				ID:         "ID456",
				Type:       testObject.Type,
				Attributes: testObject.Attributes,
			}

			testList = append(testList, thirdObject)
			testList = append(testList, firstObject)
			sort.Sort(testList)
			So(testList[0].ID, ShouldEqual, firstObject.ID)
			So(testList[1].ID, ShouldEqual, testObject.ID)
			So(testList[2].ID, ShouldEqual, thirdObject.ID)
		})

		Convey("->Send(List)", func() {

			Convey("should send a properly formatted List response", func() {
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

				responseList, parseErr := ParseList(req)
				So(parseErr, ShouldBeNil)
				So(len(responseList), ShouldEqual, 1)
			})

			Convey("should send a properly formatted empty List response", func() {
				writer := httptest.NewRecorder()
				err := Send(writer, req, List{})
				So(err, ShouldBeNil)
				So(writer.Code, ShouldEqual, http.StatusOK)

				contentLength, convErr := strconv.Atoi(writer.HeaderMap.Get("Content-Length"))
				So(convErr, ShouldBeNil)
				So(contentLength, ShouldBeGreaterThan, 0)
				So(writer.HeaderMap.Get("Content-Type"), ShouldEqual, ContentType)

				req, reqErr := testRequest(writer.Body.Bytes())
				So(reqErr, ShouldBeNil)

				responseList, parseErr := ParseList(req)
				So(parseErr, ShouldBeNil)
				So(len(responseList), ShouldEqual, 0)
			})

			Convey("should reject nil List", func() {
				writer := httptest.NewRecorder()
				var list List
				err := Send(writer, req, list)
				So(err, ShouldNotBeNil)
				So(writer.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("should accept empty List", func() {
				writer := httptest.NewRecorder()
				err := Send(writer, req, List{})
				So(err, ShouldBeNil)
			})
		})

		Convey("->UnmarshalJSON()", func() {

			Convey("should handle a data object", func() {
				jObj := `{"data": {"type": "user", "id": "sweetID123", "attributes": {"ID":"123"}}}`

				l := List{}
				err := l.UnmarshalJSON([]byte(jObj))
				So(err, ShouldBeNil)
				So(l, ShouldNotBeEmpty)
			})

			Convey("should handle a data list", func() {
				jList := `{"data": [{"type": "user", "id": "sweetID123", "attributes": {"ID":"123"}}]}`

				l := List{}
				err := l.UnmarshalJSON([]byte(jList))
				So(err, ShouldBeNil)
				So(l, ShouldNotBeEmpty)
			})

			Convey("should handle an empty array", func() {
				jObj := `{"data": []}`

				l := List{}
				err := l.UnmarshalJSON([]byte(jObj))
				So(err, ShouldBeNil)
				So(l, ShouldNotBeNil)
			})
		})

		Convey("->MarshalJSON()", func() {

			Convey("should preserve an empty list", func() {
				list := List{}

				jData, err := json.Marshal(list)
				So(err, ShouldBeNil)

				So(string(jData), ShouldEqual, "[]")
			})
		})
	})
}
