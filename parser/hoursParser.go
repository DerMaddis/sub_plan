package parser

import (
	"strconv"
	"strings"
)

func parseHours(s string) (int, int, error) {
	switch len(s) {
	case 1:
		res, err := strconv.Atoi(s)
		if err != nil {
			return 0, 0, parseError
		}
		return res, 0, nil
	default:
		split := strings.Split(s, " - ")
		if len(split) != 2 {
			return 0, 0, parseError
		}

		first, err := strconv.Atoi(split[0])
		if err != nil {
			return 0, 0, parseError
		}

		second, err := strconv.Atoi(split[1])
		if err != nil {
			return 0, 0, parseError
		}

		return first, second, nil
	}
}
