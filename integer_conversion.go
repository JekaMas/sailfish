package sailfish

import (
	"math"
	"math/big"

	"github.com/holiman/uint256"
)

// FromBigInt constructs a fixed decimal from non-negative, already-scaled
// integer units. It reads source without retaining or modifying it.
func (FixedDecimalCodec[V, U]) FromBigInt(source *big.Int) (FixedDecimal[V, U], error) {
	var zero U
	if source == nil {
		return FixedDecimal[V, U]{}, boxedErrNilSource
	}
	if source.Sign() < 0 {
		return FixedDecimal[V, U]{}, boxedErrRange
	}

	switch any(zero).(type) {
	case uint256.Int:
		// SetFromBig copies at most four machine words directly. Keeping this
		// case ahead of the native path avoids an unnecessary IsUint64 scan for
		// the wide backend.
		var value uint256.Int
		if value.SetFromBig(source) {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(value).(U)}, nil
	case uint8:
		if !source.IsUint64() || source.Uint64() > math.MaxUint8 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint8(source.Uint64())).(U)}, nil
	case uint16:
		if !source.IsUint64() || source.Uint64() > math.MaxUint16 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint16(source.Uint64())).(U)}, nil
	case uint32:
		if !source.IsUint64() || source.Uint64() > math.MaxUint32 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint32(source.Uint64())).(U)}, nil
	case uint64:
		if !source.IsUint64() {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(source.Uint64()).(U)}, nil
	default:
		return FixedDecimal[V, U]{}, boxedErrRange
	}
}

// FromU256 constructs a fixed decimal from non-negative, already-scaled
// uint256 units. Narrow backends reject values outside their integer width.
func (FixedDecimalCodec[V, U]) FromU256(source uint256.Int) (FixedDecimal[V, U], error) {
	var zero U
	switch any(zero).(type) {
	case uint8:
		if source[1]|source[2]|source[3] != 0 || source[0] > math.MaxUint8 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint8(source[0])).(U)}, nil
	case uint16:
		if source[1]|source[2]|source[3] != 0 || source[0] > math.MaxUint16 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint16(source[0])).(U)}, nil
	case uint32:
		if source[1]|source[2]|source[3] != 0 || source[0] > math.MaxUint32 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint32(source[0])).(U)}, nil
	case uint64:
		if source[1]|source[2]|source[3] != 0 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(source[0]).(U)}, nil
	case uint256.Int:
		return FixedDecimal[V, U]{units: any(source).(U)}, nil
	default:
		return FixedDecimal[V, U]{}, boxedErrRange
	}
}

// ToBigInt writes the already-scaled integer units into caller-owned dst.
// Reusing a destination with sufficient capacity avoids allocation. A fresh
// destination may allocate its math/big backing words once.
func (d FixedDecimal[V, U]) ToBigInt(dst *big.Int) error {
	if dst == nil {
		return boxedErrNilDestination
	}
	switch value := any(d.units).(type) {
	case uint8:
		dst.SetUint64(uint64(value))
	case uint16:
		dst.SetUint64(uint64(value))
	case uint32:
		dst.SetUint64(uint64(value))
	case uint64:
		dst.SetUint64(value)
	case uint256.Int:
		value.IntoBig(&dst)
	}
	return nil
}

// ToU256 returns the already-scaled integer units as one inline four-limb
// value. Every supported backend fits exactly, so conversion cannot fail.
func (d FixedDecimal[V, U]) ToU256() uint256.Int {
	return unitsToU256(d.units)
}

func unitsToU256[U Unit](source U) uint256.Int {
	switch value := any(source).(type) {
	case uint8:
		return uint256.Int{uint64(value)}
	case uint16:
		return uint256.Int{uint64(value)}
	case uint32:
		return uint256.Int{uint64(value)}
	case uint64:
		return uint256.Int{value}
	case uint256.Int:
		return value
	default:
		return uint256.Int{}
	}
}
