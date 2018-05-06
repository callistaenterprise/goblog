package service

import (
	"net/http"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/gorilla/mux"

	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/Sirupsen/logrus"
	"strings"
)

var requestsDurationHistogram *prometheus.HistogramVec

// NewRouter creates a mux.Router pointer.
func NewRouter() *mux.Router {

	initQL(&LiveGraphQLResolvers{})

	router := mux.NewRouter().StrictSlash(true)
	allRoutes := make([]string, 0)
	for _, route := range routes {
		allRoutes = append(allRoutes, strings.ToLower(route.Name))
	}
	requestsDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Request duration distribution",
			Buckets: prometheus.LinearBuckets(0, 5, 10),
		},
		allRoutes,
	)
	logrus.Infof("The histogram %v", requestsDurationHistogram)
	prometheus.MustRegister(requestsDurationHistogram)

	for _, route := range routes {
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(measure(route.HandlerFunc, route.Name))
			//Handler(measure(loadTracing(route.HandlerFunc, route.Name), route.Name))
	}


	logrus.Infoln("Successfully initialized routes including Prometheus.")
	return router
}



func measure(next http.Handler, handlerName string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()
		logrus.Infoln("ENTER - measure for " + handlerName)
		next.ServeHTTP(rw, req)
		duration := time.Since(start)
		logrus.Infof("EXIT - measure with duration: %v\n", duration.Seconds())
		obs, err := requestsDurationHistogram.GetMetricWithLabelValues(strings.ToLower(handlerName))
		if err != nil {
			logrus.Errorf("Error: %v", err)
		}
		logrus.Infoln("Got an OBS")
		obs.Observe(duration.Seconds())
		logrus.Infof("Efter\n")
	})
}

func loadTracing(next http.Handler, handlerName string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, handlerName)
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
