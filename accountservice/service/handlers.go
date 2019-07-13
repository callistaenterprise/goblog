package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"context"

	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	internalmodel "github.com/callistaenterprise/goblog/accountservice/model"
	cb "github.com/callistaenterprise/goblog/common/circuitbreaker"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/model"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/gorilla/mux"
	"io/ioutil"
)

// MessagingClient instance
var MessagingClient messaging.IMessagingClient

var myIP string
var isHealthy = true
var client = &http.Client{}

var fallbackQuote = internalmodel.Quote{
	Language: "en",
	ServedBy: "circuit-breaker",
	Text:     "May the source be with you, always."}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
	cb.Client = *client
	var err error
	myIP, err = util.ResolveIPFromHostsFile()
	if err != nil {
		myIP = util.GetIP()
	}
	fmt.Println("Init method executed")
}

func StoreAccount(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	account := &internalmodel.Account{}
	err = json.Unmarshal(data, &account)
	if err != nil {
		errorMsg := fmt.Sprintf("Error parsing JSON: %v", string(data))
		logrus.Errorln(errorMsg)
		writeJSONResponse(w, http.StatusBadRequest, []byte(errorMsg))
	}
	accountData := model.AccountData{Name: account.Name}
	postBody, err := json.Marshal(&accountData)
	storeReq, err := http.NewRequest("POST", "http://dataservice:7070/accounts", bytes.NewReader(postBody))
	tracing.AddTracingToReqFromContext(r.Context(), storeReq)
	resp, err := client.Do(storeReq)

	if err == nil && resp.StatusCode < 299 {
		respData, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(respData, &accountData)
		writeJSONResponse(w, resp.StatusCode, []byte("{\"ID\":\""+accountData.ID+"\"}"))
	} else {
		writeJSONResponse(w, http.StatusInternalServerError, []byte("{\"response\":\""+err.Error()+"\"}"))
	}
}

func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	account := &internalmodel.Account{}
	err = json.Unmarshal(data, &account)
	if err != nil {
		errorMsg := fmt.Sprintf("Error parsing JSON: %v", string(data))
		logrus.Errorln(errorMsg)
		writeJSONResponse(w, http.StatusBadRequest, []byte(errorMsg))
	}

	if account.ID == "" {
		writeJSONResponse(w, http.StatusBadRequest, []byte("PUT body is missing required field: ID"))
		return
	}

	accountData := model.AccountData{ID: account.ID, Name: account.Name}
	putBody, err := json.Marshal(&accountData)

	storeReq, err := http.NewRequest("PUT", "http://dataservice:7070/accounts", bytes.NewReader(putBody))
	tracing.AddTracingToReqFromContext(r.Context(), storeReq)
	resp, err := client.Do(storeReq)
	if err == nil && resp.StatusCode < 299 {
		respData, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(respData, &accountData)
		account.Name = accountData.Name
		account.AccountEvents = accountData.Events
		account.ServedBy = myIP
		outData, _ := json.Marshal(&account)
		writeJSONResponse(w, resp.StatusCode, outData)
	} else {
		writeJSONResponse(w, http.StatusInternalServerError, []byte("{\"response\":\""+err.Error()+"\"}"))
	}
}

// GetAccount loads an account instance, including a quote and an image URL using sub-services.
func GetAccount(w http.ResponseWriter, r *http.Request) {

	// Read the 'accountId' path parameter from the mux map
	var accountID = mux.Vars(r)["accountId"]

	account, err := getAccount(r.Context(), accountID)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	if account.ID == "" {
		writeJSONResponse(w, http.StatusNotFound, []byte("Account identified by '"+accountID+"' not found"))
		return
	}
	account.Quote = getQuote(r.Context())
	account.ImageData = getImageURL(r.Context(), accountID)
	account.ServedBy = myIP

	notifyVIP(r.Context(), account) // Send VIP notification concurrently.

	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	writeJSONResponse(w, http.StatusOK, data)
}

func fetchAccount(ctx context.Context, accountID string) (internalmodel.Account, error) {
	account, err := getAccount(ctx, accountID)
	if err != nil {
		return account, err
	}
	account.Quote = getQuote(ctx)
	account.ImageData = getImageURL(ctx, accountID)
	account.ServedBy = myIP

	notifyVIP(ctx, account) // Send VIP notification concurrently.

	// If found, marshal into JSON, write headers and content
	return account, nil
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(ctx context.Context, account internalmodel.Account) {
	if account.ID == "10000" {
		go func(account internalmodel.Account) {
			vipNotification := internalmodel.VipNotification{AccountID: account.ID, ReadAt: time.Now().UTC().String()}
			data, _ := json.Marshal(vipNotification)
			logrus.Infof("Notifying VIP account %v\n", account.ID)
			err := MessagingClient.PublishOnQueueWithContext(ctx, data, "vip_queue")
			if err != nil {
				logrus.Errorln(err.Error())
			}
			tracing.LogEventToOngoingSpan(ctx, "Sent VIP message")
		}(account)

	}
}

func getQuote(ctx context.Context) internalmodel.Quote {
	// Start a new opentracing child span
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getQuote", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	// Create the http request and pass it to the circuit breaker
	req, err := http.NewRequest("GET", "http://quotes-service:8080/api/quote?strength=4", nil)
	body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-quotes", req)
	if err == nil {
		quote := internalmodel.Quote{}
		json.Unmarshal(body, &quote)
		return quote
	}
	return fallbackQuote
}

func getAccount(ctx context.Context, accountID string) (internalmodel.Account, error) {
	// Start a new opentracing child span
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getAccountData", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	// Create the http request and pass it to the circuit breaker
	req, err := http.NewRequest("GET", "http://dataservice:7070/accounts/"+accountID, nil)
	body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-data", req)
	if err == nil {
		accountData := model.AccountData{}
		json.Unmarshal(body, &accountData)
		return toAccount(accountData), nil
	}
	logrus.Errorf("Error: %v\n", err.Error())
	return internalmodel.Account{}, err
}

func toAccount(accountData model.AccountData) internalmodel.Account {
	return internalmodel.Account{
		ID: accountData.ID, Name: accountData.Name, AccountEvents: accountData.Events,
	}
}

func getImageURL(ctx context.Context, accountID string) model.AccountImage {
	child := tracing.StartSpanFromContextWithLogEvent(ctx, "getImageUrl", "Client send")
	defer tracing.CloseSpan(child, "Client Receive")

	req, err := http.NewRequest("GET", "http://imageservice:7777/accounts/"+accountID, nil)
	body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "account-to-image", req)
	if err == nil {
		accountImage := model.AccountImage{}
		err := json.Unmarshal(body, &accountImage)
		if err == nil {
			return accountImage
		}
		panic("Unmarshalling accountImage struct went really bad. Msg: " + err.Error())
	}
	return model.AccountImage{URL: "http://path.to.placeholder", ServedBy: "fallback"}
}

// HealthCheck will return OK if the underlying BoltDB is healthy. At least healthy enough for demoing purposes.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
	dbUp := true
	if dbUp && isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJSONResponse(w, http.StatusOK, data)
		logrus.Infoln("Wrote health respo OK")
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
		writeJSONResponse(w, http.StatusServiceUnavailable, data)
	}
}

// SetHealthyState can be used fake health problems.
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

func writeJSONResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Connection", "close")
	w.WriteHeader(status)
	w.Write(data)
}

type healthCheckResponse struct {
	Status string `json:"status"`
}
