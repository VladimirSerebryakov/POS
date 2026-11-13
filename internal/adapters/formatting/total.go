package formatting

import "math"

func RoundTotal(total float64) float64 {
	scaled := math.Nextafter(total*100, math.Copysign(math.Inf(1), total))
	return math.Round(scaled) / 100
}
