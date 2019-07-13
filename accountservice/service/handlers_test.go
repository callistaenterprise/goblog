package service

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"context"
	internalmodel "github.com/callistaenterprise/goblog/accountservice/model"
	"github.com/callistaenterprise/goblog/common/circuitbreaker"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"strings"
)

var mockMessagingClient *messaging.MockMessagingClient

// mock types
var anyString = mock.AnythingOfType("string")
var anyByteArray = mock.AnythingOfType("[]uint8")

// Run this first in each test, poor substitute for a proper @Before func
func reset() {
	mockMessagingClient = &messaging.MockMessagingClient{}
	gock.InterceptClient(client)
	circuitbreaker.Client = *client
	tracing.SetTracer(opentracing.NoopTracer{})
}

func TestGetAccount(t *testing.T) {
	reset()
	defer gock.Off()
	gock.New("http://dataservice:7070").
		Get("/accounts/123").
		Reply(200).
		BodyString(`{"ID":"123", "Name":"Test Testsson", "ServedBy":"127.0.0.1"}`)
	gock.New("http://quotes-service:8080").
		Get("/api/quote").
		MatchParam("strength", "4").
		Reply(200).
		BodyString(`{"quote":"May the source be with you, always.","ipAddress":"10.0.0.5:8080","language":"en"}`)
	gock.New("http://imageservice:7777").
		Get("/accounts/10000").
		Reply(200).
		BodyString(`{"imageUrl":"http://test.path"}`)

	Convey("Given a HTTP request for /accounts/123", t, func() {
		req := httptest.NewRequest("GET", "/accounts/123", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)

				account := internalmodel.Account{}
				json.Unmarshal(resp.Body.Bytes(), &account)
				So(account.ID, ShouldEqual, "123")
				So(account.Name, ShouldEqual, "Test Testsson")
				So(account.Quote.Text, ShouldEqual, "May the source be with you, always.")
			})
		})
	})

	Convey("Given a HTTP request for /accounts/456", t, func() {
		req := httptest.NewRequest("GET", "/accounts/456", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 500", func() {
				So(resp.Code, ShouldEqual, 500)
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
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func TestGetAccountNoQuote(t *testing.T) {
	reset()
	defer gock.Off()
	gock.New("http://dataservice:7070").
		Get("/accounts/123").
		Reply(200).
		BodyString(`{"ID":"123", "Name":"Test Testsson", "ServedBy":"127.0.0.1"}`)

	gock.New("http://quotes-service:8080").
		Get("/api/quote").
		MatchParam("strength", "4").
		Reply(500).
		BodyString(`{"imageUrl":"http://test.path"}`)

	gock.New("http://imageservice:7777").
		Get("/accounts/10000").
		Reply(200).
		BodyString(`{"imageUrl":"http://test.path"}`)

	Convey("Given a HTTP request for /accounts/123", t, func() {
		req := httptest.NewRequest("GET", "/accounts/123", nil).WithContext(context.TODO())
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)

				account := internalmodel.Account{}
				json.Unmarshal(resp.Body.Bytes(), &account)
				So(account.ID, ShouldEqual, "123")
				So(account.Name, ShouldEqual, "Test Testsson")
				So(account.Quote.Text, ShouldEqual, "May the source be with you, always.")
			})
		})
	})
}

func TestNotificationIsSentForVIPAccount(t *testing.T) {
	reset()
	gock.New("http://dataservice:7070").
		Get("/accounts/10000").
		Reply(200).
		BodyString(`{"ID":"10000", "Name":"Test Testsson", "ServedBy":"127.0.0.1"}`)

	gock.New("http://quotes-service:8080").
		Get("/api/quote").
		MatchParam("strength", "4").
		Reply(200).
		BodyString(`{"quote":"May the source be with you, always.","ipAddress":"10.0.0.5:8080","language":"en"}`)
	gock.New("http://imageservice:7777").
		Get("/accounts/10000").
		Reply(200).
		BodyString(`{"imageUrl":"http://test.path"}`)

	mockMessagingClient.On("PublishOnQueueWithContext", mock.Anything, anyByteArray, anyString).Return(nil)
	MessagingClient = mockMessagingClient

	Convey("Given a HTTP req for a VIP account", t, func() {
		req := httptest.NewRequest("GET", "/accounts/10000", nil)
		resp := httptest.NewRecorder()
		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)
			Convey("Then the response should be a 200 and the MessageClient should have been invoked", func() {
				So(resp.Code, ShouldEqual, 200)
				time.Sleep(time.Millisecond * 10) // Sleep since the Assert below occurs in goroutine
				So(mockMessagingClient.AssertNumberOfCalls(t, "PublishOnQueueWithContext", 1), ShouldBeTrue)
			})
		})
	})
}


func TestHealthCheckOk(t *testing.T) {
	reset()

	Convey("Given a HTTP req for /health", t, func() {

		req := httptest.NewRequest("GET", "/health", nil)
		resp := httptest.NewRecorder()
		Convey("When served", func() {
			NewRouter().ServeHTTP(resp, req)
			Convey("Then expect 200 OK", func() {
				So(resp.Code, ShouldEqual, 200)
			})
		})
	})
}

// Tests the /ql graphQL endpoint
func TestQueryAccountUsingGraphQL(t *testing.T) {
	tracing.SetTracer(opentracing.NoopTracer{})
	initQL(&TestGraphQLResolvers{})

	query := "{Account(id:\"123\"){id,name,quote(language:\"sv\"){quote,language}}}"

	Convey("Given a GraphQL request for {Account{id,name}}", t, func() {
		req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
		req.Header.Add("Content-Type", "application/graphql")
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
				body, _ := ioutil.ReadAll(resp.Body)
				So(string(body), ShouldEqual, `{"data":{"Account":{"id":"123","name":"Test Testsson 3","quote":{"language":"sv","quote":"HEJ"}}}}`)
			})
		})
	})
}

// Tests the /ql graphQL endpoint
func TestQueryAccountSmallUsingGraphQL(t *testing.T) {
	tracing.SetTracer(opentracing.NoopTracer{})
	initQL(&TestGraphQLResolvers{})

	query := "{Account(id:\"123\"){id}}"

	Convey("Given a GraphQL request for {Account{id,name}}", t, func() {
		req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
		req.Header.Add("Content-Type", "application/graphql")
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
				body, _ := ioutil.ReadAll(resp.Body)
				So(string(body), ShouldEqual, `{"data":{"Account":{"id":"123"}}}`)
			})
		})
	})
}
