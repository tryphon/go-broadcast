package broadcast

import (
	"math"
	"reflect"
	"testing"
)

func TestIoEfficiencyMeter_Efficiency(t *testing.T) {
	var conditions = []struct {
		input      int64
		output     int64
		efficiency float64
	}{
		{100, 0, 0},
		{100, 100, 1},
		{0, 100, 1},
		{100, 50, 0.5},
	}

	for _, condition := range conditions {
		meter := IoEfficiencyMeter{}

		meter.Input(condition.input)
		meter.Output(condition.output)

		if meter.Efficiency() != condition.efficiency {
			t.Errorf("Wrong efficiency :\n got: %v\nwant: %v", meter.Efficiency(), 0)
		}
	}
}

func TestIoEfficiencyMeter_TimeWindow(t *testing.T) {
	meter := IoEfficiencyMeter{}
	meter.Input(100)

	meter.timeMark = 100
	meter.Input(100)
	if meter.inputCounter != 100 {
		t.Errorf("InputCounter should have been reset :\n got: %v\nwant: %v", meter.inputCounter, 100)
	}
}

func TestIoEfficiencyMeterHistory_Push(t *testing.T) {
	history := NewIoEfficiencyMeterHistory(10)

	history.Push(1)

	if !reflect.DeepEqual(history.Efficiencies, []float64{1}) {
		t.Errorf("Efficiency should be added in history Efficiencies :\n got: %v\nwant: %v", history.Efficiencies, []float64{1})
	}
	if history.Max != 1 {
		t.Errorf("Max should be updated :\n got: %v\nwant: %v", history.Max, 1)
	}
	if history.Average != 1 {
		t.Errorf("Average should be updated :\n got: %v\nwant: %v", history.Average, 1)
	}

	for i := 1; i <= history.size*2; i++ {
		value := 1.0 - float64(i%10)/10.0
		t.Log(value)
		history.Push(value)
	}

	t.Log(history.Efficiencies)

	if len(history.Efficiencies) != history.size {
		t.Errorf("Efficiencies should not contain more than 'size' elements :\n got: %v\nwant: %v", len(history.Efficiencies), history.size)
	}
	if history.Average != 0.55 {
		t.Errorf("Wrong Average :\n got: %v\nwant: %v", history.Average, 0.55)
	}
	if math.Abs(history.Min-0.1) >= 0.01 {
		t.Errorf("Wrong Min :\n got: %v\nwant: %v", history.Min, 0.1)
	}
}
