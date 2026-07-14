package sailfish

import "math"

const (
	maxUint64Scale   = 19
	maxUint64TextLen = 21
)

// Uint64Units is a zero-sized unit provider. Embed it in a venue type.
type Uint64Units struct{}

func (Uint64Units) unitParseString(s string, scale int) (uint64, bool, Error) {
	return parseUint64(s, scale)
}

func (Uint64Units) unitParseBytes(b []byte, scale int) (uint64, bool, Error) {
	return parseUint64(b, scale)
}

func (Uint64Units) unitAppend(dst []byte, units uint64, scale int) []byte {
	return appendUint64Decimal(dst, units, scale)
}

func (Uint64Units) unitString(units uint64, scale int) string {
	var buf [maxUint64TextLen]byte
	out := appendUint64Decimal(buf[:0], units, scale)
	return string(out)
}

func (Uint64Units) unitLen(units uint64, scale int) int {
	return scaledTextLen(decimalDigits64(units), scale)
}

func (Uint64Units) unitCompare(a, b uint64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func parseUint64[S decimalInput](s S, scale int) (uint64, bool, Error) {
	if len(s) == 0 {
		return 0, false, ErrSyntax
	}

	if scale == 0 {
		value, err := parseUint64Digits(s, 0, len(s))
		if err != "" {
			return 0, false, err
		}
		return value, len(s) == 1 || s[0] != '0', ""
	}

	// Canonical fixed-scale wire values dominate exchange payloads. Infer the
	// point location from the scale and avoid a search/state machine.
	dot := len(s) - scale - 1
	if dot >= 1 && s[dot] == '.' {
		value, err := parseUint64WithDot(s, dot)
		if err != "" {
			return 0, false, err
		}
		return value, dot == 1 || s[0] != '0', ""
	}

	return parseUint64General(s, scale)
}

func parseUint64Digits[S decimalInput](s S, begin, end int) (uint64, Error) {
	// Every unsigned decimal with at most 19 digits fits in uint64. Parse those
	// inputs in pairs without an overflow division in the inner loop.
	if end-begin < 20 {
		return parseUint64Chunk(s, begin, end)
	}

	var value uint64
	for i := begin; i < end; i++ {
		digit := s[i] - '0'
		if digit > 9 {
			return 0, ErrSyntax
		}
		if value > (math.MaxUint64-uint64(digit))/10 {
			return 0, ErrRange
		}
		value = value*10 + uint64(digit)
	}
	return value, ""
}

func parseUint64WithDot[S decimalInput](s S, dot int) (uint64, Error) {
	// Below 20 total numeric digits, both pieces and recombination fit uint64.
	if len(s)-1 < 20 {
		integer, err := parseUint64Chunk(s, 0, dot)
		if err != "" {
			return 0, err
		}
		fraction, err := parseUint64Chunk(s, dot+1, len(s))
		if err != "" {
			return 0, err
		}
		return integer*powersOf10Uint64[len(s)-dot-1] + fraction, ""
	}

	var value uint64
	for i := 0; i < len(s); i++ {
		if i == dot {
			continue
		}
		digit := s[i] - '0'
		if digit > 9 {
			return 0, ErrSyntax
		}
		if value > (math.MaxUint64-uint64(digit))/10 {
			return 0, ErrRange
		}
		value = value*10 + uint64(digit)
	}
	return value, ""
}

func parseUint64General[S decimalInput](s S, scale int) (uint64, bool, Error) {
	var value uint64
	seenDot := false
	fractionDigits := 0
	integerDigits := 0

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if seenDot || integerDigits == 0 || i == len(s)-1 {
				return 0, false, ErrSyntax
			}
			seenDot = true
			continue
		}

		digit := c - '0'
		if digit > 9 {
			return 0, false, ErrSyntax
		}
		if seenDot {
			if fractionDigits == scale {
				return 0, false, ErrPrecision
			}
			fractionDigits++
		} else {
			integerDigits++
		}

		if value > (math.MaxUint64-uint64(digit))/10 {
			return 0, false, ErrRange
		}
		value = value*10 + uint64(digit)
	}

	if integerDigits == 0 {
		return 0, false, ErrSyntax
	}
	canonicalFraction := seenDot && fractionDigits == scale
	for fractionDigits < scale {
		if value > math.MaxUint64/10 {
			return 0, false, ErrRange
		}
		value *= 10
		fractionDigits++
	}

	canonicalInteger := integerDigits == 1 || s[0] != '0'
	return value, canonicalInteger && canonicalFraction, ""
}

func appendUint64Decimal(dst []byte, units uint64, scale int) []byte {
	digits := decimalDigits64(units)
	n := scaledTextLen(digits, scale)
	dst, out := growBy(dst, n)

	switch {
	case scale == 0:
		fillUnsigned64(out, units)
	case digits > scale:
		// Fill into out[1:], shift only the integer prefix left, then place '.'.
		fillUnsigned64(out[1:], units)
		integerDigits := digits - scale
		copy(out[:integerDigits], out[1:integerDigits+1])
		out[integerDigits] = '.'
	default:
		out[0] = '0'
		out[1] = '.'
		for i := 2; i < len(out)-digits; i++ {
			out[i] = '0'
		}
		fillUnsigned64(out[len(out)-digits:], units)
	}
	return dst
}
