package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/callistaenterprise/goblog/accountservice/dbclient"
	"github.com/callistaenterprise/goblog/accountservice/model"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/gorilla/mux"
)

var DBClient dbclient.IBoltClient
var MessagingClient messaging.IMessagingClient
var isHealthy = true

var client = &http.Client{}

var LOGGER = logrus.Logger{}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
	LOGGER.Infof("Successfully initialized transport")
}

func GetAccount(w http.ResponseWriter, r *http.Request) {
	// Read the 'accountId' path parameter from the mux map
	var accountId = mux.Vars(r)["accountId"]

	// Read the account struct BoltDB
	account, err := DBClient.QueryAccount(accountId)
	account.ServedBy = util.GetIP()

	// If err, return a 404
	if err != nil {
		logrus.Errorf("Some error occured serving " + accountId + ": " + err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	notifyVIP(account) // Send VIP notification concurrently.

	// NEW call the quotes-service
	quote, err := getQuote()
	if err == nil {
		account.Quote = quote
	}

	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	writeJsonResponse(w, http.StatusOK, data)
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(account model.Account) {
	if account.Id == "10000" {
		go func(account model.Account) {
			vipNotification := model.VipNotification{AccountId: account.Id, ReadAt: time.Now().UTC().String()}
			data, _ := json.Marshal(vipNotification)
			logrus.Infof("Notifying VIP account %v\n", account.Id)
			err := MessagingClient.PublishOnQueue(data, "vip_queue")
			if err != nil {
				logrus.Errorln(err.Error())
			}
		}(account)
	}
}

func getQuote() (model.Quote, error) {
	req, _ := http.NewRequest("GET", "http://quotes-service:8080/api/quote?strength=4", nil)
	resp, err := client.Do(req)

	if err == nil && resp.StatusCode == 200 {
		quote := model.Quote{}
		bytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bytes, &quote)
		return quote, nil
	} else {
		return model.Quote{}, fmt.Errorf("Some error")
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
