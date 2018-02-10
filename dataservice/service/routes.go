package service

import "net/http"

// Route defines a single route, e.g. a human readable name, HTTP method, pattern the function that will execute when the route is called.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes defines the type Routes which is just an array (slice) of Route structs.
type Routes []Route

// Initialize our routes
var routes = Routes{
	Route{
		"GetAccountByNameWithCount", // Name
		"GET", // HTTP method
		"/accountsbyname/{accountName}", // Route pattern
		GetAccountByNameWithCount,
	},
	Route{
		"GetAccount", // Name
		"GET",        // HTTP method
		"/accounts/{accountId}", // Route pattern
		GetAccount,
	},
	Route{
		"StoreAccount", // Name
		"POST",         // HTTP method
		"/accounts",    // Route pattern
		StoreAccount,
	},
	Route{
		"UpdateAccount", // Name
		"PUT",           // HTTP method
		"/accounts",     // Route pattern
		UpdateAccount,
	},
	Route{
		"HealthCheck",
		"GET",
		"/health",
		func(writer http.ResponseWriter, r *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
		},
	},
}
