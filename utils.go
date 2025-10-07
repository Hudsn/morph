package morph

import (
	"fmt"
	"math"
	"strconv"
)

const float_equality_tolerance = 1e-9

func isFloatEqual(a float64, b float64) bool {
	return math.Abs(a-b) <= float_equality_tolerance
}

func lineColString(line int, col int) string {
	return fmt.Sprintf("%d:%d", line, col)
}

func lineAndCol(input []rune, targetIdx int) (int, int) {
	line := 1
	col := 1
	for _, r := range input[:targetIdx] {
		switch r {
		case '\n': // reset if newline
			line++
			col = 1
		default:
			col++
		}
	}
	return line, col
}

func anyToInt64(original interface{}) (int64, error) {
	switch v := original.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unable to convert input argument. unsupported type")
	}
}
