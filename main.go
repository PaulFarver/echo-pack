package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	eprom "github.com/paulfarver/echo-pack/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	server := simple()
	server2, promServer := advanced()
	go startServer(server2, 8081)
	go startServer(promServer, 9090)
	startServer(server, 8080)
}

func startServer(e *echo.Echo, port int) {
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		logrus.Fatal(err)
	}
}

func simple() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	registerExampleRoutes(e)

	count := promauto.NewCounter(prometheus.CounterOpts{Name: "mymetric"})
	e.GET("/tick", func(c echo.Context) error {
		count.Inc()
		return c.String(http.StatusOK, "Incremented metric")
	})

	// We include the prometheus middleware to get request statistics in prometheus
	e.Use(eprom.Prometheus())
	e.GET("/metrics", eprom.Expose())

	return e
}

func advanced() (server, promServer *echo.Echo) {
	server = echo.New()
	server.HideBanner = true
	registerExampleRoutes(server)

	promServer = echo.New()
	promServer.HideBanner = true

	count := prometheus.NewCounter(prometheus.CounterOpts{Name: "mymetric"})
	r := prometheus.NewRegistry()
	r.MustRegister(count)
	// We can include the prometheus middleware on one server
	server.Use(eprom.PrometheusWithConfig(eprom.PrometheusConfig{
		Registerer:    r,
		Skipper:       middleware.DefaultSkipper,
		HistogramOpts: eprom.DefaultHistogramOpts,
	}))
	server.GET("/tick", func(c echo.Context) error {
		count.Inc()
		return c.String(http.StatusOK, "Incremeneted metric")
	})

	// But expose the metrics on another server
	promServer.GET("/metrics", eprom.ExposeWithConfig(eprom.ExposeConfig{
		Gatherer:    r,
		HandlerOpts: promhttp.HandlerOpts{},
	}))

	return server, promServer
}

func registerExampleRoutes(e *echo.Echo) {
	e.GET("/hello", func(c echo.Context) error {
		logrus.Info("Hello world")
		return c.NoContent(http.StatusNoContent)
	})
	e.GET("/hello/:handle", func(c echo.Context) error {
		logrus.Info(c.Param("handle"))
		return c.String(http.StatusOK, c.Param("handle"))
	})
	e.GET("/error", func(c echo.Context) error {
		err := errors.New("Failed to handle request")
		logrus.WithError(err).Error("Failed for some reason")
		return err
	})
	e.GET("/forbidden", func(c echo.Context) error {
		return echo.ErrForbidden
	})
}
