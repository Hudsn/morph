package morph

import "math"

const float_equality_tolerance = 1e-9

func isFloatEqual(a float64, b float64) bool {
	return math.Abs(a-b) <= float_equality_tolerance
}
