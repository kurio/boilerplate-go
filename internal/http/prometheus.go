/*
Based on:
https://github.com/labstack/echo-contrib/blob/master/prometheus/prometheus.go

Example:

```

	package main

	import (
		handler "github.com/kurio/boilerplate-go/internal/http"
	)

	func main() {
		e := echo.New()

		// Enable metrics middleware
		p := handler.NewPrometheus("goboilerplate", nil)
		p.Use(e)

		e.Logger.Fatal(e.Start(":1323"))
	}

```
*/
package http

import (
	"net/http"
	"strconv"
	"time"

	echoProm "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var defaultMetricPath = "/metrics"

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
)

// reqDurBuckets is the buckets for request duration. Here, we use the prometheus defaults
var reqDurBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// reqDurMSBuckets is the buckets for request duration in milliseconds.
var reqDurMSBuckets = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// reqSzBuckets is the buckets for request size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var reqSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// resSzBuckets is the buckets for response size. Here we define a spectrom from 1KB thru 1NB up to 10MB.
var resSzBuckets = []float64{1.0 * KB, 2.0 * KB, 5.0 * KB, 10.0 * KB, 100 * KB, 500 * KB, 1.0 * MB, 2.5 * MB, 5.0 * MB, 10.0 * MB}

// Customized metrics, but based on the standard default metrics
var reqCnt = &echoProm.Metric{
	ID:          "reqCnt",
	Name:        "requests_total",
	Description: "How many HTTP requests processed.",
	Type:        "counter_vec",
	Args:        []string{"code", "operation"}}

var reqDur = &echoProm.Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Type:        "histogram_vec",
	Args:        []string{"code", "operation"},
	Buckets:     reqDurBuckets}

var reqDurMS = &echoProm.Metric{
	ID:          "reqDurMS",
	Name:        "request_duration_ms",
	Description: "The HTTP request latencies in milliseconds.",
	Type:        "histogram_vec",
	Args:        []string{"code", "operation"},
	Buckets:     reqDurMSBuckets}

var resSz = &echoProm.Metric{
	ID:          "resSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Type:        "histogram_vec",
	Args:        []string{"code", "operation"},
	Buckets:     resSzBuckets}

var reqSz = &echoProm.Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Type:        "histogram_vec",
	Args:        []string{"code", "operation"},
	Buckets:     reqSzBuckets}

var defaultMetrics = []*echoProm.Metric{
	reqCnt,
	reqDur,
	reqDurMS,
	resSz,
	reqSz,
}

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt                         *prometheus.CounterVec
	reqDur, reqDurMS, reqSz, resSz *prometheus.HistogramVec

	MetricsList []*echoProm.Metric
	MetricsPath string
	ServiceName string
	Subsystem   string
	Skipper     middleware.Skipper

	OperationLabelMappingFunc echoProm.RequestCounterLabelMappingFunc
}

// NewPrometheus generates a new set of metrics with a certain service name
func NewPrometheus(serviceName string, skipper middleware.Skipper) *Prometheus {
	if skipper == nil {
		skipper = URLSkipper
	}

	p := &Prometheus{
		MetricsList: defaultMetrics,
		MetricsPath: defaultMetricPath,
		ServiceName: serviceName,
		Subsystem:   "http",
		Skipper:     skipper,
		OperationLabelMappingFunc: func(c echo.Context) string {
			operation := "UNKNOWN"
			for _, r := range c.Echo().Routes() {
				if r.Method == c.Request().Method && r.Path == c.Path() {
					operation = r.Name
				}
			}

			return operation
		},
	}

	p.registerMetrics()
	return p
}

// NewMetric associates prometheus.Collector based on Metric.Type
func NewMetric(m *echoProm.Metric, serviceName, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem:   subsystem,
				Name:        m.Name,
				Help:        m.Description,
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			m.Args,
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem:   subsystem,
				Name:        m.Name,
				Help:        m.Description,
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			m.Args,
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem:   subsystem,
				Name:        m.Name,
				Help:        m.Description,
				Buckets:     m.Buckets,
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			m.Args,
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem:   subsystem,
				Name:        m.Name,
				Help:        m.Description,
				Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001}, // Based on http://alexandrutopliceanu.ro/post/targeted-quantiles/
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			m.Args,
		)
	}
	return metric
}

func (p *Prometheus) registerMetrics() {
	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, p.ServiceName, p.Subsystem)
		if err := prometheus.Register(metric); err != nil {
			log.Errorf("%s could not be registered in Prometheus: %v", metricDef.Name, err)
		}
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case reqDurMS:
			p.reqDurMS = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(*prometheus.HistogramVec)
		case reqSz:
			p.reqSz = metric.(*prometheus.HistogramVec)
		}
		metricDef.MetricCollector = metric
	}
}

// Use adds the middleware to the Echo engine.
func (p *Prometheus) Use(e *echo.Echo) {
	e.Use(p.HandlerFunc)
	e.GET(p.MetricsPath, prometheusHandler())
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if c.Path() == p.MetricsPath {
			return next(c)
		}
		if p.Skipper(c) {
			return next(c)
		}

		start := time.Now()
		reqSz := float64(computeApproximateRequestSize(c.Request()))

		err = next(c)
		if err != nil {
			return
		}

		status := strconv.Itoa(c.Response().Status)
		resSz := float64(c.Response().Size)
		elapsed := time.Since(start).Seconds()

		operation := p.OperationLabelMappingFunc(c)

		p.reqCnt.WithLabelValues(status, operation).Inc()

		p.reqDur.WithLabelValues(status, operation).Observe(elapsed)
		p.reqDurMS.WithLabelValues(status, operation).Observe(elapsed * 1000)
		p.reqSz.WithLabelValues(status, operation).Observe(reqSz)
		p.resSz.WithLabelValues(status, operation).Observe(resSz)

		return
	}
}

func prometheusHandler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
