package prometheus

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	Registry         *prometheus.Registry
	Skipper          middleware.Skipper
	RequestHistogram *prometheus.HistogramVec
	PathLabeler      func(c echo.Context) string
}

func DefaultPathLabeler(c echo.Context) string {
	return c.Path()
}

func NewPrometheus() (*Prometheus, error) {
	r := prometheus.NewRegistry()
	requestHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "",
		Name:      "requests",
		Help:      "",
		Buckets:   prometheus.DefBuckets,
	}, []string{"status", "path"})
	if err := r.Register(requestHistogram); err != nil {
		return nil, err
	}
	return &Prometheus{
		Registry:         r,
		Skipper:          middleware.DefaultSkipper,
		RequestHistogram: requestHistogram,
		PathLabeler:      DefaultPathLabeler,
	}, nil
}

func (p *Prometheus) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if p.Skipper(c) {
			return next(c)
		}

		start := time.Now()

		res := next(c)

		status := c.Response().Status
		elapsed := time.Since(start)

		p.RequestHistogram.WithLabelValues(fmt.Sprint(status), p.PathLabeler(c)).Observe(elapsed.Seconds())

		return res
	}
}

func (p *Prometheus) Expose() echo.HandlerFunc {
	h := promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{})
	return echo.WrapHandler(h)
}
