package functions

import (
	"strconv"
	"unicode"
)

func LuhnCheck(number string) bool {
	var sum int
	alt := false

	for i := len(number) - 1; i >= 0; i-- {
		r := rune(number[i])
		if !unicode.IsDigit(r) {
			continue
		}

		d, _ := strconv.Atoi(string(r))
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}
