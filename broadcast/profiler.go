package broadcast

import (
	"flag"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
)

type ProfilerConfig struct {
	CPU    string
	Memory string
}

func (config *ProfilerConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.CPU, strings.Join([]string{prefix, "cpu"}, "-"), "", "Enable CPU profiler with specified output file")
	flags.StringVar(&config.Memory, strings.Join([]string{prefix, "memory"}, "-"), "", "Enable memory profiler with specified output file")
}

func (config *ProfilerConfig) Apply() {
	if config.CPU != "" {
		controller := ProfilerController{
			FileName: config.CPU,
			Profiler: &CPUProfiler{},
		}
		controller.Start()
	}
	if config.Memory != "" {
		controller := ProfilerController{
			FileName: config.Memory,
			Profiler: &MemoryProfiler{},
		}
		controller.Start()
	}
}

type ProfilerController struct {
	FileName string
	Profiler Profiler
}

func (controller *ProfilerController) Start() error {
	file, err := os.Create(controller.FileName)
	if err != nil {
		return err
	}

	controller.Profiler.Start(file)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			Log.Debugf("Receive interrupt signal: %v", sig)
			controller.Profiler.Stop(file)
			os.Exit(0)
		}
	}()

	return nil
}

type Profiler interface {
	Start(file *os.File)
	Stop(file *os.File)
}

type CPUProfiler struct {
}

func (profiler *CPUProfiler) Start(file *os.File) {
	pprof.StartCPUProfile(file)
}

func (profiler *CPUProfiler) Stop(file *os.File) {
	pprof.StopCPUProfile()
}

type MemoryProfiler struct {
}

func (profiler *MemoryProfiler) Start(file *os.File) {

}

func (profiler *MemoryProfiler) Stop(file *os.File) {
	pprof.WriteHeapProfile(file)
}
