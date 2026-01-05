package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler godoc
// @summary Get Prometheus metrics.
// @description Get Prometheus metrics.
// @description Wrapper for Prometheus-supplied handler.
// @description Served on port defined by METRICS_PORT environment variable.
// @tags metrics
// @produce text/plain
// @success 200
// @router /metrics [get]
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
