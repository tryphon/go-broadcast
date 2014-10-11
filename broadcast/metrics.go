package broadcast

import (
	metrics "github.com/tryphon/go-metrics"
	"io"
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
