package sailfish

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

func FuzzPriceUint64Fraction9ParseRoundTrip(f *testing.F) {
	for _, seed := range []string{
		"0", "1", "1.2", "123.312320000", "18446744073.709551615",
		"", "!!!", " 1", "+1", "-1", "1e3", "1.0000000000",
	} {
		f.Add(seed)
	}

	codec := MustCodec[PriceUint64[Fraction9]]()
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

	codec := MustCodec[PriceUint64[Fraction9]]()
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
			fuzzNativeUnitRoundTrip(t, MustCodec[PriceUint8[Fraction2]](), uint8(units))
		case 1:
			fuzzNativeUnitRoundTrip(t, MustCodec[PriceUint16[Fraction4]](), uint16(units))
		case 2:
			fuzzNativeUnitRoundTrip(t, MustCodec[PriceUint32[Fraction9]](), uint32(units))
		case 3:
			fuzzNativeUnitRoundTrip(t, MustCodec[PriceUint64[Fraction19]](), units)
		}
	})
}

func fuzzNativeUnitRoundTrip[V Venue[U], U NativeUnit](t *testing.T, codec Codec[V, U], units U) {
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

	codec := MustCodec[uint256Scale18]()
	f.Fuzz(func(t *testing.T, limb0, limb1, limb2, limb3 uint64) {
		units := uint256.Int{limb0, limb1, limb2, limb3}
		value := codec.FromUnits(units)
		round, err := codec.ParseCompact(value.String())
		if err != nil || round.Units() != units {
			t.Fatalf("%#v -> %q -> %#v, %v", units, value.String(), round.Units(), err)
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
