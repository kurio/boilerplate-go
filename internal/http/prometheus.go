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

// Customized metrics, but based on the standard default metrics
var reqCnt = &echoProm.Metric{
	ID:          "reqCnt",
	Name:        "requests_total",
	Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	Type:        "counter_vec",
	Args:        []string{"service", "code", "method", "url"}}

var reqDur = &echoProm.Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Type:        "counter_vec",
	Args:        []string{"service", "code", "method", "url"}}

var reqDurHis = &echoProm.Metric{
	ID:          "reqDurHis",
	Name:        "request_duration_histogram_seconds",
	Description: "The Histogram for HTTP request latencies in seconds.",
	Type:        "histogram_duration_vec",
	Args:        []string{"service", "code", "operation"}}

var resSz = &echoProm.Metric{
	ID:          "resSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Type:        "counter_vec",
	Args:        []string{"service", "code", "method", "url"}}

var resSzHis = &echoProm.Metric{
	ID:          "resSzHis",
	Name:        "response_size_histogram_bytes",
	Description: "The Histogram for HTTP response sizes in bytes.",
	Type:        "histogram_size_vec",
	Args:        []string{"service", "code", "operation"}}

var reqSz = &echoProm.Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Type:        "counter_vec",
	Args:        []string{"service", "code", "method", "url"}}

var reqSzHis = &echoProm.Metric{
	ID:          "reqSzHis",
	Name:        "request_size_histogram_bytes",
	Description: "The Histogram for HTTP request sizes in bytes.",
	Type:        "histogram_size_vec",
	Args:        []string{"service", "code", "operation"}}

var defaultMetrics = []*echoProm.Metric{
	reqCnt,
	reqDur,
	reqDurHis,
	resSz,
	resSzHis,
	reqSz,
	reqSzHis,
}

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt, reqDur, reqSz, resSz  *prometheus.CounterVec
	reqDurHis, resSzHis, reqSzHis *prometheus.HistogramVec

	MetricsList []*echoProm.Metric
	MetricsPath string
	Subsystem   string
	ServiceName string
	Skipper     middleware.Skipper

	RequestCounterURLLabelMappingFunc echoProm.RequestCounterLabelMappingFunc

	// Context string to use as a prometheus URL label
	URLLabelFromContext string
}

// NewPrometheus generates a new set of metrics with a certain service name
func NewPrometheus(serviceName string, skipper middleware.Skipper) *Prometheus {
	if skipper == nil {
		skipper = URLSkipper
	}

	p := &Prometheus{
		MetricsList: defaultMetrics,
		MetricsPath: defaultMetricPath,
		Subsystem:   "http",
		Skipper:     skipper,
		ServiceName: serviceName,
		RequestCounterURLLabelMappingFunc: func(c echo.Context) string {
			return c.Path() // by default do nothing, i.e. return URL as is
		},
	}

	p.registerMetrics()
	return p
}

// NewMetric associates prometheus.Collector based on Metric.Type
func NewMetric(m *echoProm.Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram_duration_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   prometheus.ExponentialBuckets(0.05, 1.75, 11), // Minimum of 0,05s to 13,5s. Based on https://github.com/joao-fontenele/express-prometheus-middleware/blob/633cda57b64af00cecdc6a5ce9698a4b10c18f91/src/index.js#L23s
			},
			m.Args,
		)
	case "histogram_size_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   prometheus.ExponentialBuckets(50, 5, 8), // Minimum 50B to 3,9MB
			},
			m.Args,
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem:  subsystem,
				Name:       m.Name,
				Help:       m.Description,
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001}, // Based on http://alexandrutopliceanu.ro/post/targeted-quantiles/
			},
			m.Args,
		)
	}
	return metric
}

func (p *Prometheus) registerMetrics() {
	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, p.Subsystem)
		if err := prometheus.Register(metric); err != nil {
			log.Errorf("%s could not be registered in Prometheus: %v", metricDef.Name, err)
		}
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.reqDur = metric.(*prometheus.CounterVec)
		case resSz:
			p.resSz = metric.(*prometheus.CounterVec)
		case reqSz:
			p.reqSz = metric.(*prometheus.CounterVec)
		case reqDurHis:
			p.reqDurHis = metric.(*prometheus.HistogramVec)
		case resSzHis:
			p.resSzHis = metric.(*prometheus.HistogramVec)
		case reqSzHis:
			p.reqSzHis = metric.(*prometheus.HistogramVec)
		}
		metricDef.MetricCollector = metric
	}
}

func (p *Prometheus) operationMappingFunc(c echo.Context) string {
	operation := "UNKNOWN"
	for _, r := range c.Echo().Routes() {
		if r.Method == c.Request().Method && r.Path == c.Path() {
			operation = r.Name
		}
	}

	return operation
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

		if err = next(c); err != nil {
			return
		}

		status := strconv.Itoa(c.Response().Status)
		resSz := float64(c.Response().Size)
		elapsed := time.Since(start).Seconds()

		method := c.Request().Method
		url := p.RequestCounterURLLabelMappingFunc(c)
		operation := p.operationMappingFunc(c)

		if len(p.URLLabelFromContext) > 0 {
			u := c.Get(p.URLLabelFromContext)
			if u == nil {
				u = "unknown"
			}
			url = u.(string)
		}

		p.reqCnt.WithLabelValues(p.ServiceName, status, method, url).Inc()

		p.reqDurHis.WithLabelValues(p.ServiceName, status, operation).Observe(elapsed)
		p.reqSzHis.WithLabelValues(p.ServiceName, status, operation).Observe(reqSz)
		p.resSzHis.WithLabelValues(p.ServiceName, status, operation).Observe(resSz)

		p.reqDur.WithLabelValues(p.ServiceName, status, method, url).Add(elapsed)
		p.reqSz.WithLabelValues(p.ServiceName, status, method, url).Add(reqSz)
		p.resSz.WithLabelValues(p.ServiceName, status, method, url).Add(resSz)

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
