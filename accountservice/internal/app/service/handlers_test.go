package service

import (
	"context"
	"encoding/json"
	"github.com/callistaenterprise/goblog/accountservice/cmd"
	internalmodel "github.com/callistaenterprise/goblog/accountservice/internal/app/model"
	"github.com/callistaenterprise/goblog/common/circuitbreaker"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

//var mockMessagingClient *messaging.MockMessagingClient

var serviceName = "accountservice"

// mock types
var anyString = mock.AnythingOfType("string")
var anyByteArray = mock.AnythingOfType("[]uint8")

// Run this first in each test, poor substitute for a proper @Before func
func reset() {

}

func TestGetAccount(t *testing.T) {

	srv := setup()

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
		Get("/accounts/123").
		Reply(200).
		BodyString(`{"imageUrl":"http://test.path"}`)

	req := httptest.NewRequest("GET", "/accounts/123", nil)
	resp := httptest.NewRecorder()

	srv.r.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)

	account := internalmodel.Account{}
	_ = json.Unmarshal(resp.Body.Bytes(), &account)

	assert.Equal(t, "123", account.ID)
	assert.Equal(t, "Test Testsson", account.Name)
	assert.Equal(t, "May the source be with you, always.", account.Quote.Text)
}

func TestGetAccountNotFound(t *testing.T) {
	srv := setup()

	gock.New("http://dataservice:7070").
		Get("/accounts/456").Times(5).
		Reply(404)

	req := httptest.NewRequest("GET", "/accounts/456", nil)
	resp := httptest.NewRecorder()

	srv.r.ServeHTTP(resp, req)

	// Note that even if the dataservice returns HTTP 404, the accountservice will expose a 500 externally.
	assert.Equal(t, 500, resp.Code)
}

func TestGetAccountWrongPath(t *testing.T) {
	srv := setup()
	req := httptest.NewRequest("GET", "/invalid/123", nil)
	resp := httptest.NewRecorder()

	srv.r.ServeHTTP(resp, req)

	assert.Equal(t, 404, resp.Code)
}

func TestGetAccountNoQuote(t *testing.T) {
	srv := setup()
	defer gock.Off()
	gock.New("http://dataservice:7070").
		Get("/accounts/123").
		Reply(200).
		BodyString(`{"ID":"123", "Name":"Test Testsson", "ServedBy":"127.0.0.1"}`)

	gock.New("http://quotes-service:8080").
		Get("/api/quote").
		MatchParam("strength", "4").Times(4).
		Reply(500).
		BodyString(`{"imageUrl":"http://test.path"}`)

	gock.New("http://imageservice:7777").
		Get("/accounts/123").
		Reply(200).
		BodyString(`{"imageUrl":"http://test.path"}`)

	req := httptest.NewRequest("GET", "/accounts/123", nil).WithContext(context.TODO())
	resp := httptest.NewRecorder()

	srv.r.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)

	account := internalmodel.Account{}
	_ = json.Unmarshal(resp.Body.Bytes(), &account)
	assert.Equal(t, "123", account.ID)
	assert.Equal(t, "Test Testsson", account.Name)
	assert.Equal(t, "May the source be with you, always.", account.Quote.Text)

}

func TestNotificationIsSentForVIPAccount(t *testing.T) {
	srv := setup()
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

	//srv.h.messagingClient = &messaging.MockMessagingClient{}
	srv.h.messagingClient.(*messaging.MockMessagingClient).On("PublishOnQueueWithContext", mock.Anything, anyByteArray, anyString).Return(nil)

	req := httptest.NewRequest("GET", "/accounts/10000", nil)
	resp := httptest.NewRecorder()
	srv.r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	time.Sleep(time.Millisecond * 10) // Sleep since the Assert below occurs in goroutine
	assert.True(t, srv.h.messagingClient.(*messaging.MockMessagingClient).AssertNumberOfCalls(t, "PublishOnQueueWithContext", 1))
}

func TestHealthCheckOk(t *testing.T) {
	srv := setup()
	srv.h.isHealthy = true

	req := httptest.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	srv.r.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

// Tests the /ql graphQL endpoint
func TestQueryAccountUsingGraphQL(t *testing.T) {
	srv := setup()

	query := "{Account(id:\"123\"){id,name,quote(language:\"sv\"){quote,language}}}"

	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
	req.Header.Add("Content-Type", "application/graphql")
	resp := httptest.NewRecorder()

	srv.r.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"data":{"Account":{"id":"123","name":"Test Testsson 3","quote":{"language":"sv","quote":"HEJ"}}}}`, string(body))
}

func setup() *Server {
	client := &http.Client{}
	gock.InterceptClient(client)
	circuitbreaker.Client = client
	tracing.SetTracer(opentracing.NoopTracer{})
	s := NewServer(cmd.DefaultConfiguration(), NewHandler(&messaging.MockMessagingClient{}, client), &TestGraphQLResolvers{})
	s.SetupRoutes()
	return s
}

// Tests the /ql graphQL endpoint
func TestQueryAccountSmallUsingGraphQL(t *testing.T) {
	srv := setup()
	query := "{Account(id:\"123\"){id}}"

	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
	req.Header.Add("Content-Type", "application/graphql")
	resp := httptest.NewRecorder()

	srv.r.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"data":{"Account":{"id":"123"}}}`, string(body))
}
