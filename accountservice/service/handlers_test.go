package service

import (
        "encoding/json"
        "fmt"
        "net/http/httptest"
        "testing"
        "time"

        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "github.com/callistaenterprise/goblog/accountservice/model"
        "github.com/callistaenterprise/goblog/common/messaging"
        . "github.com/smartystreets/goconvey/convey"
        "github.com/stretchr/testify/mock"
        "gopkg.in/h2non/gock.v1"
        "github.com/callistaenterprise/goblog/common/tracing"
        "github.com/opentracing/opentracing-go"
        "context"
)

// mocks for boltdb and messsaging
var mockRepo = &dbclient.MockBoltClient{}
var mockMessagingClient = &messaging.MockMessagingClient{}

// mock types
var anyString = mock.AnythingOfType("string")
var anyByteArray = mock.AnythingOfType("[]uint8")

func init() {
        gock.InterceptClient(client)
        tracing.tracer = opentracing.NoopTracer{}
}

func TestGetAccount(t *testing.T) {
        defer gock.Off()
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(200).
                BodyString(`{"quote":"May the source be with you, always.","ipAddress":"10.0.0.5:8080","language":"en"}`)
        gock.New("http://imageservice:7777").
                Get("/accounts/10000").
                Reply(200).
                BodyString(`{"imageUrl":"http://test.path"}`)

        mockRepo.On("QueryAccount", mock.Anything, "123").Return(model.Account{ID: "123", Name: "Person_123"}, nil)
        mockRepo.On("QueryAccount", mock.Anything, "456").Return(model.Account{}, fmt.Errorf("Some error"))
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
                                So(account.Quote.Text, ShouldEqual, "May the source be with you, always.")
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
        gock.New("http://imageservice:7777").
                Get("/accounts/10000").
                Reply(200).
                BodyString(`{"imageUrl":"http://test.path"}`)

        mockRepo := &dbclient.MockBoltClient{}
        mockRepo.On("QueryAccount", mock.Anything, "123").Return(model.Account{ID: "123", Name: "Person_123"}, nil)
        DBClient = mockRepo

        Convey("Given a HTTP request for /accounts/123", t, func() {
                req := httptest.NewRequest("GET", "/accounts/123", nil).WithContext(context.TODO())
                resp := httptest.NewRecorder()

                Convey("When the request is handled by the Router", func() {
                        NewRouter().ServeHTTP(resp, req)

                        Convey("Then the response should be a 200", func() {
                                So(resp.Code, ShouldEqual, 200)

                                account := model.Account{}
                                json.Unmarshal(resp.Body.Bytes(), &account)
                                So(account.ID, ShouldEqual, "123")
                                So(account.Name, ShouldEqual, "Person_123")
                                So(account.Quote.Text, ShouldEqual, "May the source be with you, always.")
                        })
                })
        })
}

func TestNotificationIsSentForVIPAccount(t *testing.T) {
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(200).
                BodyString(`{"quote":"May the source be with you, always.","ipAddress":"10.0.0.5:8080","language":"en"}`)
        gock.New("http://imageservice:7777").
                Get("/accounts/10000").
                Reply(200).
                BodyString(`{"imageUrl":"http://test.path"}`)

        mockRepo.On("QueryAccount", mock.Anything, "10000").Return(model.Account{ID: "10000", Name: "Person_10000"}, nil)
        DBClient = mockRepo

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
