package broadcast

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

type MemoryBenchmark struct {
	Benchmark             *testing.B
	MaxMemoryPerOperation float64
	MinimumOperationCount int

	before uint64
	after  uint64
}

func NewMemoryBenchmark(b *testing.B) *MemoryBenchmark {
	memBenchmark := MemoryBenchmark{Benchmark: b, MaxMemoryPerOperation: 1, MinimumOperationCount: 1}
	memBenchmark.Start()
	return &memBenchmark
}

func (benchmark *MemoryBenchmark) Start() {
	benchmark.before = benchmark.CurrentResidentMemory()
}

func (benchmark *MemoryBenchmark) IsEnabled() bool {
	return benchmark.OperationCount() > benchmark.MinimumOperationCount
}

func (benchmark *MemoryBenchmark) OperationCount() int {
	return benchmark.Benchmark.N
}

func (benchmark *MemoryBenchmark) MemoryUsage() uint64 {
	return benchmark.after - benchmark.before
}

func (benchmark *MemoryBenchmark) MemoryPerOperation() float64 {
	return float64(benchmark.MemoryUsage()) / float64(benchmark.OperationCount())
}

func (benchmark *MemoryBenchmark) CurrentResidentMemory() uint64 {
	runtime.GC()

	memStatus, _ := ioutil.ReadFile(fmt.Sprintf("/proc/%d/statm", os.Getpid()))
	residentStringValue := strings.Split(string(memStatus), " ")[1]
	residentIntValue, _ := strconv.ParseUint(residentStringValue, 10, 64)
	return residentIntValue
}

func (benchmark *MemoryBenchmark) Status() string {
	return fmt.Sprintf("%dkB -> %dkB delta: %dKB for %dops %fKB/op", benchmark.before, benchmark.after, benchmark.MemoryUsage(), benchmark.OperationCount(), benchmark.MemoryPerOperation())
}

func (benchmark *MemoryBenchmark) Complete() {
	if benchmark.IsEnabled() {
		benchmark.after = benchmark.CurrentResidentMemory()

		status := benchmark.Status()

		if benchmark.MemoryPerOperation() > benchmark.MaxMemoryPerOperation {
			benchmark.Benchmark.Fatalf(fmt.Sprintf("Memory usage per operation exceeds limit (%fkB/op) : %s", benchmark.MaxMemoryPerOperation, status))
		} else {
			Log.Debugf("Mem stats: %s", status)
		}
	}
}
