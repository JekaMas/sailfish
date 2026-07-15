package sailfish

import "github.com/holiman/uint256"

// DecimalPlaces is the exact number of fractional digits represented after
// the decimal point. For example, DecimalPlaces(5) means raw units 12_331_232
// represent the decimal value 123.31232.
type DecimalPlaces uint8

// Unit is the closed set of scaled-integer storage backends supported by
// FixedDecimal.
type Unit interface {
	comparable
	uint8 | uint16 | uint32 | uint64 | uint256.Int
}

// NativeUnit is the subset backed by Go's native unsigned integer types.
type NativeUnit interface {
	comparable
	uint8 | uint16 | uint32 | uint64
}

// StaticDecimalPlaces supplies an exact compile-time count of fractional
// decimal places. Implement it on a zero-sized value type with a value
// receiver.
type StaticDecimalPlaces interface {
	FractionalDecimalPlaces() DecimalPlaces
}

// unitSystem is intentionally sealed. A format selects its unit backend by
// embedding one of the concrete Uint*Units providers.
type unitSystem[U Unit] interface {
	unitParseString(string, int) (U, bool, Error)
	unitParseBytes([]byte, int) (U, bool, Error)
	unitAppend([]byte, U, int) []byte
	unitString(U, int) string
	unitLen(U, int) int
	unitCompare(U, U) int
	unitCBORLen(U) int
	unitAppendCBOR([]byte, U) []byte
	unitParseCBOR([]byte) (U, Error)
	unitParseCBORFirst([]byte) (U, int, Error)
}

// FixedDecimalFormat binds a semantic decimal kind, exact fractional decimal
// places, and one scaled-integer backend. Prefer PriceInUint*Units and
// AmountInUint*Units. A custom format is normally a zero-sized type:
//
//	type QuotePriceWith5DecimalPlaces struct{ sailfish.Uint64Units }
//	func (QuotePriceWith5DecimalPlaces) FractionalDecimalPlaces() sailfish.DecimalPlaces {
//		return 5
//	}
type FixedDecimalFormat[U Unit] interface {
	StaticDecimalPlaces
	unitSystem[U]
}

func checkedFractionalDecimalPlaces[V FixedDecimalFormat[U], U Unit]() (int, Error) {
	var format V
	decimalPlaces := int(format.FractionalDecimalPlaces())
	if decimalPlaces > maxFractionalDecimalPlaces[U]() {
		return 0, ErrUnsupportedFractionalDecimalPlaces
	}
	return decimalPlaces, ""
}

func maxFractionalDecimalPlaces[U Unit]() int {
	var unit U
	switch any(unit).(type) {
	case uint8:
		return maxUint8Scale
	case uint16:
		return maxUint16Scale
	case uint32:
		return maxUint32Scale
	case uint64:
		return maxUint64Scale
	case uint256.Int:
		return maxUint256Scale
	default:
		return -1
	}
}

func unitDecimalDigits[U Unit]() int {
	var unit U
	switch any(unit).(type) {
	case uint8:
		return 3
	case uint16:
		return 5
	case uint32:
		return 10
	case uint64:
		return 20
	case uint256.Int:
		return 78
	default:
		return 0
	}
}

func fractionalDecimalPlacesOf[V FixedDecimalFormat[U], U Unit]() int {
	decimalPlaces, _ := checkedFractionalDecimalPlaces[V, U]()
	return decimalPlaces
}
