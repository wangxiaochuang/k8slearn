package integer

func IntMax(a, b int) int {
	if b > a {
		return b
	}
	return a
}

func IntMin(a, b int) int {
	if b < a {
		return b
	}
	return a
}

func Int32Max(a, b int32) int32 {
	if b > a {
		return b
	}
	return a
}

func Int32Min(a, b int32) int32 {
	if b < a {
		return b
	}
	return a
}

func Int64Max(a, b int64) int64 {
	if b > a {
		return b
	}
	return a
}

func Int64Min(a, b int64) int64 {
	if b < a {
		return b
	}
	return a
}

func RoundToInt32(a float64) int32 {
	if a < 0 {
		return int32(a - 0.5)
	}
	return int32(a + 0.5)
}
