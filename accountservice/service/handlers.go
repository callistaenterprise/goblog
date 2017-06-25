package service

import (
        "net/http"
        "github.com/gorilla/mux"
        "encoding/json"
        "github.com/callistaenterprise/goblog/accountservice/dbclient"
        "github.com/callistaenterprise/goblog/common/messaging"
        "github.com/callistaenterprise/goblog/accountservice/model"
        "time"
        "github.com/Sirupsen/logrus"
        "io/ioutil"
        "strconv"
        "github.com/callistaenterprise/goblog/common/util"
        "github.com/callistaenterprise/goblog/common/tracing"
        "github.com/opentracing/opentracing-go"
        "github.com/afex/hystrix-go/hystrix"
)

var DBClient dbclient.IGormClient
var MessagingClient messaging.IMessagingClient
var isHealthy = true

var client = &http.Client{}

func init() {
        var transport http.RoundTripper = &http.Transport{
                DisableKeepAlives: true,
        }
        client.Transport = transport
}

func GetAccount(w http.ResponseWriter, r *http.Request) {
        span := tracing.StartHTTPTrace(r, "GetAccount")
        defer span.Finish()

	// Read the 'accountId' path parameter from the mux map
        var accountId = mux.Vars(r)["accountId"]

        // Read the account struct
        child := tracing.Tracer.StartSpan("QueryAccount", opentracing.ChildOf(span.Context()))
        account, err := DBClient.QueryAccount(accountId)
        child.Finish()

        child = tracing.Tracer.StartSpan("GetIP", opentracing.ChildOf(span.Context()))
        account.ServedBy = util.GetIP()
        child.Finish()

        // If err, return a 404
        if err != nil {
                logrus.Errorf("Some error occured serving " + accountId + ": " + err.Error())
                w.WriteHeader(http.StatusNotFound)
                return
        }

        notifyVIP(account)   // Send VIP notification concurrently.

        // NEW call the quotes-service
        child = tracing.Tracer.StartSpan("getQuote", opentracing.ChildOf(span.Context()))
        account.Quote, err = getQuote(child)
        child.Finish()
        
        child = tracing.Tracer.StartSpan("GetImage", opentracing.ChildOf(span.Context()))
        account.ImageUrl, err = getImageUrl(accountId, child)
        child.Finish()
        
        // If found, marshal into JSON, write headers and content
        data, _ := json.Marshal(account)
        writeJsonResponse(w, http.StatusOK, data)
        logrus.Infof("Successfully served account %v", accountId)
}

func callWithCircuitBreaker(breakerName string, span opentracing.Span, url string) ([]byte, error) {
        output := make(chan []byte, 1)
        errors := hystrix.Go(breakerName, func() error {
                // talk to other services
                req, _ := http.NewRequest("GET", url, nil)
                tracing.AddTracingToReq(req, span)

                resp, err := client.Do(req)
                if err != nil {
                        logrus.Errorf("Request to %v returned error: %v", breakerName, err.Error())
                        return err
                }
                responseBody, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                        logrus.Errorf("Error reading %v response: %v", breakerName, err.Error())
                        return err
                }
                output <- responseBody

                // A bit ugly, return nil to indicate nothing bad happened.
                return nil
        }, func(err error) error {
                logrus.Errorf("Breaker %v opened, error: %v", breakerName, err.Error())
                return err
        })

        select {
        case out := <-output:
                return out, nil

        case err := <-errors:
                return nil, err
        }
        return nil, nil
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(account model.Account) {
        if account.ID == "10000" {
                go func(account model.Account) {
                        vipNotification := model.VipNotification{AccountId: account.ID, ReadAt: time.Now().UTC().String()}
                        data, _ := json.Marshal(vipNotification)
                        logrus.Printf("Notifying VIP account %v\n", account.ID)
                        err := MessagingClient.PublishOnQueue(data, "vip_queue")
                        if err != nil {
                                logrus.Errorln(err.Error())
                        }
                }(account)
        }
}

var fallbackQuote = model.Quote{
        Language:"en",
        ServedBy: "circuit-breaker",
        Text: "May the source be with you, always."}

func getQuote(span opentracing.Span) (model.Quote, error) {
        body, err := callWithCircuitBreaker("quotes-service", span, "http://quotes-service:8080/api/quote?strength=4")

        if err == nil {
                quote := model.Quote{}
                json.Unmarshal(body, &quote)
                return quote, nil
        } else {
                logrus.Errorf("Got error getting quote: %v", err.Error())
                return fallbackQuote, err
        }
}

func getImageUrl(accountId string, span opentracing.Span) (string, error) {

        body, err := callWithCircuitBreaker("imageservice", span, "http://imageservice:7777/" + accountId)

        if err == nil {
                return string(body), nil
        } else {
                logrus.Errorf("Got error getting imageUrl: %v", err.Error())
                return "http://path.to.placeholder", err
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
                logrus.Println("Invalid request to SetHealthyState, allowed values are true or false")
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

