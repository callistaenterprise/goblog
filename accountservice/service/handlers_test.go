package service

import (
        . "github.com/smartystreets/goconvey/convey"
        "testing"
        "net/http/httptest"
        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "github.com/callistaenterprise/goblog/accountservice/model"
        "fmt"
        "encoding/json"
        "github.com/callistaenterprise/goblog/accountservice/messaging"
        "github.com/stretchr/testify/mock"
        "time"
)




func TestGetAccount(t *testing.T) {
        mockRepo := &dbclient.MockBoltClient{}

        mockRepo.On("QueryAccount", "123").Return(model.Account{Id:"123", Name:"Person_123"}, nil)
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
                                So(account.Id, ShouldEqual, "123")
                                So(account.Name, ShouldEqual, "Person_123")
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

func TestNotificationIsSentForVIPAccount(t *testing.T) {

        DBClient := &dbclient.MockBoltClient{}
        DBClient.On("QueryAccount", "10000").Return(model.Account{Id:"10000", Name:"Person_10000"}, nil)

        MessagingClient := &messaging.MockMessagingClient{}
        MessagingClient.On("SendMessage", mock.Anything, mock.Anything, mock.Anything).Return(nil)

        Convey("Given a HTTP req for a VIP account", t, func() {
                req := httptest.NewRequest("GET", "/accounts/10000", nil)
                resp := httptest.NewRecorder()
                Convey("When the request is handled by the Router", func() {
                        NewRouter().ServeHTTP(resp, req)
                        Convey("Then the response should be a 200 and the MessageClient should have been invoked", func() {
                                So(resp.Code, ShouldEqual, 200)
                                time.Sleep(time.Millisecond * 10)    // Sleep since the Assert below occurs in goroutine
                                So(MessagingClient.AssertNumberOfCalls(t, "SendMessage", 1), ShouldBeTrue)
                        })
        })})
}
