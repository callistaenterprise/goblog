package service

import (
	"github.com/callistaenterprise/goblog/common/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Initialize our routes
var routes = router.Routes{
	router.Route{
		"GetAccountByNameWithCount", // Name
		"GET", // HTTP method
		"/accountsbyname/{accountName}", // Route pattern
		GetAccountByNameWithCount,
		true,
	},
	router.Route{
		"LoadAccount", // Name
		"GET",         // HTTP method
		"/accounts/{accountId}", // Route pattern
		GetAccount,
		true,
	},
	router.Route{
		"StoreAccount", // Name
		"POST",         // HTTP method
		"/accounts",    // Route pattern
		StoreAccount,
		true,
	},
	router.Route{
		"UpdateAccount", // Name
		"PUT",           // HTTP method
		"/accounts",     // Route pattern
		UpdateAccount,
		true,
	},
	router.Route{
		"RandomAccount", // Name
		"GET",           // HTTP method
		"/random",       // Route pattern
		RandomAccount,
		false,
	},
	router.Route{
		"Seed",  // Name
		"GET",   // HTTP method
		"/seed", // Route pattern
		SeedAccounts,
		false,
	},
	router.Route{
		"HealthCheck",
		"GET",
		"/health",
		func(writer http.ResponseWriter, r *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
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
