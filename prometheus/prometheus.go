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
	r              *prometheus.Registry
	Skipper        middleware.Skipper
	requestCounter *prometheus.CounterVec
	requestSummary *prometheus.SummaryVec
	pathLabeler    func(c echo.Context) string
}

func DefaultPathLabeler(c echo.Context) string {
	return c.Path()
}

func NewPrometheus() *Prometheus {
	r := prometheus.NewRegistry()
	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "",
		Name:      "",
		Help:      "",
	}, []string{"status", "path"})
	requestSummary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "",
		Name:       "",
		Help:       "",
		Objectives: map[float64]float64{},
	}, []string{"status", "path"})
	r.Register(requestCounter)
	return &Prometheus{
		r:              r,
		Skipper:        middleware.DefaultSkipper,
		requestCounter: requestCounter,
		requestSummary: requestSummary,
	}
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

		p.requestCounter.WithLabelValues(fmt.Sprint(status), p.pathLabeler(c)).Inc()
		p.requestSummary.WithLabelValues(fmt.Sprint(status), p.pathLabeler(c)).Observe(elapsed.Seconds())

		return res
	}
}

func (p *Prometheus) Expose() echo.HandlerFunc {
	h := promhttp.HandlerFor(p.r, promhttp.HandlerOpts{})
	return echo.WrapHandler(h)
}
