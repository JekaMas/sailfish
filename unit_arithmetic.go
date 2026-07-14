package sailfish

import (
	"math/bits"

	"github.com/holiman/uint256"
)

// Arithmetic dispatches once on the closed Unit set. Keeping it outside the
// venue method dictionary lowers uint256 mutation overhead on current Go
// compilers.
func addUnits[U Unit](a, b U) (U, bool) {
	switch left := any(a).(type) {
	case uint8:
		right := any(b).(uint8)
		result := left + right
		return any(result).(U), result < left
	case uint16:
		right := any(b).(uint16)
		result := left + right
		return any(result).(U), result < left
	case uint32:
		right := any(b).(uint32)
		result := left + right
		return any(result).(U), result < left
	case uint64:
		right := any(b).(uint64)
		result, carry := bits.Add64(left, right, 0)
		return any(result).(U), carry != 0
	case uint256.Int:
		right := any(b).(uint256.Int)
		var result uint256.Int
		var carry uint64
		result[0], carry = bits.Add64(left[0], right[0], 0)
		result[1], carry = bits.Add64(left[1], right[1], carry)
		result[2], carry = bits.Add64(left[2], right[2], carry)
		result[3], carry = bits.Add64(left[3], right[3], carry)
		return any(result).(U), carry != 0
	default:
		panic("sailfish: unreachable unit type")
	}
}

func subUnits[U Unit](a, b U) (U, bool) {
	switch left := any(a).(type) {
	case uint8:
		right := any(b).(uint8)
		return any(left - right).(U), left < right
	case uint16:
		right := any(b).(uint16)
		return any(left - right).(U), left < right
	case uint32:
		right := any(b).(uint32)
		return any(left - right).(U), left < right
	case uint64:
		right := any(b).(uint64)
		result, borrow := bits.Sub64(left, right, 0)
		return any(result).(U), borrow != 0
	case uint256.Int:
		right := any(b).(uint256.Int)
		var result uint256.Int
		var borrow uint64
		result[0], borrow = bits.Sub64(left[0], right[0], 0)
		result[1], borrow = bits.Sub64(left[1], right[1], borrow)
		result[2], borrow = bits.Sub64(left[2], right[2], borrow)
		result[3], borrow = bits.Sub64(left[3], right[3], borrow)
		return any(result).(U), borrow != 0
	default:
		panic("sailfish: unreachable unit type")
	}
}
