package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/trace"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, handler, and status code.",
	}, []string{"method", "handler", "code"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"method", "handler", "code"})
)

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.code = code
	sw.ResponseWriter.WriteHeader(code)
}

// Instrument wraps h, recording request count and latency with the given handler label.
func Instrument(pattern string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}
		start := time.Now()
		h.ServeHTTP(sw, r)
		duration := time.Since(start).Seconds()
		code := strconv.Itoa(sw.code)
		httpRequestsTotal.WithLabelValues(r.Method, pattern, code).Inc()
		obs := httpRequestDuration.WithLabelValues(r.Method, pattern, code)
		// Gap E: embed trace ID as exemplar so Grafana can link histogram samples
		// to Tempo traces. otelhttp sets a span on the context before calling this
		// handler, so SpanFromContext is valid when traces are flowing.
		if sc := trace.SpanFromContext(r.Context()).SpanContext(); sc.IsValid() {
			if eo, ok := obs.(prometheus.ExemplarObserver); ok {
				eo.ObserveWithExemplar(duration, prometheus.Labels{"traceID": sc.TraceID().String()})
				return
			}
		}
		obs.Observe(duration)
	})
}
