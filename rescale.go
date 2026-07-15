package sailfish

import (
	"math"
	"math/bits"

	"github.com/holiman/uint256"
)

// powersOf10Uint256 is initialized once and treated as immutable. A table
// removes scale-dependent multiplication loops from exact rescaling and
// rational conversion hot paths.
var powersOf10Uint256 = func() [maxUint256Scale + 1]uint256.Int { //nolint:gochecknoglobals
	var powers [maxUint256Scale + 1]uint256.Int
	powers[0] = uint256.Int{1}
	ten := uint256.Int{10}
	for i := 1; i < len(powers); i++ {
		powers[i].Mul(&powers[i-1], &ten)
	}
	return powers
}()

// Rescale converts value to an explicitly selected format without rounding.
// Downscaling rejects discarded nonzero units; upscaling rejects overflow.
func Rescale[
	ToV FixedDecimalFormat[ToU], ToU Unit,
	FromV FixedDecimalFormat[FromU], FromU Unit,
](value FixedDecimal[FromV, FromU]) (FixedDecimal[ToV, ToU], error) {
	toScale, err := checkedFractionalDecimalPlaces[ToV, ToU]()
	if err != "" {
		return FixedDecimal[ToV, ToU]{}, boxedError(err)
	}
	if isNativeUnit[ToU]() {
		units, err := rescaleToNative(value, toScale)
		if err != "" {
			return FixedDecimal[ToV, ToU]{}, boxedError(err)
		}
		result, err := fixedDecimalFromNativeUnits[ToV, ToU](units)
		if err != "" {
			return FixedDecimal[ToV, ToU]{}, boxedError(err)
		}
		return result, nil
	}
	units, err := rescaleToU256(value, toScale)
	if err != "" {
		return FixedDecimal[ToV, ToU]{}, boxedError(err)
	}
	return (FixedDecimalCodec[ToV, ToU]{}).FromU256(units)
}

// AddAs exactly rescales both operands to the selected result format and adds
// them. Use the same-format Add method when no rescaling is required.
func AddAs[
	ResultV FixedDecimalFormat[ResultU], ResultU Unit,
	AV FixedDecimalFormat[AU], AU Unit,
	BV FixedDecimalFormat[BU], BU Unit,
](a FixedDecimal[AV, AU], b FixedDecimal[BV, BU]) (FixedDecimal[ResultV, ResultU], error) {
	resultScale, err := checkedFractionalDecimalPlaces[ResultV, ResultU]()
	if err != "" {
		return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
	}
	if isNativeUnit[ResultU]() {
		left, err := rescaleToNative(a, resultScale)
		if err != "" {
			return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
		}
		right, err := rescaleToNative(b, resultScale)
		if err != "" {
			return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
		}
		sum, carry := bits.Add64(left, right, 0)
		if carry != 0 {
			return FixedDecimal[ResultV, ResultU]{}, boxedErrOverflow
		}
		result, err := fixedDecimalFromNativeUnits[ResultV, ResultU](sum)
		if err != "" {
			return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
		}
		return result, nil
	}
	left, err := rescaleToU256(a, resultScale)
	if err != "" {
		return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
	}
	right, err := rescaleToU256(b, resultScale)
	if err != "" {
		return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
	}
	sum, overflow := addUnits(left, right)
	if overflow {
		return FixedDecimal[ResultV, ResultU]{}, boxedErrOverflow
	}
	return (FixedDecimalCodec[ResultV, ResultU]{}).FromU256(sum)
}

// SubAs exactly rescales both operands to the selected result format and
// subtracts them. Use the same-format Sub method when no rescaling is required.
func SubAs[
	ResultV FixedDecimalFormat[ResultU], ResultU Unit,
	AV FixedDecimalFormat[AU], AU Unit,
	BV FixedDecimalFormat[BU], BU Unit,
](a FixedDecimal[AV, AU], b FixedDecimal[BV, BU]) (FixedDecimal[ResultV, ResultU], error) {
	resultScale, err := checkedFractionalDecimalPlaces[ResultV, ResultU]()
	if err != "" {
		return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
	}
	if isNativeUnit[ResultU]() {
		left, err := rescaleToNative(a, resultScale)
		if err != "" {
			return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
		}
		right, err := rescaleToNative(b, resultScale)
		if err != "" {
			return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
		}
		difference, borrow := bits.Sub64(left, right, 0)
		if borrow != 0 {
			return FixedDecimal[ResultV, ResultU]{}, boxedErrUnderflow
		}
		result, err := fixedDecimalFromNativeUnits[ResultV, ResultU](difference)
		if err != "" {
			return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
		}
		return result, nil
	}
	left, err := rescaleToU256(a, resultScale)
	if err != "" {
		return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
	}
	right, err := rescaleToU256(b, resultScale)
	if err != "" {
		return FixedDecimal[ResultV, ResultU]{}, boxedError(err)
	}
	difference, underflow := subUnits(left, right)
	if underflow {
		return FixedDecimal[ResultV, ResultU]{}, boxedErrUnderflow
	}
	return (FixedDecimalCodec[ResultV, ResultU]{}).FromU256(difference)
}

func isNativeUnit[U Unit]() bool {
	var zero U
	switch any(zero).(type) {
	case uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func rescaleToNative[V FixedDecimalFormat[U], U Unit](
	value FixedDecimal[V, U],
	toScale int,
) (uint64, Error) {
	fromScale, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return 0, err
	}
	delta := max(fromScale, toScale) - min(fromScale, toScale)
	if delta > maxUint64Scale {
		wide, err := rescaleToU256(value, toScale)
		if err != "" {
			return 0, err
		}
		if wide[1]|wide[2]|wide[3] != 0 {
			return 0, ErrRange
		}
		return wide[0], ""
	}
	var units uint64
	switch unitValue := any(value.units).(type) {
	case uint8:
		units = uint64(unitValue)
	case uint16:
		units = uint64(unitValue)
	case uint32:
		units = uint64(unitValue)
	case uint64:
		units = unitValue
	case uint256.Int:
		if unitValue[1]|unitValue[2]|unitValue[3] != 0 {
			wide, err := rescaleToU256(value, toScale)
			if err != "" {
				return 0, err
			}
			if wide[1]|wide[2]|wide[3] != 0 {
				return 0, ErrRange
			}
			return wide[0], ""
		}
		units = unitValue[0]
	}
	if fromScale == toScale {
		return units, ""
	}
	if fromScale < toScale {
		hi, result := bits.Mul64(units, powersOf10Uint64[delta])
		if hi != 0 {
			return 0, ErrRange
		}
		return result, ""
	}

	power := powersOf10Uint64[delta]
	result := units / power
	if units-result*power != 0 {
		return 0, ErrPrecision
	}
	return result, ""
}

func fixedDecimalFromNativeUnits[V FixedDecimalFormat[U], U Unit](
	units uint64,
) (FixedDecimal[V, U], Error) {
	var zero U
	switch any(zero).(type) {
	case uint8:
		if units > math.MaxUint8 {
			return FixedDecimal[V, U]{}, ErrRange
		}
		return FixedDecimal[V, U]{units: any(uint8(units)).(U)}, ""
	case uint16:
		if units > math.MaxUint16 {
			return FixedDecimal[V, U]{}, ErrRange
		}
		return FixedDecimal[V, U]{units: any(uint16(units)).(U)}, ""
	case uint32:
		if units > math.MaxUint32 {
			return FixedDecimal[V, U]{}, ErrRange
		}
		return FixedDecimal[V, U]{units: any(uint32(units)).(U)}, ""
	case uint64:
		return FixedDecimal[V, U]{units: any(units).(U)}, ""
	default:
		return FixedDecimal[V, U]{}, ErrRange
	}
}

func rescaleToU256[V FixedDecimalFormat[U], U Unit](
	value FixedDecimal[V, U],
	toScale int,
) (uint256.Int, Error) {
	fromScale, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return uint256.Int{}, err
	}
	units := unitsToU256(value.units)
	if fromScale == toScale {
		return units, ""
	}
	if fromScale < toScale {
		delta := toScale - fromScale
		if delta <= maxUint64Scale {
			result, overflow := uint256MulSmallAdd(units, powersOf10Uint64[delta], 0)
			if overflow {
				return uint256.Int{}, ErrRange
			}
			return result, ""
		}
		var result uint256.Int
		if _, overflow := result.MulOverflow(&units, &powersOf10Uint256[delta]); overflow {
			return uint256.Int{}, ErrRange
		}
		return result, ""
	}

	delta := fromScale - toScale
	if delta <= maxUint64Scale {
		quotient, remainder := uint256DivMod64(units, powersOf10Uint64[delta])
		if remainder != 0 {
			return uint256.Int{}, ErrPrecision
		}
		return quotient, ""
	}
	var quotient, remainder uint256.Int
	quotient.DivMod(&units, &powersOf10Uint256[delta], &remainder)
	if !remainder.IsZero() {
		return uint256.Int{}, ErrPrecision
	}
	return quotient, ""
}
