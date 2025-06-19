package metric

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	username        string
	password        string
}

type Option func(*Metrics)

func WithBasicAuth(username, password string) Option {
	return func(m *Metrics) {
		m.username = username
		m.password = password
	}
}

func NewMetrics(opts ...Option) *Metrics {
	m := &Metrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"path", "method", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"path", "method", "status"},
		),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path // for 404 path
		}

		m.requestsTotal.WithLabelValues(path, c.Request.Method, status).Inc()
		m.requestDuration.WithLabelValues(path, c.Request.Method, status).Observe(duration.Seconds())
	}
}

func (m *Metrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.username != "" && m.password != "" {
			user, pass, ok := c.Request.BasicAuth()
			if !ok || user != m.username || pass != m.password {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}
