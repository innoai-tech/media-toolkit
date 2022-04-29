package index

func max[T int64 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func min[T int64 | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}
