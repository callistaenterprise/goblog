package service

import (
        . "github.com/smartystreets/goconvey/convey"
        "testing"
        "net/http/httptest"
        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "github.com/callistaenterprise/goblog/accountservice/model"
        "fmt"
        "encoding/json"
        "github.com/callistaenterprise/goblog/common/messaging"
        "gopkg.in/h2non/gock.v1"
        "github.com/stretchr/testify/mock"
        "time"
        "github.com/callistaenterprise/goblog/common/tracing"
        zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

var mockRepo = &dbclient.MockGormClient{}
var mockMessagingClient = &messaging.MockMessagingClient{}

// mock types
var anyString = mock.AnythingOfType("string")
var anyByteArray = mock.AnythingOfType("[]uint8")

func init() {
        gock.InterceptClient(client)
        tracing.Tracer, _ = zipkin.NewTracer(
                zipkin.NewRecorder(zipkin.NopCollector{}, false, "127.0.0.1:0", "accountservice"))
}

func TestGetAccount(t *testing.T) {
        defer gock.Off()
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(200).
                BodyString(`{"quote":"May the source be with you. Always.","ipAddress":"10.0.0.5:8080","language":"en"}`)
        
        mockRepo.On("QueryAccount", "123").Return(model.Account{ID:"123", Name:"Person_123"}, nil)
        mockRepo.On("QueryAccount", "456").Return(model.Account{}, fmt.Errorf("Some error"))
        DBClient = mockRepo

        Convey("Given a HTTP request for /accounts/123", t, func() {
                req := httptest.NewRequest("GET", "/accounts/123", nil)
                resp := httptest.NewRecorder()

                Convey("When the request is handled by the Router", func() {
                        NewRouter().ServeHTTP(resp, req)

                        Convey("Then the response should be a 200", func() {
                                So(resp.Code, ShouldEqual, 200)

                                account := model.Account{}
                                json.Unmarshal(resp.Body.Bytes(), &account)
                                So(account.ID, ShouldEqual, "123")
                                So(account.Name, ShouldEqual, "Person_123")
                                So(account.Quote.Text, ShouldEqual, "May the source be with you. Always.")
                        })
                })
        })

        Convey("Given a HTTP request for /accounts/456", t, func() {
                req := httptest.NewRequest("GET", "/accounts/456", nil)
                resp := httptest.NewRecorder()

                Convey("When the request is handled by the Router", func() {
                        NewRouter().ServeHTTP(resp, req)

                        Convey("Then the response should be a 404", func() {
                                So(resp.Code, ShouldEqual, 404)
                        })
                })
        })
}

func TestGetAccountWrongPath(t *testing.T) {

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
        defer gock.Off()
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(500)

        mockRepo := &dbclient.MockGormClient{}
        mockRepo.On("QueryAccount", "123").Return(model.Account{ID:"123", Name:"Person_123"}, nil)

        Convey("Given a HTTP request for /accounts/123", t, func() {
                req := httptest.NewRequest("GET", "/accounts/123", nil)
                resp := httptest.NewRecorder()

                Convey("When the request is handled by the Router", func() {
                        NewRouter().ServeHTTP(resp, req)

                        Convey("Then the response should be a 200", func() {
                                So(resp.Code, ShouldEqual, 200)

                                account := model.Account{}
                                json.Unmarshal(resp.Body.Bytes(), &account)
                                So(account.ID, ShouldEqual, "123")
                                So(account.Name, ShouldEqual, "Person_123")
                                So(account.Quote, ShouldBeZeroValue)
                        })
                })
        })
}

func TestNotificationIsSentForVIPAccount(t *testing.T) {

        defer gock.Off()
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(200).
                BodyString(`{"quote":"May the source be with you. Always.","ipAddress":"10.0.0.5:8080","language":"en"}`)

        mockRepo.On("QueryAccount", "10000").Return(model.Account{ID:"10000", Name:"Person_10000"}, nil)
        DBClient = mockRepo

        mockMessagingClient.On("PublishOnQueue", anyByteArray, anyString).Return(nil)
        MessagingClient = mockMessagingClient

        Convey("Given a HTTP req for a VIP account", t, func() {
                req := httptest.NewRequest("GET", "/accounts/10000", nil)
                resp := httptest.NewRecorder()
                Convey("When the request is handled by the Router", func() {
                        NewRouter().ServeHTTP(resp, req)
                        Convey("Then the response should be a 200 and the MessageClient should have been invoked", func() {
                                So(resp.Code, ShouldEqual, 200)
                                time.Sleep(time.Millisecond * 10)    // Sleep since the Assert below occurs in goroutine
                                So(mockMessagingClient.AssertNumberOfCalls(t, "PublishOnQueue", 1), ShouldBeTrue)
                        })
        })})
}
