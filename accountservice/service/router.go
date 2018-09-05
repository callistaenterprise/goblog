package service

import (
	"net/http"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/gorilla/mux"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

var metricMap = make(map[string]*prometheus.SummaryVec)
var histogramMap = make(map[string]prometheus.Histogram)
// NewRouter creates a mux.Router pointer.
func NewRouter() *mux.Router {

	initQL(&LiveGraphQLResolvers{})

	router := mux.NewRouter().StrictSlash(true)
	//allRoutes := make([]string, 0)
	for _, route := range routes {
		//allRoutes = append(allRoutes, strings.ToLower(route.Name))
		if route.Monitor {
			metricMap[route.Name] = buildSummaryVec(route.Name, route.Method + " " + route.Pattern)
			histogramMap[route.Name] = buildHistogram(route.Name + "_histogram", route.Method + " " + route.Pattern)
		}
	}

	for _, route := range routes {
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(measure(loadTracing(route.HandlerFunc, route), route))
	}

	logrus.Infoln("Successfully initialized routes including Prometheus.")
	return router
}

func loadTracing(next http.Handler, route Route) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, route.Name)
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
