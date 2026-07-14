package sailfish

const digitPairs = "0001020304050607080910111213141516171819" +
	"2021222324252627282930313233343536373839" +
	"4041424344454647484950515253545556575859" +
	"6061626364656667686970717273747576777879" +
	"8081828384858687888990919293949596979899"

var powersOf10Uint64 = [...]uint64{ //nolint:gochecknoglobals // immutable lookup table
	1,
	10,
	100,
	1_000,
	10_000,
	100_000,
	1_000_000,
	10_000_000,
	100_000_000,
	1_000_000_000,
	10_000_000_000,
	100_000_000_000,
	1_000_000_000_000,
	10_000_000_000_000,
	100_000_000_000_000,
	1_000_000_000_000_000,
	10_000_000_000_000_000,
	100_000_000_000_000_000,
	1_000_000_000_000_000_000,
	10_000_000_000_000_000_000,
}

func growBy(dst []byte, n int) ([]byte, []byte) {
	start := len(dst)
	if cap(dst)-start >= n {
		dst = dst[:start+n]
	} else {
		dst = append(dst, make([]byte, n)...)
	}
	return dst, dst[start:]
}

func decimalDigits64(v uint64) int {
	if v < 1e10 {
		if v < 1e5 {
			if v < 1e2 {
				if v < 10 {
					return 1
				}
				return 2
			}
			if v < 1e3 {
				return 3
			}
			if v < 1e4 {
				return 4
			}
			return 5
		}
		if v < 1e7 {
			if v < 1e6 {
				return 6
			}
			return 7
		}
		if v < 1e8 {
			return 8
		}
		if v < 1e9 {
			return 9
		}
		return 10
	}
	if v < 1e15 {
		if v < 1e12 {
			if v < 1e11 {
				return 11
			}
			return 12
		}
		if v < 1e13 {
			return 13
		}
		if v < 1e14 {
			return 14
		}
		return 15
	}
	if v < 1e17 {
		if v < 1e16 {
			return 16
		}
		return 17
	}
	if v < 1e18 {
		return 18
	}
	if v < 1e19 {
		return 19
	}
	return 20
}

func fillUnsigned64(dst []byte, value uint64) {
	i := len(dst)
	for value >= 100 {
		q := value / 100
		r := int(value - q*100)
		i -= 2
		dst[i] = digitPairs[r*2]
		dst[i+1] = digitPairs[r*2+1]
		value = q
	}

	r := int(value) * 2
	i--
	dst[i] = digitPairs[r+1]
	if value >= 10 {
		i--
		dst[i] = digitPairs[r]
	}
}

func fillFixed64(dst []byte, value uint64) {
	i := len(dst)
	for i >= 2 {
		q := value / 100
		r := int(value - q*100)
		i -= 2
		dst[i] = digitPairs[r*2]
		dst[i+1] = digitPairs[r*2+1]
		value = q
	}
	if i == 1 {
		dst[0] = byte(value) + '0'
	}
}

func scaledTextLen(digits, scale int) int {
	if scale == 0 {
		return digits
	}
	if digits > scale {
		return digits + 1
	}
	return scale + 2
}
