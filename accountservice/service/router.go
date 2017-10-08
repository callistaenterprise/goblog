package service

import (
	"github.com/gorilla/mux"
	"net/http"
	"context"
        "github.com/callistaenterprise/goblog/common/tracing"
)

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {

		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(loadTracing(route.HandlerFunc))

	}
	return router
}

func loadTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
                span := tracing.StartHTTPTrace(req, "GetAccount")
                defer span.Finish()
		ctx := context.WithValue(req.Context(), "opentracing-span", span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
