// Package metrics holds thumbnail-service business metric counters.
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	Generated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "thumbnail_generated_total",
		Help: "Total number of thumbnails generated.",
	})

	Errors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "thumbnail_errors_total",
		Help: "Total number of thumbnail generation failures.",
	})

	Duration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "thumbnail_duration_seconds",
		Help:    "Thumbnail generation duration in seconds.",
		Buckets: prometheus.DefBuckets,
	})
)

func init() {
	prometheus.MustRegister(Generated, Errors, Duration)
}