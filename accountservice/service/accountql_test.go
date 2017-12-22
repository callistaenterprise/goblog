package service

import (
    "net/http/httptest"
    . "github.com/smartystreets/goconvey/convey"
    "testing"
    "github.com/callistaenterprise/goblog/common/tracing"
    "github.com/opentracing/opentracing-go"
    "io/ioutil"
    "encoding/json"
    "github.com/graphql-go/graphql"
)

// Query that showcases variables, field selection and parameters.
var fetchAccountQuery = `query fetchAccount($accid: String!) {
    Account(id:$accid) {
        id,name,events(eventName:"CREATED") {
            eventName
        },quote(language:"en") {
            quote
        }
    }
}`

// Query that showcases variables, field selection and parameters.
var fetchAliasedAccountsQuery = `query fetchAccounts{
    acc1: Account(id:"123") {
        id,name,events(eventName:"CREATED") {
            eventName
        },quote(language:"en") {
            quote
        }
    }
    acc2: Account(id:"124") {
        id,name,events(eventName:"CREATED") {
            eventName
        },quote(language:"en") {
            quote
        }
    }
}`

var fetchAllAccountsQuery = `query fetchAllAccounts {
    AllAccounts{
        id,name,events(eventName:"CREATED"){
            eventName,created
        },quote(language:"en"){
            quote
        }
    }
}`

func TestFetchAccount(t *testing.T) {
    initQL()
    Convey("Given a GraphQL request for account 123", t, func() {
        vars := make(map[string]interface{})
        vars["accid"] = "123"
        params := graphql.Params{Schema: Schema, VariableValues: vars, RequestString: fetchAccountQuery} //{...AccountFragment,quote{quote,language}}}fragment AccountFragment on Account{id,name}"}

        Convey("When the request is executed", func() {
            r := graphql.Do(params)
            rJSON, _ := json.Marshal(r)

            Convey("Then the response should be as expected", func() {
                So(len(r.Errors), ShouldEqual, 0)
                So(string(rJSON), ShouldEqual, `{"data":{"Account":{"events":[{"eventName":"CREATED"}],"id":"123","name":"Test Testsson 3","quote":null}}}`)
            })
        })
    })
}

func TestFetchAliasedAccounts(t *testing.T) {
    initQL()
    Convey("Given a GraphQL request for account 123 and 124", t, func() {
        params := graphql.Params{Schema: Schema, RequestString: fetchAliasedAccountsQuery} //{...AccountFragment,quote{quote,language}}}fragment AccountFragment on Account{id,name}"}

        Convey("When the request is executed", func() {
            r := graphql.Do(params)

            Convey("Then the response should contain two entries", func() {
                So(len(r.Errors), ShouldEqual, 0)
                So(len(r.Data.(map[string]interface{})), ShouldEqual, 2)
            })
        })
    })
}

func TestFetchAllAccounts(t *testing.T) {
    initQL()
    Convey("Given a GraphQL request for account 123", t, func() {
        params := graphql.Params{Schema: Schema, RequestString: fetchAllAccountsQuery} //{...AccountFragment,quote{quote,language}}}fragment AccountFragment on Account{id,name}"}

        Convey("When the request is executed", func() {
            r := graphql.Do(params)

            Convey("Then the response should contain all 10 items.", func() {
                So(len(r.Errors), ShouldEqual, 0)
                So(len(r.Data.(map[string]interface{})["AllAccounts"].([]interface{})), ShouldEqual, 10)
            })
        })
    })
}

func TestQueryAccount(t *testing.T) {
    tracing.SetTracer(opentracing.NoopTracer{})
    initQL()
    Convey("Given a GraphQL request for {Account{id,name}}", t, func() {
        req := httptest.NewRequest("GET", "/ql?query={Account(id:\"123\"){id,name,quote(language:\"sv\"){quote,language}}}", nil)
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

func TestQueryAccountSmall(t *testing.T) {
    tracing.SetTracer(opentracing.NoopTracer{})
    initQL()

    Convey("Given a GraphQL request for {Account{id,name}}", t, func() {
        req := httptest.NewRequest("GET", "/ql?query={Account(id:\"123\"){id}}", nil)
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
