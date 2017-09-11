package service

import "net/http"

/**
 * Derived from http://thenewstack.io/make-a-restful-json-api-go/
 */
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{

	Route{
		"ProcessImage",
		"POST",
		"/image",
		ProcessImage,
	},
	Route{
		"ProcessImage",
		"GET",
		"/file/{filename}",
		ProcessImageFromFile,
	},
	Route{
		"GetAccountImage",
		"GET",
		"/accounts/{accountId}",
		GetAccountImage,
	},
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
