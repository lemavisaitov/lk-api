package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/lemavisaitov/lk-api/internal/cache"
	"github.com/lemavisaitov/lk-api/internal/logger"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"net/http"
	"runtime"
)

var (
	HttpStatusMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Count of HTTP responses, labeled by status code and method",
		},
		[]string{"status", "method"},
	)
	GoroutinesMetric = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "num_goroutines",
			Help: "Current number of goroutines",
		},
		func() float64 {
			return float64(runtime.NumGoroutine())
		},
	)
	CPUNumMetric = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "num_of_cpu",
			Help: "Current number of CPU (Machienes)",
		},
		func() float64 {
			return float64(runtime.NumCPU())
		},
	)
	CacheMemoryUsage = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "cache_memory_usage",
			Help: "Amount of memory occupied by the cache",
		},
		GetCacheMetrics,
	)
)

func InitMetrics(port string, cache *cache.CacheDecorator) {
	c = cache
	prometheus.MustRegister(GoroutinesMetric)
	prometheus.MustRegister(HttpStatusMetric)
	prometheus.MustRegister(CacheMemoryUsage)
	prometheus.MustRegister(CPUNumMetric)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Info("starting metrics server",
			zap.String("port", port),
		)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logger.Fatal("Failed to start metrics server: %v",
				zap.Error(errors.Wrap(err, "")))
		}
	}()
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		statusCode := c.Writer.Status()
		method := c.Request.Method
		HttpStatusMetric.WithLabelValues(http.StatusText(statusCode), method).Inc()
		c.Next()
	}
}

var c *cache.CacheDecorator

func GetCacheMetrics() float64 {
	return float64(c.MemoryUsage())
}
