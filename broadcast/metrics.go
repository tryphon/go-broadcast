package broadcast

import (
	"flag"
	metrics "github.com/tryphon/go-metrics"
	"github.com/tryphon/go-metrics/librato"
	"io"
	"strings"
	"time"
)

type LocalMetrics struct {
	prefix string
}

func (local *LocalMetrics) Name(name string) string {
	if local.prefix != "" {
		return strings.Join([]string{local.prefix, name}, ".")
	} else {
		return name
	}
}

func (local *LocalMetrics) Counter(name string) metrics.Counter {
	return metrics.GetOrRegisterCounter(local.Name(name), nil)
}

func (local *LocalMetrics) Gauge(name string) metrics.Gauge {
	return metrics.GetOrRegisterGauge(local.Name(name), nil)
}

func (local *LocalMetrics) Histogram(name string) metrics.Histogram {
	return metrics.GetOrRegisterHistogram(local.Name(name), nil, metrics.NewExpDecaySample(1028, 0.015))
}

type MetricsReadCloser struct {
	reader  io.ReadCloser
	counter metrics.Counter
}

func NewMetricsReadCloser(reader io.ReadCloser, counterName string) *MetricsReadCloser {
	counter := metrics.GetOrRegisterCounter(counterName, nil)
	return &MetricsReadCloser{reader: reader, counter: counter}
}

func (metricsReadCloser *MetricsReadCloser) Read(buffer []byte) (n int, err error) {
	n, err = metricsReadCloser.reader.Read(buffer)
	if err == nil && metricsReadCloser.counter != nil {
		metricsReadCloser.counter.Inc(int64(n))
	}
	return
}

func (metricsReadCloser *MetricsReadCloser) Close() error {
	return metricsReadCloser.reader.Close()
}

type MetricsConfig struct {
	Librato MetricsLibratoConfig `json:",omitempty"`
}

func (config *MetricsConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.Librato.Flags(flags, strings.Join([]string{prefix, "librato"}, "-"))
}

func (config *MetricsConfig) Apply() {
	config.Librato.Apply()
}

type MetricsLibratoConfig struct {
	Account string `json:",omitempty"`
	Token   string `json:",omitempty"`
}

func (config *MetricsLibratoConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Account, strings.Join([]string{prefix, "account"}, "-"), "", "The Librato account")
	flags.StringVar(&config.Token, strings.Join([]string{prefix, "token"}, "-"), "", "The Librato token")
}

func (config *MetricsLibratoConfig) Apply() {
	if config.Account != "" && config.Token != "" {
		go librato.Librato(metrics.DefaultRegistry,
			10e9,
			config.Account,
			config.Token,
			"gobroadcast",
			[]float64{0.95},
			time.Millisecond,
		)
	}
}
