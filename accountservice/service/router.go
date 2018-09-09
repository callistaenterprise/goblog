package service

import (
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

var summaryMap = make(map[string]*prometheus.SummaryVec)

// var histogramMap = make(map[string]prometheus.Histogram)
// NewRouter creates a mux.Router pointer.
func NewRouter() *mux.Router {

	initQL(&LiveGraphQLResolvers{})

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {

		// If route should be monitored, create summaryVec for endpoint.
		if route.Monitor {
			summaryMap[route.Name] = buildSummaryVec(route.Name, route.Method+" "+route.Pattern)
		}

		// Add route to router, including middleware chaining.
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(withMonitoring(withTracing(route.HandlerFunc, route), route))
	}

	logrus.Infoln("Successfully initialized routes including Prometheus.")
	return router
}

func withTracing(next http.Handler, route Route) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, route.Name)
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
