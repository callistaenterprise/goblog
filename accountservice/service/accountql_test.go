package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"encoding/json"
	"github.com/graphql-go/graphql"
)

// ffd508b5-5f87-4246-9867-ead4ecb01357
// Query that showcases variables, field selection and parameters.
var fetchAccountQuery = `query fetchAccount($accid: String!) {
    Account(id:$accid) {
        id,name,events(eventName:"CREATED") {
            eventName
        },quote(language:"en") {
            quote
        },imageData{id,url}
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
	initQL(&TestGraphQLResolvers{})
	Convey("Given a GraphQL request for account 123", t, func() {
		vars := make(map[string]interface{})
		vars["accid"] = "123"
		params := graphql.Params{Schema: schema, VariableValues: vars, RequestString: fetchAccountQuery}

		Convey("When the request is executed", func() {
			r := graphql.Do(params)
			rJSON, _ := json.Marshal(r)

			Convey("Then the response should be as expected", func() {
				So(len(r.Errors), ShouldEqual, 0)
				So(string(rJSON), ShouldEqual, `{"data":{"Account":{"events":[{"eventName":"CREATED"}],"id":"123","imageData":{"id":"123","url":"http://fake.path/image.png"},"name":"Test Testsson 3","quote":{"quote":"HEJ"}}}}`)
			})
		})
	})
}

func TestFetchAliasedAccounts(t *testing.T) {
	initQL(&TestGraphQLResolvers{})
	Convey("Given a GraphQL request for account 123 and 124", t, func() {
		params := graphql.Params{Schema: schema, RequestString: fetchAliasedAccountsQuery} //{...AccountFragment,quote{quote,language}}}fragment AccountFragment on Account{id,name}"}

		Convey("When the request is executed", func() {
			r := graphql.Do(params)

			Convey("Then the response should contain two entries", func() {
				So(len(r.Errors), ShouldEqual, 0)
				So(len(r.Data.(map[string]interface{})), ShouldEqual, 2)
			})
		})
	})
}
