package service

import (
	"encoding/json"
	"fmt"
	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/dataservice/cmd"
	"github.com/callistaenterprise/goblog/dataservice/internal/pkg/dbclient/mock_dbclient"
	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

var serviceName = "dataservice"

// Run this first in each test, poor substitute for a proper @Before func
func reset() {
	tracing.SetTracer(opentracing.NoopTracer{})
}

func TestGetAccount(t *testing.T) {
	reset()
	ctrl := gomock.NewController(t)
	mockRepo := mock_dbclient.NewMockIGormClient(ctrl)

	mockRepo.EXPECT().QueryAccount(gomock.Any(), "123").Return(model.AccountData{ID: "123", Name: "Person_123"}, nil)
	mockRepo.EXPECT().QueryAccount(gomock.Any(), "456").Return(model.AccountData{}, fmt.Errorf(""))

	s := NewServer(mockRepo, cmd.DefaultConfiguration())
	s.SetupRoutes()

	Convey("Given a HTTP request for /accounts/123", t, func() {
		req := httptest.NewRequest("GET", "/accounts/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			s.r.ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)

				account := model.AccountData{}
				_ = json.Unmarshal(resp.Body.Bytes(), &account)
				So(account.ID, ShouldEqual, "123")
				So(account.Name, ShouldEqual, "Person_123")
			})
		})
	})

	Convey("Given a HTTP request for /accounts/456", t, func() {
		req := httptest.NewRequest("GET", "/accounts/456", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			s.r.ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
				responseBody, _ := ioutil.ReadAll(resp.Body)
				So(string(responseBody), ShouldContain, "404 page not found")
			})
		})
	})
}

func TestHealth(t *testing.T) {
	tracing.SetTracer(opentracing.NoopTracer{})
	s := NewServer(nil, cmd.DefaultConfiguration())
	s.SetupRoutes()

	req := httptest.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	s.r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

func TestGetAccountWrongPath(t *testing.T) {
	reset()

	s := NewServer(nil, cmd.DefaultConfiguration())
	s.SetupRoutes()
	Convey("Given a HTTP request for /invalid/123", t, func() {
		req := httptest.NewRequest("GET", "/invalid/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			s.r.ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}
