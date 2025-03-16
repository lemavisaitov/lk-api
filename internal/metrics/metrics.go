package metrics

import (
	"net/http"
	"runtime"
	"time"

	"github.com/lemavisaitov/lk-api/internal/cache"
	"github.com/lemavisaitov/lk-api/internal/logger"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
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
		server := &http.Server{
			Addr:         ":" + port,
			ReadTimeout:  5 * time.Second,   // Таймаут чтения
			WriteTimeout: 10 * time.Second,  // Таймаут записи
			IdleTimeout:  120 * time.Second, // Таймаут простоя
		}
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal("Failed to start metrics server: %v",
				zap.Error(errors.Wrap(err, "")))
		}
	}()
}

func HttpStatusMetricInc(statusCode int, method string) {
	HttpStatusMetric.WithLabelValues(http.StatusText(statusCode), method).Inc()
}

var c *cache.CacheDecorator

func GetCacheMetrics() float64 {
	return float64(c.MemoryUsage())
}
