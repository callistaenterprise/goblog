package service

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/callistaenterprise/goblog/common/tracing"
)

/**
 * From http://thenewstack.io/make-a-restful-json-api-go/
 */
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(loadTracing(handler))
	}
	return router
}

func loadTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, "imageservice")
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
