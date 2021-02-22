package prometheus

import (
	"github.com/labstack/echo/v4"
)

type Prometheus struct {
}

func NewPrometheus() *Prometheus {
	return nil
}

func (p *Prometheus) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
