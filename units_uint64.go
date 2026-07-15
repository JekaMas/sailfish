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

	// Eight- and sixteen-digit canonical values map exactly to one or two
	// validated SWAR words after removing the known decimal point. Keep this
	// shape path here rather than behind a helper: on short exchange values, a
	// call costs a material fraction of the entire parse. Runtime dot shifts
	// retain one implementation for every scale; generated per-scale masks were
	// faster in isolation but rejected because they duplicate code by both scale
	// and token length.
	digits := len(s) - 1
	if digits == 8 || digits == 16 {
		dot := digits - scale
		if dot >= 1 && dot < len(s)-1 && s[dot] == '.' {
			first := loadEightBytes(s, 0)
			if digits == 8 {
				packed := removeKnownDot(first, s[8], dot)
				value, ok := parseEightDigits(packed)
				if !ok {
					return 0, false, ErrSyntax
				}
				return uint64(value), dot == 1 || s[0] != '0', ""
			}

			var second uint64
			switch {
			case dot < 8:
				first = removeKnownDot(first, s[8], dot)
				second = loadEightBytes(s, 9)
			case dot == 8:
				second = loadEightBytes(s, 9)
			default:
				second = removeKnownDot(loadEightBytes(s, 8), s[16], dot-8)
			}
			high, ok := parseEightDigits(first)
			if !ok {
				return 0, false, ErrSyntax
			}
			low, ok := parseEightDigits(second)
			if !ok {
				return 0, false, ErrSyntax
			}
			return uint64(high)*100_000_000 + uint64(low), dot == 1 || s[0] != '0', ""
		}
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
	digits := len(s) - 1
	if digits == 8 || digits == 16 {
		return parseUint64KnownDot(s, dot, digits)
	}

	// Below 20 total numeric digits, both pieces and recombination fit uint64.
	if digits >= 20 {
		return parseUint64WithDotOverflow(s, dot)
	}

	// Keep both pairwise segment parsers in this function. The generic chunk
	// helper is not inlined by current Go compilers; avoiding its two calls is
	// material for short CEX prices and amounts.
	var value uint64
	begin := 0
	if dot&1 != 0 {
		digit := s[0] - '0'
		if digit > 9 {
			return 0, ErrSyntax
		}
		value = uint64(digit)
		begin = 1
	}
	for ; begin < dot; begin += 2 {
		a := s[begin] - '0'
		b := s[begin+1] - '0'
		if a > 9 || b > 9 {
			return 0, ErrSyntax
		}
		value = value*100 + uint64(a)*10 + uint64(b)
	}

	fractionBegin := dot + 1
	fractionDigits := len(s) - fractionBegin
	if fractionDigits&1 != 0 {
		digit := s[fractionBegin] - '0'
		if digit > 9 {
			return 0, ErrSyntax
		}
		value = value*10 + uint64(digit)
		fractionBegin++
	}
	for ; fractionBegin < len(s); fractionBegin += 2 {
		a := s[fractionBegin] - '0'
		b := s[fractionBegin+1] - '0'
		if a > 9 || b > 9 {
			return 0, ErrSyntax
		}
		value = value*100 + uint64(a)*10 + uint64(b)
	}
	return value, ""
}

// parseUint64KnownDot removes the already-located decimal point while packing
// exactly 8 or 16 digits into SWAR words. This validates every digit once and
// replaces the serial pairwise multiply/add chain used for irregular lengths.
// The enclosing canonical parser proves the point location from the scale;
// this helper still checks the byte so direct internal use fails closed.
func parseUint64KnownDot[S decimalInput](s S, dot, digits int) (uint64, Error) {
	if dot < 1 || dot >= len(s)-1 || s[dot] != '.' {
		return 0, ErrSyntax
	}

	first := loadEightBytes(s, 0)
	if digits == 8 {
		packed := removeKnownDot(first, s[8], dot)
		value, ok := parseEightDigits(packed)
		if !ok {
			return 0, ErrSyntax
		}
		return uint64(value), ""
	}

	var second uint64
	switch {
	case dot < 8:
		first = removeKnownDot(first, s[8], dot)
		second = loadEightBytes(s, 9)
	case dot == 8:
		second = loadEightBytes(s, 9)
	default:
		second = removeKnownDot(loadEightBytes(s, 8), s[16], dot-8)
	}
	high, ok := parseEightDigits(first)
	if !ok {
		return 0, ErrSyntax
	}
	low, ok := parseEightDigits(second)
	if !ok {
		return 0, ErrSyntax
	}
	return uint64(high)*100_000_000 + uint64(low), ""
}

// removeKnownDot compacts an eight-byte window containing one point and
// appends the next digit. Bytes before the point stay in place; bytes after it
// shift down by one byte. The result is a contiguous little-endian SWAR word.
func removeKnownDot(raw uint64, last byte, dot int) uint64 {
	lowerMask := uint64(1)<<(dot*8) - 1
	lower := raw & lowerMask
	upper := raw >> ((dot + 1) * 8)
	return lower | upper<<(dot*8) | uint64(last)<<56
}

func parseUint64WithDotOverflow[S decimalInput](s S, dot int) (uint64, Error) {
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
		power := powersOf10Uint64[scale]
		integer := units / power
		fraction := units - integer*power
		integerDigits := digits - scale
		fillUnsigned64(out[:integerDigits], integer)
		out[integerDigits] = '.'
		fillFixed64(out[integerDigits+1:], fraction)
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
