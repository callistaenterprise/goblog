package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"flag"
	"time"
	"net/http"
	"github.com/spf13/viper"
)

var (
	addr              = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	uniformDomain     = flag.Float64("uniform.domain", 0.0002, "The domain for the uniform distribution.")
	normDomain        = flag.Float64("normal.domain", 0.0002, "The domain for the normal distribution.")
	normMean          = flag.Float64("normal.mean", 0.00001, "The mean for the normal distribution.")
	oscillationPeriod = flag.Duration("oscillation-period", 10*time.Minute, "The duration of the rate oscillation period.")
)

func buildMetric(metricName string, metricHelp string) *prometheus.SummaryVec {
	rpcDuration := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: viper.GetString("service_name"),
			Name:      metricName,
			Help:      metricHelp,
		},
		[]string{"service"},
	)
	prometheus.Register(rpcDuration)

	return rpcDuration
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
