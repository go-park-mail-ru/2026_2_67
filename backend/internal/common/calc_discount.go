package common

func CalcPriceWithDiscount(price int, discount uint) int {
	return roundHalfEven(price*(100-int(discount)), 100)
}

func roundHalfEven(num, den int) int {
	if den <= 0 {
		panic("roundHalfEven: den must be positive")
	}
	q := num / den
	r := num % den

	if r < 0 {
		r += den
		q--
	}

	switch diff := 2*r - den; {
	case diff < 0:
		return q
	case diff > 0:
		return q + 1
	default:
		if q%2 == 0 {
			return q
		}
		return q + 1
	}
}
