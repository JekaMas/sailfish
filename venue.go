package sailfish

import "github.com/holiman/uint256"

// Notion is the fixed number of digits after the decimal point.
type Notion uint8

// Unit is the closed set of scaled-integer storage backends supported by
// Decimal.
type Unit interface {
	comparable
	uint8 | uint16 | uint32 | uint64 | uint256.Int
}

// NativeUnit is the subset backed by Go's native unsigned integer types.
type NativeUnit interface {
	comparable
	uint8 | uint16 | uint32 | uint64
}

// VenueScale supplies a fixed decimal scale. Implement it on a zero-sized
// value type with a value receiver.
type VenueScale interface {
	NotionScale() Notion
}

// unitSystem is intentionally sealed. A venue selects its unit backend by
// embedding one of the concrete Uint*Units providers.
type unitSystem[U Unit] interface {
	unitParseString(string, int) (U, bool, Error)
	unitParseBytes([]byte, int) (U, bool, Error)
	unitAppend([]byte, U, int) []byte
	unitString(U, int) string
	unitLen(U, int) int
	unitCompare(U, U) int
}

// Venue binds a fixed scale to one unit backend.
//
// A custom venue is normally a zero-sized type. Prefer PriceUint* and
// AmountUint* when their semantic distinction applies:
//
//	type QuoteFraction5 struct{ sailfish.Uint64Units }
//	func (QuoteFraction5) NotionScale() sailfish.Notion { return 5 }
type Venue[U Unit] interface {
	VenueScale
	unitSystem[U]
}

func checkedScale[V Venue[U], U Unit]() (int, Error) {
	var venue V
	scale := int(venue.NotionScale())
	if scale > maxScale[U]() {
		return 0, ErrScale
	}
	return scale, ""
}

func maxScale[U Unit]() int {
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
		panic("sailfish: unreachable unit type")
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
		panic("sailfish: unreachable unit type")
	}
}

func mustScale[V Venue[U], U Unit]() int {
	scale, err := checkedScale[V, U]()
	if err != "" {
		panic(err)
	}
	return scale
}
