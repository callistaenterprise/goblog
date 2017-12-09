package service

import (
	"net/http"

	"github.com/callistaenterprise/goblog/common/tracing"
	ct "github.com/eriklupander/cloudtoolkit"
	"github.com/gorilla/mux"
)

/**
 * From http://thenewstack.io/make-a-restful-json-api-go/
 */
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range securedRoutes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = loadTracing(handler)
		handler = ct.OAuth2Handler(handler)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	for _, route := range unsecuredRoutes {
		var handler http.Handler

		handler = route.HandlerFunc
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

func loadTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, "GetAccountSecured")
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
