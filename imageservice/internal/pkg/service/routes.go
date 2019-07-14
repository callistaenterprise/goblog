package service

import (
	"github.com/callistaenterprise/goblog/common/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

/**
 * Derived from http://thenewstack.io/make-a-restful-json-api-go/
 */

var routes = router.Routes{

	router.Route{
		"ProcessImage",
		"GET",
		"/file/{filename}",
		ProcessImageFromFile,
		true,
	},
	router.Route{
		"GetAccountImage",
		"GET",
		"/accounts/{accountId}",
		GetAccountImage,
		true,
	},
	router.Route{
		"UpdateAccountImage",
		"PUT",
		"/accounts",
		UpdateAccountImage,
		true,
	},
	router.Route{
		"CreateAccountImage",
		"POST",
		"/accountS",
		CreateAccountImage,
		true,
	},
	router.Route{
		"HealthCheck",
		"GET",
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			w.Write([]byte("OK"))
		},
		false,
	},
	router.Route{
		"Prometheus",
		"GET",
		"/metrics",
		promhttp.Handler().ServeHTTP,
		false,
	},
}
