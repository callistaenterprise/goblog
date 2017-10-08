package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/callistaenterprise/goblog/accountservice/dbclient"
	"github.com/callistaenterprise/goblog/accountservice/model"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/gorilla/mux"
	cb "github.com/callistaenterprise/goblog/common/circuitbreaker"
	"context"
	"github.com/callistaenterprise/goblog/common/tracing"
)

var DBClient dbclient.IBoltClient
var MessagingClient messaging.IMessagingClient
var isHealthy = true

var client = &http.Client{}

var fallbackQuote = model.Quote{
	Language:"en",
	ServedBy: "circuit-breaker",
	Text: "May the source be with you, always."}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
	cb.Client = *client
}

func GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountId' path parameter from the mux map
	var accountId = mux.Vars(r)["accountId"]

	// Read the account struct BoltDB
	account, err := DBClient.QueryAccount(r.Context(), accountId)
	account.ServedBy = util.GetIP()

	// If err, return a 404
	if err != nil {
		logrus.Errorf("Some error occured serving " + accountId + ": " + err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	notifyVIP(r.Context(), account) // Send VIP notification concurrently.

        account.Quote = getQuote(r.Context())
	account.ImageUrl = getImageUrl(r.Context(), accountId)

	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	writeJsonResponse(w, http.StatusOK, data)
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(ctx context.Context, account model.Account) {
	if account.Id == "10000" {
		go func(account model.Account) {
			vipNotification := model.VipNotification{AccountId: account.Id, ReadAt: time.Now().UTC().String()}
			data, _ := json.Marshal(vipNotification)
			logrus.Infof("Notifying VIP account %v\n", account.Id)
			err := MessagingClient.PublishOnQueueWithContext(ctx, data, "vip_queue")
			if err != nil {
				logrus.Errorln(err.Error())
			}
                        tracing.LogEventToOngoingSpan(ctx, "Sent VIP message")
		}(account)

	}
}

func getQuote(ctx context.Context) (model.Quote) {
        // Start a new opentracing child span
        child := tracing.StartSpanFromContextWithLogEvent(ctx, "getQuote", "Client send")
        defer tracing.CloseSpan(child, "Client Receive")

        // Create the http request and pass it to the circuit breaker
        req, err := http.NewRequest("GET", "http://quotes-service:8080/api/quote?strength=4", nil)
	body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "quotes-service", req)
        if err == nil {
        	quote := model.Quote{}
		json.Unmarshal(body, &quote)
		return quote
	} else {
		return fallbackQuote
	}
}

func getImageUrl(ctx context.Context, accountId string) (string) {
        child := tracing.StartSpanFromContextWithLogEvent(ctx, "getImageUrl", "Client send")
        defer tracing.CloseSpan(child, "Client Receive")

        req, err := http.NewRequest("GET", "http://imageservice:7777/accounts/" + accountId, nil)
        body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "imageservice", req)
        if err == nil {
		return string(body)
	} else {
		return "http://path.to.placeholder"
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
	dbUp := DBClient.Check()
	if dbUp && isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJsonResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
		writeJsonResponse(w, http.StatusServiceUnavailable, data)
	}
}

func SetHealthyState(w http.ResponseWriter, r *http.Request) {
	// Read the 'accountId' path parameter from the mux map
	var state, err = strconv.ParseBool(mux.Vars(r)["state"])
	if err != nil {
		logrus.Errorln("Invalid request to SetHealthyState, allowed values are true or false")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isHealthy = state
	w.WriteHeader(http.StatusOK)
}

func writeJsonResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

type healthCheckResponse struct {
	Status string `json:"status"`
}
