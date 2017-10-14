package service

import "net/http"

// Route type, derived from http://thenewstack.io/make-a-restful-json-api-go/
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is a type of slice of Route
type Routes []Route

var routes = Routes{
	Route{
		"HealthCheck",
		"GET",
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			w.Write([]byte("OK"))
		},
	},
}
