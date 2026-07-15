package sailfish

import "github.com/holiman/uint256"

const maxUnitDigits = 78

// fillUnitDigits writes an unscaled unit without allocation. Cross-format
// comparison uses it to avoid potentially overflowing rescaling multiplication.
func fillUnitDigits[U Unit](dst *[maxUnitDigits]byte, units U) int {
	switch value := any(units).(type) {
	case uint8:
		digits := decimalDigits64(uint64(value))
		fillUnsigned64(dst[:digits], uint64(value))
		return digits
	case uint16:
		digits := decimalDigits64(uint64(value))
		fillUnsigned64(dst[:digits], uint64(value))
		return digits
	case uint32:
		digits := decimalDigits64(uint64(value))
		fillUnsigned64(dst[:digits], uint64(value))
		return digits
	case uint64:
		digits := decimalDigits64(value)
		fillUnsigned64(dst[:digits], value)
		return digits
	case uint256.Int:
		chunks, count, digits := splitUint256Decimal(value)
		fillUnsigned256(dst[:digits], chunks, count, digits)
		return digits
	default:
		return 0
	}
}
