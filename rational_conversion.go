package sailfish

import (
	"math"
	"math/big"
	"math/bits"

	"github.com/holiman/uint256"
)

// BigRatWorkspace owns temporary integer words used by ToBigRat. Its zero
// value is ready for use. Reuse one workspace and destination per goroutine;
// neither type is safe for concurrent mutation.
type BigRatWorkspace struct {
	numerator   big.Int
	denominator big.Int
}

// FromBigRat constructs a fixed decimal only when source has an exact
// representation at the format's fractional decimal places. It performs no
// rounding and does not retain or modify source.
func (c FixedDecimalCodec[V, U]) FromBigRat(source *big.Rat) (FixedDecimal[V, U], error) {
	if source == nil {
		return FixedDecimal[V, U]{}, boxedErrNilSource
	}
	if source.Sign() < 0 {
		return FixedDecimal[V, U]{}, boxedErrRange
	}
	// Match the other codec hot paths: NewFixedDecimalCodec validates once and
	// caches the scale, while a zero-value codec derives it from V.
	scale := c.fractionalDecimalPlaces()

	var zero U
	switch any(zero).(type) {
	case uint8, uint16, uint32, uint64:
		return c.fromBigRatNative(source, scale)
	case uint256.Int:
		return c.fromBigRatUint256(source, scale)
	default:
		return FixedDecimal[V, U]{}, boxedErrRange
	}
}

func (c FixedDecimalCodec[V, U]) fromBigRatNative(
	source *big.Rat,
	scale int,
) (FixedDecimal[V, U], error) {
	numerator := source.Num()
	denominator := source.Denom()
	if !numerator.IsUint64() {
		return FixedDecimal[V, U]{}, boxedErrRange
	}
	if !denominator.IsUint64() {
		return FixedDecimal[V, U]{}, boxedErrPrecision
	}
	den := denominator.Uint64()
	power := powersOf10Uint64[scale]
	if power%den != 0 {
		return FixedDecimal[V, U]{}, boxedErrPrecision
	}
	hi, units := bits.Mul64(numerator.Uint64(), power/den)
	if hi != 0 {
		return FixedDecimal[V, U]{}, boxedErrRange
	}
	var zero U
	switch any(zero).(type) {
	case uint8:
		if units > math.MaxUint8 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint8(units)).(U)}, nil
	case uint16:
		if units > math.MaxUint16 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint16(units)).(U)}, nil
	case uint32:
		if units > math.MaxUint32 {
			return FixedDecimal[V, U]{}, boxedErrRange
		}
		return FixedDecimal[V, U]{units: any(uint32(units)).(U)}, nil
	case uint64:
		return FixedDecimal[V, U]{units: any(units).(U)}, nil
	default:
		return FixedDecimal[V, U]{}, boxedErrRange
	}
}

func (c FixedDecimalCodec[V, U]) fromBigRatUint256(
	source *big.Rat,
	scale int,
) (FixedDecimal[V, U], error) {
	var numerator, denominator uint256.Int
	if numerator.SetFromBig(source.Num()) {
		return FixedDecimal[V, U]{}, boxedErrRange
	}
	if denominator.SetFromBig(source.Denom()) {
		return FixedDecimal[V, U]{}, boxedErrPrecision
	}

	var factor uint256.Int
	if denominator[1]|denominator[2]|denominator[3] == 0 {
		var remainder uint64
		factor, remainder = uint256DivMod64(powersOf10Uint256[scale], denominator[0])
		if remainder != 0 {
			return FixedDecimal[V, U]{}, boxedErrPrecision
		}
	} else {
		var remainder uint256.Int
		factor.DivMod(&powersOf10Uint256[scale], &denominator, &remainder)
		if !remainder.IsZero() {
			return FixedDecimal[V, U]{}, boxedErrPrecision
		}
	}

	var units uint256.Int
	if _, overflow := units.MulOverflow(&numerator, &factor); overflow {
		return FixedDecimal[V, U]{}, boxedErrRange
	}
	return FixedDecimal[V, U]{units: any(units).(U)}, nil
}

// ToBigRat writes the exact decimal value into caller-owned dst. Reusing dst
// and workspace avoids steady-state allocation. The first wide call may grow
// their math/big backing words once.
func (d FixedDecimal[V, U]) ToBigRat(dst *big.Rat, workspace *BigRatWorkspace) error {
	if dst == nil {
		return boxedErrNilDestination
	}
	if workspace == nil {
		return boxedErrNilWorkspace
	}
	scale, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return boxedError(err)
	}
	if d.setIntegralBigRat(dst, workspace, scale) {
		return nil
	}
	if err := d.ToBigInt(&workspace.numerator); err != nil {
		return err
	}
	if scale <= maxUint64Scale {
		workspace.denominator.SetUint64(powersOf10Uint64[scale])
	} else {
		denominator := &workspace.denominator
		powersOf10Uint256[scale].IntoBig(&denominator)
	}
	dst.SetFrac(&workspace.numerator, &workspace.denominator)
	return nil
}

// setIntegralBigRat bypasses math/big's rational GCD normalization when the
// fixed decimal is a whole integer. This is a common balance/quantity case and
// turns the caller-owned path into a zero-allocation integer assignment.
func (d FixedDecimal[V, U]) setIntegralBigRat(
	dst *big.Rat,
	workspace *BigRatWorkspace,
	scale int,
) bool {
	if scale <= maxUint64Scale {
		power := powersOf10Uint64[scale]
		switch units := any(d.units).(type) {
		case uint8:
			return setIntegralNativeBigRat(dst, uint64(units), power, scale)
		case uint16:
			return setIntegralNativeBigRat(dst, uint64(units), power, scale)
		case uint32:
			return setIntegralNativeBigRat(dst, uint64(units), power, scale)
		case uint64:
			return setIntegralNativeBigRat(dst, units, power, scale)
		}
	}

	units, ok := any(d.units).(uint256.Int)
	if !ok {
		return false
	}
	if scale > 0 && (units[0]&1 != 0 || (units[0]%5+units[1]%5+units[2]%5+units[3]%5)%5 != 0) {
		return false
	}
	var quotient, remainder uint256.Int
	if scale <= maxUint64Scale {
		var remainder64 uint64
		quotient, remainder64 = uint256DivMod64(units, powersOf10Uint64[scale])
		if remainder64 != 0 {
			return false
		}
	} else {
		power := powersOf10Uint256[scale]
		quotient.DivMod(&units, &power, &remainder)
		if !remainder.IsZero() {
			return false
		}
	}
	if quotient[1]|quotient[2]|quotient[3] == 0 {
		dst.SetUint64(quotient[0])
		return true
	}
	integer := &workspace.numerator
	quotient.IntoBig(&integer)
	dst.SetInt(&workspace.numerator)
	return true
}

func setIntegralNativeBigRat(dst *big.Rat, units, power uint64, scale int) bool {
	if scale > 0 && units%10 != 0 {
		return false
	}
	quotient := units / power
	if units-quotient*power != 0 {
		return false
	}
	dst.SetUint64(quotient)
	return true
}
