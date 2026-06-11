package Display

import (
	"math"
)

type FrameTemperatureStatistics struct {
	MinTemp  uint16
	MaxTemp  uint16
	MeanTemp uint16
	StdDev   float64
}

func FrameStatsRLE(rle []Pair) (out FrameTemperatureStatistics) {
	minIdx := 0
	maxIdx := 0
	mean := 0.0
	length := 0.0

	var f []float64

	for i, current_pair := range rle {
		if rle[minIdx].Idx > current_pair.Idx {
			minIdx = i
		}
		if rle[maxIdx].Idx < current_pair.Idx {
			maxIdx = i
		}
		mean += float64(current_pair.Idx) * float64(current_pair.Length)
		length += float64(current_pair.Length)

		for range current_pair.Length {
			f = append(f, float64(current_pair.Idx)*0.01)
		}
	}

	mean /= length
	var varianceSum float64
	for _, current_pair := range rle {
		diff := float64(current_pair.Idx) - mean
		varianceSum += (diff * diff) * float64(current_pair.Length)
	}

	variance := varianceSum / length
	stddev := math.Sqrt(variance)

	out = FrameTemperatureStatistics{
		MinTemp:  rle[minIdx].Idx,
		MaxTemp:  rle[maxIdx].Idx,
		MeanTemp: uint16(mean),
		StdDev:   stddev,
	}
	return
}
