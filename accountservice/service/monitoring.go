package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"net/http"
	"github.com/spf13/viper"
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

func buildHistogram(metricName string, metricHelp string) prometheus.Histogram {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: viper.GetString("service_name"),
		Name:      metricName,
		Help:      metricHelp,
		Buckets:   prometheus.LinearBuckets(0, 0.001, 20),
	})
	prometheus.Register(histogram)
	return histogram
}

func measure(next http.Handler, route Route) http.Handler {

	// Just return the next handler if route shouldn't be monitored
	if !route.Monitor {
		return next
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(rw, req)
		duration := time.Since(start)
		// Get duration holders
		metric := metricMap[route.Name]
		histogram := histogramMap[route.Name]

		// Store values.
		metric.WithLabelValues("normal").Observe(duration.Seconds())
		histogram.Observe(duration.Seconds())
	})
}
