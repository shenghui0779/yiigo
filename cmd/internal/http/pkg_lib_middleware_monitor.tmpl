package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "demo",
		Subsystem: "api",
		Name:      "requests_count",
		Help:      "The total number of http request",
	}, []string{"method", "path", "status"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "demo",
		Subsystem: "api",
		Name:      "duration_seconds",
		Help:      "The http request latency in seconds",
	}, []string{"method", "path", "status"})
)

func init() {
	prometheus.MustRegister(httpRequestCounter)
	prometheus.MustRegister(httpRequestDuration)
}

// Monitor 监控请求次数，时长
func Monitor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now().Local()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		httpRequestCounter.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": strconv.Itoa(ww.Status()),
		}).Inc()

		next.ServeHTTP(ww, r)

		httpRequestDuration.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": strconv.Itoa(ww.Status()),
		}).Observe(time.Since(begin).Seconds())
	})
}
