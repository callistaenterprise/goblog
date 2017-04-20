package service

import (
        . "github.com/smartystreets/goconvey/convey"
        "testing"
        "net/http/httptest"
        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "github.com/callistaenterprise/goblog/accountservice/model"
        "fmt"
        "encoding/json"
        "gopkg.in/h2non/gock.v1"
)


func init() {
        gock.InterceptClient(client)
}

func TestGetAccount(t *testing.T) {
        defer gock.Off()
        gock.New("http://quotes-service:8080").
                Get("/api/quote").
                MatchParam("strength", "4").
                Reply(200).
                BodyString(`{"quote":"May the source be with you. Always.","ipAddress":"10.0.0.5:8080","language":"en"}`)

        
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

        mockRepo := &dbclient.MockBoltClient{}
        mockRepo.On("QueryAccount", "123").Return(model.Account{Id:"123", Name:"Person_123"}, nil)

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
                                So(account.Quote, ShouldBeZeroValue)
                        })
                })
        })
}