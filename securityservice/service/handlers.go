package service

import (
	"net/http"

	"fmt"

	"github.com/callistaenterprise/goblog/common/circuitbreaker"
	ct "github.com/eriklupander/cloudtoolkit"
)

var client = &http.Client{}

func ConfigureHttpClient() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
}

/**
 * Takes the POST body, decodes, processes and finally writes the result to the response.
 */
func SecuredGetAccount(w http.ResponseWriter, r *http.Request) {

	url := fmt.Sprintf("http://accountservice:6767%s", r.URL.Path)
	req, err := http.NewRequest("GET", url, nil)
	body, err := circuitbreaker.PerformHTTPRequestCircuitBreaker(r.Context(), "securityservice->accountservice", req)

	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	} else {
		writeServerError(w, err.Error())
	}
}

func writeServerError(w http.ResponseWriter, msg string) {
	ct.Log.Println(msg)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}
