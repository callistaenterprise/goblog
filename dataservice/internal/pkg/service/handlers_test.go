package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/dataservice/internal/pkg/dbclient"
	"github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

var mockRepo = &dbclient.MockGormClient{}

var serviceName = "dataservice"

// mock types
var anyString = mock.AnythingOfType("string")
var anyByteArray = mock.AnythingOfType("[]uint8")

// Run this first in each test, poor substitute for a proper @Before func
func reset() {
	mockRepo = &dbclient.MockGormClient{}
	tracing.SetTracer(opentracing.NoopTracer{})
}

func TestGetAccount(t *testing.T) {
	reset()

	mockRepo.On("QueryAccount", mock.Anything, "123").Return(model.AccountData{ID: "123", Name: "Person_123"}, nil)
	mockRepo.On("QueryAccount", mock.Anything, "456").Return(model.AccountData{}, fmt.Errorf(""))
	DBClient = mockRepo

	Convey("Given a HTTP request for /accounts/123", t, func() {
		req := httptest.NewRequest("GET", "/accounts/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(serviceName).ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)

				account := model.AccountData{}
				json.Unmarshal(resp.Body.Bytes(), &account)
				So(account.ID, ShouldEqual, "123")
				So(account.Name, ShouldEqual, "Person_123")
			})
		})
	})

	Convey("Given a HTTP request for /accounts/456", t, func() {
		req := httptest.NewRequest("GET", "/accounts/456", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(serviceName).ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
				responseBody, _ := ioutil.ReadAll(resp.Body)
				So(string(responseBody), ShouldEqual, "Account not found")
			})
		})
	})
}

func TestGetAccountWrongPath(t *testing.T) {
	reset()
	Convey("Given a HTTP request for /invalid/123", t, func() {
		req := httptest.NewRequest("GET", "/invalid/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(serviceName).ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}
