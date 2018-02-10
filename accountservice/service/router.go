package service

import (
	"net/http"
	 gqlhandler "github.com/graphql-go/graphql-go-handler"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/gorilla/mux"
)

// NewRouter creates a mux.Router pointer.
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {

		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(loadTracing(route.HandlerFunc, route.Name))

	}

	initQL()

	h := gqlhandler.New(&gqlhandler.Config{
		Schema: &Schema,
		Pretty: true,
	})

	router.Methods("POST").
	Name("Graphql").Path("/graphql").Handler(h)
	return router
}

func loadTracing(next http.Handler, handlerName string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, handlerName)
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
