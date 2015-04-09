package broadcast

import (
	"math"
	"testing"
)

func TestMath_dBFSToPeak(t *testing.T) {
	var conditions = []struct {
		dbFSValue float64
		peakValue float64
	}{
		{0, 1},
		{-3, 0.5},
		{-6, 0.25},
		{-30, 0.001},
	}

	for _, condition := range conditions {
		if math.Abs(dBFSToPeak(condition.dbFSValue)-condition.peakValue) > condition.peakValue/100 {
			t.Errorf("Wrong peak value for %f :\n got: %v\nwant: %v", condition.dbFSValue, dBFSToPeak(condition.dbFSValue), condition.peakValue)
		}
	}
}
