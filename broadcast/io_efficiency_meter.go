package broadcast

import (
	"math"
	"time"
)

type IoEfficiencyMeter struct {
	inputCounter  int64
	outputCounter int64
	timeMark      int64

	timeWindowDuration time.Duration
	history            *IoEfficiencyMeterHistory

	Metrics  *LocalMetrics
	EventLog *LocalEventLog
}

func (meter *IoEfficiencyMeter) metrics() *LocalMetrics {
	if meter.Metrics == nil {
		meter.Metrics = &LocalMetrics{}
	}
	return meter.Metrics
}

func (meter *IoEfficiencyMeter) eventLog() *LocalEventLog {
	if meter.EventLog == nil {
		meter.EventLog = &LocalEventLog{}
	}
	return meter.EventLog
}

func (meter *IoEfficiencyMeter) checkTimeWindow() {
	if meter.timeWindowDuration == 0 {
		meter.timeWindowDuration = 10 * time.Second
	}

	timeMark := time.Now().Unix() / int64(meter.timeWindowDuration.Seconds())

	if meter.timeMark == 0 || meter.timeMark != timeMark {
		meter.timeMark = timeMark

		savedEfficiency := meter.Efficiency()
		meter.Reset()

		if savedEfficiency <= 0.9 {
			meter.eventLog().NewEvent("Bad network performance")
		}

		meter.History().Push(savedEfficiency)

		savedEfficiencyInPercentage := int64(savedEfficiency * 100)
		meter.metrics().Gauge("Efficiency").Update(savedEfficiencyInPercentage)
		meter.metrics().Histogram("EfficiencyHistory").Update(savedEfficiencyInPercentage)
	}
}

func (meter *IoEfficiencyMeter) History() *IoEfficiencyMeterHistory {
	if meter.history == nil {
		meter.history = NewIoEfficiencyMeterHistory(30)
	}
	return meter.history
}

func (meter *IoEfficiencyMeter) Reset() {
	meter.inputCounter = 0
	meter.outputCounter = 0
}

func (meter *IoEfficiencyMeter) Input(count int64) {
	meter.checkTimeWindow()
	meter.inputCounter += count
}

func (meter *IoEfficiencyMeter) Output(count int64) {
	meter.checkTimeWindow()
	meter.outputCounter += count
}

func (meter *IoEfficiencyMeter) Efficiency() float64 {
	if meter.outputCounter >= meter.inputCounter {
		return 1
	}

	return float64(meter.outputCounter) / float64(meter.inputCounter)
}

type IoEfficiencyMeterHistory struct {
	Efficiencies []float64
	Max          float64
	Min          float64
	Average      float64

	size int
}

func NewIoEfficiencyMeterHistory(size int) *IoEfficiencyMeterHistory {
	return &IoEfficiencyMeterHistory{
		Efficiencies: make([]float64, 0),
		Min:          math.NaN(),
		Max:          math.NaN(),
		Average:      math.NaN(),
		size:         size,
	}
}

func (history *IoEfficiencyMeterHistory) IsEmpty() bool {
	return len(history.Efficiencies) == 0
}

func (history *IoEfficiencyMeterHistory) Push(efficiency float64) {
	tail := history.Efficiencies
	if len(history.Efficiencies) >= history.size {
		tail = history.Efficiencies[1:]
	}

	history.Efficiencies = append(tail, efficiency)

	var sum float64
	for _, value := range history.Efficiencies {
		sum += value
		if math.IsNaN(history.Max) || value > history.Max {
			history.Max = value
		}
		if math.IsNaN(history.Min) || value < history.Min {
			history.Min = value
		}
	}
	history.Average = sum / float64(len(history.Efficiencies))
}
