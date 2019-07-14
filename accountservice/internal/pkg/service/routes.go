package service

import (
	. "github.com/callistaenterprise/goblog/common/router"
	gqlhandler "github.com/graphql-go/graphql-go-handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialize our routes
var routes = Routes{
	Route{
		"StoreAccount", // Name
		"POST",         // HTTP method
		"/accounts",    // Route pattern
		StoreAccount,
		true,
	},
	Route{
		"UpdateAccount", // Name
		"PUT",           // HTTP method
		"/accounts",     // Route pattern
		UpdateAccount,
		true,
	},
	Route{
		"GetAccount",            // Name
		"GET",                   // HTTP method
		"/accounts/{accountId}", // Route pattern
		GetAccount,
		true,
	},
	Route{
		"GraphQL",  // Name
		"POST",     // HTTP method
		"/graphql", // Route pattern
		gqlhandler.New(&gqlhandler.Config{
			Schema: &schema,
			Pretty: false,
		}).ServeHTTP,
		true,
	},
	Route{
		"HealthCheck",
		"GET",
		"/health",
		HealthCheck,
		false,
	},
	Route{
		"Testability",
		"GET",
		"/testability/healthy/{state}",
		SetHealthyState,
		false,
	},
	Route{
		"Prometheus",
		"GET",
		"/metrics",
		promhttp.Handler().ServeHTTP,
		false,
	},
}
