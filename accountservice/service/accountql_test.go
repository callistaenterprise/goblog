package service

import (
	"encoding/json"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"testing"
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
	vars := make(map[string]interface{})
	vars["accid"] = "123"
	params := graphql.Params{Schema: schema, VariableValues: vars, RequestString: fetchAccountQuery}

	r := graphql.Do(params)
	rJSON, _ := json.Marshal(r)

	assert.Equal(t, 0, len(r.Errors))
	assert.Equal(t, string(rJSON), `{"data":{"Account":{"events":[{"eventName":"CREATED"}],"id":"123","imageData":{"id":"123","url":"http://fake.path/image.png"},"name":"Test Testsson 3","quote":{"quote":"HEJ"}}}}`)
}

func TestFetchAliasedAccounts(t *testing.T) {
	initQL(&TestGraphQLResolvers{})
	params := graphql.Params{Schema: schema, RequestString: fetchAliasedAccountsQuery} //{...AccountFragment,quote{quote,language}}}fragment AccountFragment on Account{id,name}"}

	r := graphql.Do(params)

	assert.Equal(t, 0, len(r.Errors))
	assert.Equal(t, 2, len(r.Data.(map[string]interface{})))

}
