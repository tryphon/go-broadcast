package broadcast

import (
	"bytes"
	metrics "github.com/rcrowley/go-metrics"
	"io/ioutil"
	"testing"
)

func TestMetricsReadCloser_Read(t *testing.T) {
	metricsReadCloser := NewMetricsReadCloser(ioutil.NopCloser(bytes.NewBufferString("dummy")), "test")
	read, _ := metricsReadCloser.Read(make([]byte, 1024))
	if metrics.GetOrRegisterCounter("test", nil).Count() != int64(read) {
		t.Errorf("Counter should count read bytes :\n got: %v\nwant: %v", metrics.GetOrRegisterCounter("test", nil).Count(), int64(read))
	}
}
