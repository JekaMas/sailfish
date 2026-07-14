package sailfish

import (
	"math/bits"

	"github.com/holiman/uint256"
)

const (
	maxUint256Scale   = 77
	maxUint256TextLen = 79
	uint256ChunkBase  = uint64(10_000_000_000_000_000_000) // 1e19
	maxUint256Chunks  = 5
)

// Uint256Units is a zero-sized unit provider. Embed it in a venue type.
type Uint256Units struct{}

func (Uint256Units) unitParseString(s string, scale int) (uint256.Int, bool, Error) {
	return parseUint256(s, scale)
}

func (Uint256Units) unitParseBytes(b []byte, scale int) (uint256.Int, bool, Error) {
	return parseUint256(b, scale)
}

func (Uint256Units) unitAppend(dst []byte, units uint256.Int, scale int) []byte {
	return appendUint256Decimal(dst, units, scale)
}

func (Uint256Units) unitString(units uint256.Int, scale int) string {
	var buf [maxUint256TextLen]byte
	out := appendUint256Decimal(buf[:0], units, scale)
	return string(out)
}

func (Uint256Units) unitLen(units uint256.Int, scale int) int {
	if units[1]|units[2]|units[3] == 0 {
		return scaledTextLen(decimalDigits64(units[0]), scale)
	}
	_, _, digits := splitUint256Decimal(units)
	return scaledTextLen(digits, scale)
}

func (Uint256Units) unitCompare(a, b uint256.Int) int {
	for i := 3; i >= 0; i-- {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func parseUint256[S decimalInput](s S, scale int) (uint256.Int, bool, Error) {
	if len(s) == 0 {
		return uint256.Int{}, false, ErrSyntax
	}

	if scale == 0 {
		value, err := parseUint256Digits(s, 0, len(s))
		if err != "" {
			return uint256.Int{}, false, err
		}
		return value, len(s) == 1 || s[0] != '0', ""
	}

	dot := len(s) - scale - 1
	if dot >= 1 && s[dot] == '.' {
		if len(s)-1 <= 19 {
			value, err := parseUint64WithDot(s, dot)
			if err != "" {
				return uint256.Int{}, false, err
			}
			return uint256.Int{value}, dot == 1 || s[0] != '0', ""
		}
		value, err := parseUint256WithDot(s, dot)
		if err != "" {
			return uint256.Int{}, false, err
		}
		return value, dot == 1 || s[0] != '0', ""
	}

	return parseUint256General(s, scale)
}

func parseUint256Digits[S decimalInput](s S, begin, end int) (uint256.Int, Error) {
	firstDigits := (end - begin) % 19
	if firstDigits == 0 {
		firstDigits = 19
	}

	var chunk uint64
	var err Error
	if firstDigits >= 8 {
		chunk, err = parseUint64DenseChunk(s, begin, begin+firstDigits)
	} else {
		chunk, err = parseUint64Chunk(s, begin, begin+firstDigits)
	}
	if err != "" {
		return uint256.Int{}, err
	}
	value := uint256.Int{chunk}
	begin += firstDigits

	for begin < end {
		chunk, err = parseUint64DenseChunk(s, begin, begin+19)
		if err != "" {
			return uint256.Int{}, err
		}
		var overflow bool
		value, overflow = uint256MulSmallAdd(value, uint256ChunkBase, chunk)
		if overflow {
			return uint256.Int{}, ErrRange
		}
		begin += 19
	}
	return value, ""
}

func parseUint256WithDot[S decimalInput](s S, dot int) (uint256.Int, Error) {
	totalDigits := len(s) - 1
	firstDigits := totalDigits % 19
	if firstDigits == 0 {
		firstDigits = 19
	}

	position := 0
	chunk, next, err := parseUint64ChunkWithDot(s, position, dot, firstDigits)
	if err != "" {
		return uint256.Int{}, err
	}
	value := uint256.Int{chunk}
	position = next
	consumed := firstDigits

	for consumed < totalDigits {
		chunk, next, err = parseUint64ChunkWithDot(s, position, dot, 19)
		if err != "" {
			return uint256.Int{}, err
		}
		var overflow bool
		value, overflow = uint256MulSmallAdd(value, uint256ChunkBase, chunk)
		if overflow {
			return uint256.Int{}, ErrRange
		}
		position = next
		consumed += 19
	}
	return value, ""
}

func parseUint64ChunkWithDot[S decimalInput](
	s S,
	position, dot, digits int,
) (uint64, int, Error) {
	if position == dot {
		position++
	}

	if position > dot || position+digits <= dot {
		var value uint64
		var err Error
		if digits >= 8 {
			value, err = parseUint64DenseChunk(s, position, position+digits)
		} else {
			value, err = parseUint64Chunk(s, position, position+digits)
		}
		return value, position + digits, err
	}

	leftDigits := dot - position
	left, err := parseUint64Chunk(s, position, dot)
	if err != "" {
		return 0, position, err
	}
	rightDigits := digits - leftDigits
	right, err := parseUint64Chunk(s, dot+1, dot+1+rightDigits)
	if err != "" {
		return 0, position, err
	}
	return left*powersOf10Uint64[rightDigits] + right,
		dot + 1 + rightDigits, ""
}

func appendUint256Digits[S decimalInput](
	value uint256.Int,
	s S,
	begin, end int,
) (uint256.Int, Error) {
	for begin < end {
		digits := min(end-begin, 19)
		var chunk uint64
		var err Error
		if digits >= 8 {
			chunk, err = parseUint64DenseChunk(s, begin, begin+digits)
		} else {
			chunk, err = parseUint64Chunk(s, begin, begin+digits)
		}
		if err != "" {
			return uint256.Int{}, err
		}
		var overflow bool
		value, overflow = uint256MulSmallAdd(value, powersOf10Uint64[digits], chunk)
		if overflow {
			return uint256.Int{}, ErrRange
		}
		begin += digits
	}
	return value, ""
}

func parseUint256General[S decimalInput](s S, scale int) (uint256.Int, bool, Error) {
	dot := -1
	fractionDigits := 0
	integerDigits := 0

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if dot >= 0 || integerDigits == 0 || i == len(s)-1 {
				return uint256.Int{}, false, ErrSyntax
			}
			dot = i
			continue
		}

		digit := c - '0'
		if digit > 9 {
			return uint256.Int{}, false, ErrSyntax
		}
		if dot >= 0 {
			if fractionDigits == scale {
				return uint256.Int{}, false, ErrPrecision
			}
			fractionDigits++
		} else {
			integerDigits++
		}
	}

	if integerDigits == 0 {
		return uint256.Int{}, false, ErrSyntax
	}

	integerEnd := len(s)
	if dot >= 0 {
		integerEnd = dot
	}
	value, err := parseUint256Digits(s, 0, integerEnd)
	if err != "" {
		return uint256.Int{}, false, err
	}
	if dot >= 0 {
		value, err = appendUint256Digits(value, s, dot+1, len(s))
		if err != "" {
			return uint256.Int{}, false, err
		}
	}

	canonicalFraction := dot >= 0 && fractionDigits == scale
	missing := scale - fractionDigits
	for missing >= 19 {
		var overflow bool
		value, overflow = uint256MulSmallAdd(value, uint256ChunkBase, 0)
		if overflow {
			return uint256.Int{}, false, ErrRange
		}
		missing -= 19
	}
	if missing != 0 {
		var overflow bool
		value, overflow = uint256MulSmallAdd(value, powersOf10Uint64[missing], 0)
		if overflow {
			return uint256.Int{}, false, ErrRange
		}
	}

	canonicalInteger := integerDigits == 1 || s[0] != '0'
	return value, canonicalInteger && canonicalFraction, ""
}

// uint256MulSmallAdd computes value*multiplier+add. multiplier and add are
// decimal chunks smaller than 1e19, which bounds each propagated carry to one
// uint64 limb.
func uint256MulSmallAdd(
	value uint256.Int,
	multiplier uint64,
	add uint64,
) (uint256.Int, bool) {
	var result uint256.Int

	hi, lo := bits.Mul64(value[0], multiplier)
	var addCarry uint64
	result[0], addCarry = bits.Add64(lo, add, 0)
	carry := hi + addCarry

	hi, lo = bits.Mul64(value[1], multiplier)
	result[1], addCarry = bits.Add64(lo, carry, 0)
	carry = hi + addCarry

	hi, lo = bits.Mul64(value[2], multiplier)
	result[2], addCarry = bits.Add64(lo, carry, 0)
	carry = hi + addCarry

	hi, lo = bits.Mul64(value[3], multiplier)
	result[3], addCarry = bits.Add64(lo, carry, 0)
	return result, hi != 0 || addCarry != 0
}

func appendUint256Decimal(dst []byte, units uint256.Int, scale int) []byte {
	if units[1]|units[2]|units[3] == 0 {
		return appendUint64Decimal(dst, units[0], scale)
	}
	chunks, chunkCount, digits := splitUint256Decimal(units)
	n := scaledTextLen(digits, scale)
	dst, out := growBy(dst, n)

	switch {
	case scale == 0:
		fillUnsigned256(out, chunks, chunkCount, digits)
	case digits > scale:
		fillUnsigned256(out[1:], chunks, chunkCount, digits)
		integerDigits := digits - scale
		copy(out[:integerDigits], out[1:integerDigits+1])
		out[integerDigits] = '.'
	default:
		out[0] = '0'
		out[1] = '.'
		for i := 2; i < len(out)-digits; i++ {
			out[i] = '0'
		}
		fillUnsigned256(out[len(out)-digits:], chunks, chunkCount, digits)
	}
	return dst
}

func splitUint256Decimal(value uint256.Int) ([maxUint256Chunks]uint64, int, int) {
	var chunks [maxUint256Chunks]uint64
	if value[0]|value[1]|value[2]|value[3] == 0 {
		return chunks, 1, 1
	}

	count := 0
	for value[0]|value[1]|value[2]|value[3] != 0 {
		var remainder uint64
		value, remainder = uint256DivMod64(value, uint256ChunkBase)
		chunks[count] = remainder
		count++
	}
	digits := (count-1)*19 + decimalDigits64(chunks[count-1])
	return chunks, count, digits
}

func uint256DivMod64(value uint256.Int, divisor uint64) (uint256.Int, uint64) {
	var quotient uint256.Int
	var remainder uint64

	// Skip known-zero high limbs. The quotient shrinks on every decimal-chunk
	// iteration, avoiding unnecessary hardware divisions.
	switch {
	case value[3] != 0:
		quotient[3], remainder = bits.Div64(0, value[3], divisor)
		quotient[2], remainder = bits.Div64(remainder, value[2], divisor)
		quotient[1], remainder = bits.Div64(remainder, value[1], divisor)
		quotient[0], remainder = bits.Div64(remainder, value[0], divisor)
	case value[2] != 0:
		quotient[2], remainder = bits.Div64(0, value[2], divisor)
		quotient[1], remainder = bits.Div64(remainder, value[1], divisor)
		quotient[0], remainder = bits.Div64(remainder, value[0], divisor)
	case value[1] != 0:
		quotient[1], remainder = bits.Div64(0, value[1], divisor)
		quotient[0], remainder = bits.Div64(remainder, value[0], divisor)
	default:
		quotient[0] = value[0] / divisor
		remainder = value[0] - quotient[0]*divisor
	}
	return quotient, remainder
}

func fillUnsigned256(
	dst []byte,
	chunks [maxUint256Chunks]uint64,
	chunkCount int,
	digits int,
) {
	topDigits := digits - (chunkCount-1)*19
	fillUnsigned64(dst[:topDigits], chunks[chunkCount-1])
	pos := topDigits
	for i := chunkCount - 2; i >= 0; i-- {
		fillFixed19(dst[pos:pos+19], chunks[i])
		pos += 19
	}
}
