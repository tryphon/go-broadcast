package broadcast

import "math"

func dBFSToPeak(dbValue float64) float64 {
	return math.Exp(dbValue * math.Log(10) / 10)
}
