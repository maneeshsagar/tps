package currency

import (
	"fmt"
	"strconv"
	"strings"
)

// RupeesToPaise converts "100.50" to 10050
func RupeesToPaise(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid format")
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid format")
	}
	paise := whole * 100

	if len(parts) == 2 {
		dec := parts[1]
		if len(dec) == 1 {
			dec += "0"
		}
		p, err := strconv.ParseInt(dec, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid format")
		}
		if whole >= 0 && !strings.HasPrefix(s, "-") {
			paise += p
		} else {
			paise -= p
		}
	}
	return paise, nil
}

// PaiseToRupees converts 10050 to "100.50"
func PaiseToRupees(p int64) string {
	r := p / 100
	rem := p % 100
	if rem < 0 {
		rem = -rem
	}
	return fmt.Sprintf("%d.%02d", r, rem)
}
