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

// var DBClient dbclient.IBoltClient
var DBClient dbclient.IGormClient

// MessagingClient instance
var MessagingClient messaging.IMessagingClient

var myIp string
var isHealthy = true
var client = &http.Client{}

var fallbackQuote = model.Quote{
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
    myIp, err = util.ResolveIPFromHostsFile()
    if err != nil {
        myIp = util.GetIP()
    }
}

// GetAccount loads an account instance, including a quote and an image URL using sub-services.
func GetAccount(w http.ResponseWriter, r *http.Request) {

    // Read the 'accountId' path parameter from the mux map
    var accountID = mux.Vars(r)["accountId"]

    // Read the account struct BoltDB
    account, err := DBClient.QueryAccount(r.Context(), accountID)
    account.ServedBy = myIp

    // If err, return a 404
    if err != nil {
        logrus.Errorf("Some error occured serving " + accountID + ": " + err.Error())
        w.WriteHeader(http.StatusNotFound)
        return
    }

    notifyVIP(r.Context(), account) // Send VIP notification concurrently.

    account.Quote = getQuote(r.Context())
    account.ImageData = getImageURL(r.Context(), accountID)

    // If found, marshal into JSON, write headers and content
    data, _ := json.Marshal(account)
    writeJSONResponse(w, http.StatusOK, data)
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(ctx context.Context, account model.Account) {
    if account.ID == "10000" {
        go func(account model.Account) {
            vipNotification := model.VipNotification{AccountID: account.ID, ReadAt: time.Now().UTC().String()}
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

func getQuote(ctx context.Context) (model.Quote) {
    // Start a new opentracing child span
    child := tracing.StartSpanFromContextWithLogEvent(ctx, "getQuote", "Client send")
    defer tracing.CloseSpan(child, "Client Receive")

    // Create the http request and pass it to the circuit breaker
    req, err := http.NewRequest("GET", "http://quotes-service:8080/api/quote?strength=4", nil)
    body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "accountservice->quotes-service", req)
    if err == nil {
        quote := model.Quote{}
        json.Unmarshal(body, &quote)
        return quote
    }
    return fallbackQuote
}

func getImageURL(ctx context.Context, accountID string) (model.AccountImage) {
    child := tracing.StartSpanFromContextWithLogEvent(ctx, "getImageUrl", "Client send")
    defer tracing.CloseSpan(child, "Client Receive")

    req, err := http.NewRequest("GET", "http://imageservice:7777/accounts/"+accountID, nil)
    body, err := cb.PerformHTTPRequestCircuitBreaker(tracing.UpdateContext(ctx, child), "accountservice->imageservice", req)
    if err == nil {
        accountImage := model.AccountImage{}
        err := json.Unmarshal(body, &accountImage)
        if err == nil {
            return accountImage
        } else {
            panic("Unmarshalling accountImage struct went really bad. Msg: " + err.Error())
        }

    }
    return model.AccountImage{URL: "http://path.to.placeholder", ServedBy: "fallback"}
}

// HealthCheck will return OK if the underlying BoltDB is healthy. At least healthy enough for demoing purposes.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
    dbUp := DBClient.Check()
    if dbUp && isHealthy {
        data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
        writeJSONResponse(w, http.StatusOK, data)
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
