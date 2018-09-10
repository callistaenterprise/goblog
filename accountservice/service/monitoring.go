package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
	"time"
)

func buildSummaryVec(metricName string, metricHelp string) *prometheus.SummaryVec {
	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: viper.GetString("service_name"),
			Name:      metricName,
			Help:      metricHelp,
		},
		[]string{"service"},
	)
	prometheus.Register(summaryVec)
	return summaryVec
}

func withMonitoring(next http.Handler, route Route) http.Handler {

	// Just return the next handler if route shouldn't be monitored
	if !route.Monitor {
		return next
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(rw, req)
		duration := time.Since(start)

		// Get summary holder
		summary := summaryMap[route.Name]

		// Store duration of request
		summary.WithLabelValues("duration").Observe(duration.Seconds())

		// Store size of response, if possible.
		size, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		if err == nil {
			summary.WithLabelValues("size").Observe(float64(size))
		}
	})
}
