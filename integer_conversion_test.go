package sailfish

import (
	"errors"
	"math"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/holiman/uint256"
)

var integerConversionWideSink FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]

func TestFixedDecimalBigIntConversions(t *testing.T) {
	t.Parallel()

	t.Run("uint64", func(t *testing.T) {
		t.Parallel()
		codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
		for _, units := range []uint64{0, 1, math.MaxUint64} {
			source := new(big.Int).SetUint64(units)
			value, err := codec.FromBigInt(source)
			if err != nil || value.Units() != units {
				t.Fatalf("FromBigInt(%d) = %d, %v", units, value.Units(), err)
			}
			var destination big.Int
			if err = value.ToBigInt(&destination); err != nil || destination.Cmp(source) != 0 {
				t.Fatalf("ToBigInt(%d) = %s, %v", units, destination.String(), err)
			}
		}
	})

	t.Run("uint256", func(t *testing.T) {
		t.Parallel()
		codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
		for _, units := range []uint256.Int{
			{},
			{1},
			{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64},
		} {
			source := units.ToBig()
			value, err := codec.FromBigInt(source)
			if err != nil || value.Units() != units {
				t.Fatalf("FromBigInt(%v) = %v, %v", units, value.Units(), err)
			}
			var destination big.Int
			if err = value.ToBigInt(&destination); err != nil || destination.Cmp(source) != 0 {
				t.Fatalf("ToBigInt(%v) = %s, %v", units, destination.String(), err)
			}
		}
	})
}

func TestFixedDecimalU256Conversions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{"uint8", func(t *testing.T) {
			assertU256RoundTrip(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]](), uint8(math.MaxUint8))
		}},
		{"uint16", func(t *testing.T) {
			assertU256RoundTrip(t, testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]](), uint16(math.MaxUint16))
		}},
		{"uint32", func(t *testing.T) {
			assertU256RoundTrip(t, testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces5]](), uint32(math.MaxUint32))
		}},
		{"uint64", func(t *testing.T) {
			assertU256RoundTrip(t, testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]](), uint64(math.MaxUint64))
		}},
		{"uint256", func(t *testing.T) {
			units := uint256.Int{1, 2, 3, 4}
			codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
			value, err := codec.FromU256(units)
			if err != nil || value.Units() != units || value.ToU256() != units {
				t.Fatalf("round trip = %v, %v, %v", value.Units(), value.ToU256(), err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestFixedDecimalIntegerConversionsRejectInvalidInput(t *testing.T) {
	t.Parallel()

	codec8 := testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()

	if _, err := codec8.FromBigInt(nil); !errors.Is(err, ErrNilSource) {
		t.Fatalf("nil source error = %v", err)
	}
	if _, err := codec8.FromBigInt(big.NewInt(-1)); !errors.Is(err, ErrRange) {
		t.Fatalf("negative source error = %v", err)
	}
	if _, err := codec8.FromBigInt(big.NewInt(256)); !errors.Is(err, ErrRange) {
		t.Fatalf("uint8 overflow error = %v", err)
	}
	if _, err := codec8.FromU256(uint256.Int{256}); !errors.Is(err, ErrRange) {
		t.Fatalf("uint8 U256 overflow error = %v", err)
	}
	over256 := new(big.Int).Lsh(big.NewInt(1), 256)
	if _, err := codec256.FromBigInt(over256); !errors.Is(err, ErrRange) {
		t.Fatalf("uint256 overflow error = %v", err)
	}
	value := codec256.FromUnits(uint256.Int{1, 2, 3, 4})
	if err := value.ToBigInt(nil); !errors.Is(err, ErrNilDestination) {
		t.Fatalf("nil destination error = %v", err)
	}
}

func TestFixedDecimalIntegerConversionsPreserveRepresentationOwnership(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	value, err := codec.Parse("123.31232")
	if err != nil {
		t.Fatal(err)
	}
	source := new(big.Int).SetUint64(value.Units())
	fromBig, err := codec.FromBigInt(source)
	if err != nil {
		t.Fatal(err)
	}
	if fromBig.HasRepresentation() {
		t.Fatal("numeric conversion retained unrelated text")
	}
	var destination big.Int
	if err = value.ToBigInt(&destination); err != nil {
		t.Fatal(err)
	}
	if !value.HasRepresentation() || value.String() != "123.31232" {
		t.Fatal("read-only conversion changed retained representation")
	}
	source.SetUint64(1)
	if fromBig.Units() != 12_331_232 {
		t.Fatal("converted decimal aliases big.Int source")
	}
	destination.SetUint64(2)
	if value.Units() != 12_331_232 {
		t.Fatal("converted big.Int aliases decimal units")
	}

	wideCodec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	wideSource := uint256.Int{1, 2, 3, 4}
	wideValue, err := wideCodec.FromU256(wideSource)
	if err != nil {
		t.Fatal(err)
	}
	wideSource[0] = 99
	wideOutput := wideValue.ToU256()
	wideOutput[1] = 99
	if wideValue.Units() != (uint256.Int{1, 2, 3, 4}) {
		t.Fatal("converted decimal aliases uint256 source or output")
	}

	zero := wideCodec.FromUnits(uint256.Int{})
	destination.SetUint64(99)
	if err = zero.ToBigInt(&destination); err != nil || destination.Sign() != 0 {
		t.Fatalf("zero ToBigInt did not clear destination: %s, %v", destination.String(), err)
	}
}

func TestFixedDecimalIntegerConversionProperties(t *testing.T) {
	t.Parallel()

	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	rng := rand.New(rand.NewPCG(0x1234, 0x5678))
	for range 10_000 {
		units64 := rng.Uint64()
		big64 := new(big.Int).SetUint64(units64)
		value64, err := codec64.FromBigInt(big64)
		if err != nil || value64.ToU256() != (uint256.Int{units64}) {
			t.Fatalf("uint64 conversion failed: %d %v", units64, err)
		}

		units256 := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		big256 := units256.ToBig()
		value256, err := codec256.FromBigInt(big256)
		if err != nil || value256.ToU256() != units256 {
			t.Fatalf("uint256 FromBigInt failed: %v %v", units256, err)
		}
		var round big.Int
		if err = value256.ToBigInt(&round); err != nil || round.Cmp(big256) != 0 {
			t.Fatalf("uint256 ToBigInt failed: %v %v", units256, err)
		}
	}
}

func TestFixedDecimalIntegerConversionsDoNotAllocate(t *testing.T) {
	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	big64 := new(big.Int).SetUint64(math.MaxUint64)
	units256 := uint256.Int{1, 2, 3, 4}
	big256 := units256.ToBig()
	value64 := codec64.FromUnits(math.MaxUint64)
	value256 := codec256.FromUnits(units256)
	var destination big.Int
	_ = value256.ToBigInt(&destination)

	assertAllocs(t, "from BigInt uint64", 0, func() {
		allocationPriceSink, allocationErrorSink = codec64.FromBigInt(big64)
	})
	assertAllocs(t, "from BigInt uint256", 0, func() {
		integerConversionWideSink, allocationErrorSink = codec256.FromBigInt(big256)
	})
	assertAllocs(t, "from U256 uint64", 0, func() {
		allocationPriceSink, allocationErrorSink = codec64.FromU256(uint256.Int{math.MaxUint64})
	})
	assertAllocs(t, "from U256 uint256", 0, func() {
		integerConversionWideSink, allocationErrorSink = codec256.FromU256(units256)
	})
	assertAllocs(t, "to U256 uint64", 0, func() {
		benchU256Sink = value64.ToU256()
	})
	assertAllocs(t, "to U256 uint256", 0, func() {
		benchU256Sink = value256.ToU256()
	})
	assertAllocs(t, "to reused BigInt uint64", 0, func() {
		allocationErrorSink = value64.ToBigInt(&destination)
	})
	assertAllocs(t, "to reused BigInt uint256", 0, func() {
		allocationErrorSink = value256.ToBigInt(&destination)
	})
}

func assertU256RoundTrip[V FixedDecimalFormat[U], U NativeUnit](t *testing.T, codec FixedDecimalCodec[V, U], units U) {
	t.Helper()
	want := uint256.Int{uint64(units)}
	value, err := codec.FromU256(want)
	if err != nil || value.Units() != units || value.ToU256() != want {
		t.Fatalf("round trip = %v, %v, %v", value.Units(), value.ToU256(), err)
	}
}
