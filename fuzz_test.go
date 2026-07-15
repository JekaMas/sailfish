package sailfish

import (
	"math/big"
	"testing"

	"github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

func FuzzPriceInUint64UnitsDecimalPlaces9ParseRoundTrip(f *testing.F) {
	for _, seed := range []string{
		"0", "1", "1.2", "123.312320000", "18446744073.709551615",
		"", "!!!", " 1", "+1", "-1", "1e3", "1.0000000000",
	} {
		f.Add(seed)
	}

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]]()
	f.Fuzz(func(t *testing.T, input string) {
		value, err := codec.ParseCompact(input)
		if err != nil {
			return
		}
		canonical := value.String()
		round, err := codec.ParseCompact(canonical)
		if err != nil || !round.Equal(value) {
			t.Fatalf("%q -> %q -> %#v, %v", input, canonical, round, err)
		}
	})
}

func FuzzUint64UnitsRoundTrip(f *testing.F) {
	for _, seed := range []uint64{0, 1, 9, 10, 99, 100, 1_000_000_000, ^uint64(0)} {
		f.Add(seed)
	}

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]]()
	f.Fuzz(func(t *testing.T, units uint64) {
		value := codec.FromUnits(units)
		round, err := codec.ParseCompact(value.String())
		if err != nil || round.Units() != units {
			t.Fatalf("%d -> %q -> %d, %v", units, value.String(), round.Units(), err)
		}
	})
}

func FuzzNativeUnitWidthsRoundTrip(f *testing.F) {
	for _, seed := range []struct {
		units uint64
		width uint8
	}{
		{0, 0}, {255, 0}, {256, 1}, {65_535, 1},
		{65_536, 2}, {4_294_967_295, 2}, {^uint64(0), 3},
	} {
		f.Add(seed.units, seed.width)
	}

	f.Fuzz(func(t *testing.T, units uint64, width uint8) {
		switch width % 4 {
		case 0:
			fuzzNativeUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces2]](), uint8(units))
		case 1:
			fuzzNativeUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces4]](), uint16(units))
		case 2:
			fuzzNativeUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces9]](), uint32(units))
		case 3:
			fuzzNativeUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces19]](), units)
		}
	})
}

func fuzzNativeUnitRoundTrip[V FixedDecimalFormat[U], U NativeUnit](t *testing.T, codec FixedDecimalCodec[V, U], units U) {
	t.Helper()

	value := codec.FromUnits(units)
	round, err := codec.ParseCompact(value.String())
	if err != nil || round.Units() != units {
		t.Fatalf("%v -> %q -> %v, %v", units, value.String(), round.Units(), err)
	}
}

func FuzzUint256UnitsRoundTrip(f *testing.F) {
	seeds := []uint256.Int{
		{},
		{1},
		{^uint64(0)},
		{1, 2, 3, 4},
		{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}

	codec := testFixedDecimalCodec[uint256DecimalPlaces18]()
	f.Fuzz(func(t *testing.T, limb0, limb1, limb2, limb3 uint64) {
		units := uint256.Int{limb0, limb1, limb2, limb3}
		value := codec.FromUnits(units)
		round, err := codec.ParseCompact(value.String())
		if err != nil || round.Units() != units {
			t.Fatalf("%#v -> %q -> %#v, %v", units, value.String(), round.Units(), err)
		}
	})
}

func FuzzIntegerUnitConversions(f *testing.F) {
	seeds := []uint256.Int{
		{},
		{1},
		{^uint64(0)},
		{1, 2, 3, 4},
		{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}

	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	f.Fuzz(func(t *testing.T, limb0, limb1, limb2, limb3 uint64) {
		units := uint256.Int{limb0, limb1, limb2, limb3}
		fromU256, err := codec.FromU256(units)
		if err != nil || fromU256.ToU256() != units {
			t.Fatalf("U256 round trip: %#v, %v", fromU256.ToU256(), err)
		}

		source := units.ToBig()
		fromBig, err := codec.FromBigInt(source)
		if err != nil || fromBig.ToU256() != units {
			t.Fatalf("BigInt input: %#v, %v", fromBig.ToU256(), err)
		}
		var destination big.Int
		if err = fromBig.ToBigInt(&destination); err != nil || destination.Cmp(source) != 0 {
			t.Fatalf("BigInt output: %s, %v", destination.String(), err)
		}
	})
}

func FuzzBigRatExactRoundTrip(f *testing.F) {
	for _, seed := range []uint64{0, 1, 10, 99_999, 100_000, ^uint64(0)} {
		f.Add(seed)
	}

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	denominator := new(big.Int).SetUint64(100_000)
	f.Fuzz(func(t *testing.T, units uint64) {
		source := new(big.Rat).SetFrac(new(big.Int).SetUint64(units), denominator)
		value, err := codec.FromBigRat(source)
		if err != nil || value.Units() != units {
			t.Fatalf("FromBigRat(%s) = %d, %v", source.String(), value.Units(), err)
		}
		var destination big.Rat
		var workspace BigRatWorkspace
		if err = value.ToBigRat(&destination, &workspace); err != nil || destination.Cmp(source) != 0 {
			t.Fatalf("ToBigRat(%d) = %s, %v; want %s", units, destination.String(), err, source.String())
		}
	})
}

func FuzzCrossScaleExactArithmetic(f *testing.F) {
	for _, seed := range [][2]uint32{{0, 0}, {120, 3}, {999_999, 1}, {1, 999_999}} {
		f.Add(seed[0], seed[1])
	}

	codec2 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]()
	codec3 := testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces3]]()
	f.Fuzz(func(t *testing.T, leftUnits, rightUnits uint32) {
		left := codec2.FromUnits(uint64(leftUnits))
		right := codec3.FromUnits(rightUnits)
		sum, err := AddAs[PriceInUint64Units[DecimalPlaces5]](left, right)
		want := uint64(leftUnits)*1_000 + uint64(rightUnits)*100
		if err != nil || sum.Units() != want {
			t.Fatalf("sum(%d,%d) = %d, %v; want %d", leftUnits, rightUnits, sum.Units(), err, want)
		}
		round, err := SubAs[PriceInUint64Units[DecimalPlaces5]](sum, right)
		if err != nil || round.Units() != uint64(leftUnits)*1_000 {
			t.Fatalf("round(%d,%d) = %d, %v", leftUnits, rightUnits, round.Units(), err)
		}
	})
}

func FuzzJSONRoundTrip(f *testing.F) {
	for _, seed := range []string{"0", "1.2", "123.31232", "!!!", `\u0031.20000`} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		var value price5
		if err := json.Unmarshal(encoded, &value); err != nil {
			return
		}
		round, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var decoded price5
		if err := json.Unmarshal(round, &decoded); err != nil || !decoded.Equal(value) {
			t.Fatalf("%q -> %s -> %s: %v", input, encoded, round, err)
		}
	})
}

func FuzzCBORUint64RoundTrip(f *testing.F) {
	for _, seed := range []uint64{0, 1, 23, 24, 255, 256, 65_535, 65_536, ^uint64(0)} {
		f.Add(seed)
	}

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]]()
	f.Fuzz(func(t *testing.T, units uint64) {
		value := codec.FromUnits(units)
		var buffer [MaxCBORSize]byte
		wire := value.AppendCBOR(buffer[:0])
		decoded, err := codec.ParseCBOR(wire)
		if err != nil || decoded.Units() != units {
			t.Fatalf("%d -> %x -> %d, %v", units, wire, decoded.Units(), err)
		}
	})
}

func FuzzCBORUint256RoundTrip(f *testing.F) {
	seeds := []uint256.Int{{}, {1}, {^uint64(0)}, {0, 1}, {1, 2, 3, 4}, {^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}

	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	f.Fuzz(func(t *testing.T, limb0, limb1, limb2, limb3 uint64) {
		units := uint256.Int{limb0, limb1, limb2, limb3}
		value := codec.FromUnits(units)
		var buffer [MaxCBORSize]byte
		wire := value.AppendCBOR(buffer[:0])
		decoded, err := codec.ParseCBOR(wire)
		if err != nil || decoded.Units() != units {
			t.Fatalf("%#v -> %x -> %#v, %v", units, wire, decoded.Units(), err)
		}
	})
}

func FuzzCBORDecoderAcceptsOnlyPreferredRoundTrips(f *testing.F) {
	for _, seed := range [][]byte{
		{0x00},
		{0x17},
		{0x18, 0x18},
		{0x1b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0xc2, 0x49, 0x01, 0, 0, 0, 0, 0, 0, 0, 0},
		{},
		{0x18, 0x00},
		{0x81, 0x00},
	} {
		f.Add(seed)
	}

	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	f.Fuzz(func(t *testing.T, raw []byte) {
		value, err := codec.ParseCBOR(raw)
		if err != nil {
			return
		}
		var buffer [MaxCBORSize]byte
		canonical := value.AppendCBOR(buffer[:0])
		if string(canonical) != string(raw) {
			t.Fatalf("decoder accepted non-preferred %x; preferred is %x", raw, canonical)
		}
	})
}

func FuzzCBORFirstConsumesExactlyOnePreferredValue(f *testing.F) {
	for _, seed := range [][]byte{
		{0x00},
		{0x17, 0xff},
		{0x18, 0x18, 0x01},
		{0x1b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00},
		{0xc2, 0x49, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0x01},
		{},
		{0x18, 0x00, 0x01},
		{0xc2},
	} {
		f.Add(seed)
	}

	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	f.Fuzz(func(t *testing.T, raw []byte) {
		value, rest, err := codec.ParseCBORFirst(raw)
		if err != nil {
			if rest != nil {
				t.Fatalf("failed decode returned rest %x", rest)
			}
			return
		}
		consumed := len(raw) - len(rest)
		if consumed <= 0 {
			t.Fatalf("decoded without consuming input %x", raw)
		}
		var buffer [MaxCBORSize]byte
		canonical := value.AppendCBOR(buffer[:0])
		if string(canonical) != string(raw[:consumed]) {
			t.Fatalf("decoder consumed non-preferred %x; preferred is %x", raw[:consumed], canonical)
		}
	})
}
