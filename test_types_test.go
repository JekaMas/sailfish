package sailfish

import "github.com/holiman/uint256"

type uint64DecimalPlaces0 struct{ Uint64Units }

func (uint64DecimalPlaces0) FractionalDecimalPlaces() DecimalPlaces { return 0 }

type uint64DecimalPlaces19 struct{ Uint64Units }

func (uint64DecimalPlaces19) FractionalDecimalPlaces() DecimalPlaces { return 19 }

type uint64DecimalPlaces20 struct{ Uint64Units }

func (uint64DecimalPlaces20) FractionalDecimalPlaces() DecimalPlaces { return 20 }

type uint256DecimalPlaces0 struct{ Uint256Units }

func (uint256DecimalPlaces0) FractionalDecimalPlaces() DecimalPlaces { return 0 }

type uint256DecimalPlaces6 struct{ Uint256Units }

func (uint256DecimalPlaces6) FractionalDecimalPlaces() DecimalPlaces { return 6 }

type uint256DecimalPlaces18 struct{ Uint256Units }

func (uint256DecimalPlaces18) FractionalDecimalPlaces() DecimalPlaces { return 18 }

type uint256DecimalPlaces37 struct{ Uint256Units }

func (uint256DecimalPlaces37) FractionalDecimalPlaces() DecimalPlaces { return 37 }

type uint256DecimalPlaces77 struct{ Uint256Units }

func (uint256DecimalPlaces77) FractionalDecimalPlaces() DecimalPlaces { return 77 }

type uint256DecimalPlaces78 struct{ Uint256Units }

func (uint256DecimalPlaces78) FractionalDecimalPlaces() DecimalPlaces { return 78 }

type price5 = FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64]
type wide18 = FixedDecimal[uint256DecimalPlaces18, uint256.Int]

const maxUint256Decimal = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
const maxUint256PlusOne = "115792089237316195423570985008687907853269984665640564039457584007913129639936"

func testFixedDecimalCodec[V FixedDecimalFormat[U], U Unit]() FixedDecimalCodec[V, U] {
	codec, _ := NewFixedDecimalCodec[V]()
	return codec
}

func testUint256FixedDecimalCodec(scale DecimalPlaces) Uint256FixedDecimalCodec {
	codec, _ := NewUint256FixedDecimalCodec(scale)
	return codec
}
