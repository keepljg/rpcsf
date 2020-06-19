package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var once sync.Once

type Prom struct {
	timer   *prometheus.HistogramVec
	counter *prometheus.CounterVec
	state   *prometheus.GaugeVec
}

func NewProm() *Prom {
	once.Do(func() {
		prometheus.MustRegister(prometheus.NewGoCollector()) // 开启go程序metric
	})
	return &Prom{}
}

func (p *Prom) RegisterTimer(name string, help string, labels []string) *Prom {
	if p == nil || p.timer != nil {
		return p
	}
	p.timer = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(p.timer)
	return p
}

func (p *Prom) RegisterCounter(name string, help string, labels []string) *Prom {
	if p == nil || p.counter != nil {
		return p
	}
	p.counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(p.counter)
	return p
}

func (p *Prom) RegisterState(name string, help string, labels []string) *Prom {
	if p == nil || p.counter != nil {
		return p
	}
	p.state = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)
	prometheus.MustRegister(p.state)
	return p
}

func (p *Prom) Timing(name string, time float64, ext ...string) {
	if p.timer != nil {
		p.timer.WithLabelValues(append([]string{name}, ext...)...).Observe(time)
	}
}

func (p *Prom) Incr(name string, ext ...string) {
	if p.counter != nil {
		p.counter.WithLabelValues(append([]string{name}, ext...)...).Inc()
	}
}

func (p *Prom) State(name string, v float64, ext ...string) {
	if p.state != nil {
		p.state.WithLabelValues(append([]string{name}, ext...)...).Set(v)
	}
}
