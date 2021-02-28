package prometheus

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	DefaultHistogramOpts = prometheus.HistogramOpts{
		Namespace: "echo",
		Subsystem: "",
		Name:      "requests",
		Help:      "A histogram of request times and status codes",
		Buckets:   prometheus.DefBuckets,
	}
	DefaultPrometheusConfig = PrometheusConfig{
		Registerer:    prometheus.DefaultRegisterer,
		Skipper:       middleware.DefaultSkipper,
		HistogramOpts: DefaultHistogramOpts,
	}
	DefaultExposeConfig = ExposeConfig{
		Gatherer:    prometheus.DefaultGatherer,
		HandlerOpts: promhttp.HandlerOpts{},
	}
)

type PrometheusConfig struct {
	Registerer    prometheus.Registerer
	Skipper       middleware.Skipper
	HistogramOpts prometheus.HistogramOpts
}

func Prometheus() echo.MiddlewareFunc {
	return PrometheusWithConfig(DefaultPrometheusConfig)
}

func PrometheusWithConfig(config PrometheusConfig) echo.MiddlewareFunc {
	requestHistogram := prometheus.NewHistogramVec(config.HistogramOpts, []string{"status", "path"})
	config.Registerer.MustRegister(requestHistogram)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			start := time.Now()
			var status int

			err := next(c)
			if err != nil {
				status = guessStatus(err)
			} else {
				status = c.Response().Status
			}

			elapsed := time.Since(start)
			requestHistogram.WithLabelValues(fmt.Sprint(status), c.Path()).Observe(elapsed.Seconds())

			return err
		}
	}
}

type ExposeConfig struct {
	Gatherer    prometheus.Gatherer
	HandlerOpts promhttp.HandlerOpts
}

func Expose() echo.HandlerFunc {
	return ExposeWithConfig(DefaultExposeConfig)
}

func ExposeWithConfig(config ExposeConfig) echo.HandlerFunc {
	h := promhttp.HandlerFor(config.Gatherer, config.HandlerOpts)
	return echo.WrapHandler(h)
}

// guessStatus attempts to estimate the status code of an error from echo.
// Because of a design flaw in echo, we can't invoke the error handler
// in order to set the status code. This is because returning the error
// later in order to keep the error handling of other middlewares will
// trigger error handling again.
func guessStatus(e error) int {
	h, ok := e.(*echo.HTTPError)
	if !ok {
		return http.StatusInternalServerError
	}
	return h.Code
}
