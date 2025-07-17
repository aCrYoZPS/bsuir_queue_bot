package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseArray(input string) ([]int8, error) {
	parts := strings.Split(input, ",")
	result := make([]int8, 0, len(parts))

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		val, err := strconv.ParseInt(trimmed, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid int8 value %q: %w", trimmed, err)
		}
		result = append(result, int8(val))
	}

	return result, nil
}

func ArrayToString(arr []int8) string {
	var stringValues []string
	for _, item := range arr {
		stringValues = append(stringValues, strconv.Itoa(int(item)))
	}
	return strings.Join(stringValues, ",")
}
