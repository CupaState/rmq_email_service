package metrics

import (
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics
type Metrics interface {
	IncHits(status int, method, path string)
	ObserveResponseTime(status int, method, path string, observeTime float64)
}

type PrometheusMetrics struct {
	HitsTotal prometheus.Counter
	Hits 			*prometheus.CounterVec
	Times 		*prometheus.HistogramVec
}

func CreateMetrics(address string, name string) (Metrics, error) {
	var metr PrometheusMetrics
	e_msg := "prometheus.Register"
	metr.HitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_hits_total",
	})

	if err := prometheus.Register(metr.HitsTotal); err != nil {
		log.Fatalf("%s: %s", e_msg, err)
		return nil, err
	}

	metr.Hits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name + "_hits",
		},
		[]string{"status", "method", "path"},
	)

	if err := prometheus.Register(metr.Hits); err != nil {
		log.Fatalf("%s: %s", e_msg, err)
		return nil, err
	}

	metr.Times = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name + "_times",
		},
		[]string{"status", "method", "path"},
	)

	if err := prometheus.Register(metr.Times); err != nil {
		log.Fatalf("%s: %s", e_msg, err)
		return nil, err
	}

	if err := prometheus.Register(prometheus.NewBuildInfoCollector()); err != nil {
		log.Fatalf("%s: %s", e_msg, err)
		return nil, err
	}

	return &metr, nil
}
 
func (p *PrometheusMetrics) IncHits(status int, method, path string) {
	p.HitsTotal.Inc()
	p.Hits.WithLabelValues(strconv.Itoa(status), method, path).Inc()
}

func (p *PrometheusMetrics) ObserveResponseTime(status int, method, path string, observeTime float64) {
	p.Times.WithLabelValues(strconv.Itoa(status), method, path).Observe(observeTime)
}
