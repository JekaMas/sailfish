package sailfish

import "math/bits"

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
	// bits.Len64 gives the binary magnitude. Multiplication by 1233/4096
	// estimates the corresponding decimal power with at most one threshold
	// correction. bits.Sub64 exposes that correction as a borrow bit, so the
	// generated arm64 path uses CLZ and SUBS/NGC instead of a data-dependent
	// comparison tree. OR 1 maps zero to the one-digit magnitude without a
	// separate zero branch.
	nonzero := v | 1
	estimate := bits.Len64(nonzero) * 1233 >> 12
	_, borrow := bits.Sub64(nonzero, powersOf10Uint64[estimate], 0)
	return estimate + 1 - int(borrow)
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
