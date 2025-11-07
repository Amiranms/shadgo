//go:build !solution

package speller

import (
	"strings"
)

var fromZeroToTwenty = map[int64]string{0: " zero", 1: " one", 2: " two",
	3: " three", 4: " four", 5: " five",
	6: " six", 7: " seven", 8: " eight",
	9: " nine", 10: " ten", 11: " eleven",
	12: " twelve", 13: " thirteen", 14: " fourteen",
	15: " fifteen", 16: " sixteen", 17: " seventeen",
	18: " eighteen", 19: " nineteen"}

var TenMultiples = []string{"", "", " twenty", " thirty",
	" forty", " fifty", " sixty", " seventy", " eighty",
	" ninety"}

var hundredConstant string = " hundred"

var digits = []string{"", " thousand ", " million ", " billion "}

func absInt(x int64) int64 {
	return absDiffInt(x, 0)
}

func absDiffInt(x, y int64) int64 {
	if x < y {
		return y - x
	}
	return x - y
}

func Spell(n int64) string {
	resultSpell := ""
	minus := ""
	if n < 0 {
		minus = "minus "
	}
	n = absInt(n)

	if smallValue, ok := fromZeroToTwenty[n]; ok {
		return minus + strings.TrimSpace(smallValue)
	}

	for digitIterator, i := n%1000, 0; n > 0; digitIterator, i = n%1000, i+1 {
		if digitIterator == 0 && n > 0 {
			n /= 1000
			continue
		}
		hundreds := digitIterator / 100
		tens := digitIterator % 100
		resultSpell = digits[i] + resultSpell
		if tens > 0 {

			if strVal, ok := fromZeroToTwenty[tens]; ok {
				resultSpell = strVal + resultSpell
			} else {

				if tens%10 != 0 {
					resultSpell = "-" + strings.TrimSpace(fromZeroToTwenty[tens%10]) + resultSpell
				}
				resultSpell = TenMultiples[tens/10] + resultSpell
			}
		}
		if hundreds > 0 {
			resultSpell = strings.TrimSpace(fromZeroToTwenty[hundreds]) + hundredConstant + resultSpell
		} else {
			resultSpell = strings.TrimSpace(resultSpell)
		}
		n /= 1000

	}
	resultSpell = minus + strings.TrimSpace(resultSpell)
	return resultSpell
}
