package collector

// roundTo1 rounds a float64 to 1 decimal place.
func roundTo1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
