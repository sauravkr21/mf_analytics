package analytics

import "sort"

// ----------------------
// Rolling Returns
// ----------------------
func Returns(navs []float64, window int) []float64 {
	var r []float64

	for i := 0; i+window < len(navs); i++ {
		if navs[i] == 0 {
			continue // avoid division by zero
		}
		val := (navs[i+window] - navs[i]) / navs[i] * 100
		r = append(r, val)
	}

	return r
}

// ----------------------
// Statistics
// ----------------------
func Stats(arr []float64) (min, max, median, p25, p75 float64) {
	if len(arr) == 0 {
		return
	}

	sort.Float64s(arr)

	min = arr[0]
	max = arr[len(arr)-1]

	n := len(arr)

	// Median
	if n%2 == 0 {
		median = (arr[n/2-1] + arr[n/2]) / 2
	} else {
		median = arr[n/2]
	}

	// Percentiles (simple index-based)
	p25 = arr[n/4]
	p75 = arr[(3*n)/4]

	return
}

// ----------------------
// Maximum Drawdown
// ----------------------
func MaxDrawdown(navs []float64) float64 {
	if len(navs) == 0 {
		return 0
	}

	peak := navs[0]
	maxDD := 0.0

	for _, v := range navs {
		if v > peak {
			peak = v
		}
		dd := (v - peak) / peak * 100
		if dd < maxDD {
			maxDD = dd
		}
	}

	return maxDD
}