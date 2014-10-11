package broadcast

import (
	"flag"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/librato"
	"io"
	"strings"
	"time"
)

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
	Librato MetricsLibratoConfig
}

func (config *MetricsConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.Librato.Flags(flags, strings.Join([]string{prefix, "librato"}, "-"))
}

func (config *MetricsConfig) Apply() {
	config.Librato.Apply()
}

type MetricsLibratoConfig struct {
	Account string
	Token   string
}

func (config *MetricsLibratoConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Account, strings.Join([]string{prefix, "account"}, "-"), "", "The Librato account")
	flags.StringVar(&config.Token, strings.Join([]string{prefix, "token"}, "-"), "", "The Librato token")
}

func (config *MetricsLibratoConfig) Apply() {
	go librato.Librato(metrics.DefaultRegistry,
		10e9,
		config.Account,
		config.Token,
		"gobroadcast",
		[]float64{0.95},
		time.Millisecond,
	)
}
