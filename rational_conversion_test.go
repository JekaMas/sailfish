package sailfish

import (
	"errors"
	"math/big"
	"math/rand"
	"testing"

	"github.com/holiman/uint256"
)

func TestFixedDecimalBigRatExactConversions(t *testing.T) {
	t.Parallel()

	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	source, ok := new(big.Rat).SetString("123.31232")
	if !ok {
		t.Fatal("invalid test rational")
	}
	value, err := codec64.FromBigRat(source)
	if err != nil || value.Units() != 12_331_232 {
		t.Fatalf("FromBigRat = %d, %v", value.Units(), err)
	}

	var destination big.Rat
	var workspace BigRatWorkspace
	if err = value.ToBigRat(&destination, &workspace); err != nil || destination.Cmp(source) != 0 {
		t.Fatalf("ToBigRat = %s, %v", destination.String(), err)
	}

	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	wideSource := new(big.Rat).SetFrac(
		new(big.Int).Lsh(big.NewInt(1), 200),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
	)
	wide, err := codec256.FromBigRat(wideSource)
	if err != nil || wide.Units() != (uint256.Int{0, 0, 0, 256}) {
		t.Fatalf("wide FromBigRat = %#v, %v", wide.Units(), err)
	}
	if err = wide.ToBigRat(&destination, &workspace); err != nil || destination.Cmp(wideSource) != 0 {
		t.Fatalf("wide ToBigRat = %s, %v", destination.String(), err)
	}
}

func TestBigRatConversionsMatchExactReference(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(23))
	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	var destination big.Rat
	var workspace BigRatWorkspace
	denominator64 := new(big.Int).SetUint64(100_000)
	denominator256 := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	for range 10_000 {
		units64 := rng.Uint64()
		want64 := new(big.Rat).SetFrac(new(big.Int).SetUint64(units64), denominator64)
		value64 := codec64.FromUnits(units64)
		if err := value64.ToBigRat(&destination, &workspace); err != nil || destination.Cmp(want64) != 0 {
			t.Fatalf("uint64 ToBigRat(%d) = %s, %v; want %s", units64, destination.String(), err, want64.String())
		}
		round64, err := codec64.FromBigRat(want64)
		if err != nil || round64.Units() != units64 {
			t.Fatalf("uint64 FromBigRat(%s) = %d, %v", want64.String(), round64.Units(), err)
		}

		units256 := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		want256 := new(big.Rat).SetFrac(units256.ToBig(), denominator256)
		value256 := codec256.FromUnits(units256)
		if err := value256.ToBigRat(&destination, &workspace); err != nil || destination.Cmp(want256) != 0 {
			t.Fatalf("uint256 ToBigRat(%#v) = %s, %v; want %s", units256, destination.String(), err, want256.String())
		}
		round256, err := codec256.FromBigRat(want256)
		if err != nil || round256.Units() != units256 {
			t.Fatalf("uint256 FromBigRat(%s) = %#v, %v", want256.String(), round256.Units(), err)
		}
	}
}

func TestFixedDecimalBigRatRejectsInexactAndInvalidInput(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	if _, err := codec.FromBigRat(nil); !errors.Is(err, ErrNilSource) {
		t.Fatalf("nil source = %v", err)
	}
	if _, err := codec.FromBigRat(big.NewRat(-1, 1)); !errors.Is(err, ErrRange) {
		t.Fatalf("negative source = %v", err)
	}
	if _, err := codec.FromBigRat(big.NewRat(1, 3)); !errors.Is(err, ErrPrecision) {
		t.Fatalf("inexact source = %v", err)
	}
	if _, err := codec.FromBigRat(new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 80))); !errors.Is(err, ErrRange) {
		t.Fatalf("overflow source = %v", err)
	}

	value := codec.FromUnits(1)
	var destination big.Rat
	var workspace BigRatWorkspace
	if err := value.ToBigRat(nil, &workspace); !errors.Is(err, ErrNilDestination) {
		t.Fatalf("nil destination = %v", err)
	}
	if err := value.ToBigRat(&destination, nil); !errors.Is(err, ErrNilWorkspace) {
		t.Fatalf("nil workspace = %v", err)
	}
}

func TestFixedDecimalBigRatAcrossUnitBackendsAndScaleZero(t *testing.T) {
	t.Parallel()

	u8, err := testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces2]]().FromBigRat(big.NewRat(51, 20))
	if err != nil || u8.Units() != 255 {
		t.Fatalf("uint8 rational = %d, %v", u8.Units(), err)
	}
	u16, err := testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces4]]().FromBigRat(big.NewRat(13_107, 2_000))
	if err != nil || u16.Units() != 65_535 {
		t.Fatalf("uint16 rational = %d, %v", u16.Units(), err)
	}
	u32, err := testFixedDecimalCodec[AmountInUint32Units[DecimalPlaces6]]().FromBigRat(big.NewRat(123_456_789, 1_000_000))
	if err != nil || u32.Units() != 123_456_789 {
		t.Fatalf("uint32 rational = %d, %v", u32.Units(), err)
	}
	integerCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces0]]()
	integer, err := integerCodec.FromBigRat(big.NewRat(42, 1))
	if err != nil || integer.Units() != 42 {
		t.Fatalf("scale-zero integer = %d, %v", integer.Units(), err)
	}
	if _, err = integerCodec.FromBigRat(big.NewRat(1, 2)); !errors.Is(err, ErrPrecision) {
		t.Fatalf("scale-zero fraction = %v", err)
	}
	wideDenominator := new(big.Int).Lsh(big.NewInt(1), 200)
	if _, err = testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]().FromBigRat(
		new(big.Rat).SetFrac(big.NewInt(1), wideDenominator),
	); !errors.Is(err, ErrPrecision) {
		t.Fatalf("wide inexact denominator = %v", err)
	}
}

func TestFixedDecimalBigRatSteadyStateDoesNotAllocate(t *testing.T) {
	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	rational64 := big.NewRat(3_082_808, 25_000)
	rational256 := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 200))
	value64 := codec64.FromUnits(12_331_232)
	value256 := codec256.FromUnits(uint256.Int{1, 2, 3, 4})
	var destination big.Rat
	var workspace BigRatWorkspace
	_ = value256.ToBigRat(&destination, &workspace)

	assertAllocs(t, "from BigRat uint64", 0, func() {
		allocationPriceSink, allocationErrorSink = codec64.FromBigRat(rational64)
	})
	assertAllocs(t, "from BigRat uint256", 0, func() {
		integerConversionWideSink, allocationErrorSink = codec256.FromBigRat(rational256)
	})
	// math/big.Rat.SetFrac owns and normalizes copies of both operands through
	// unexported words. Its public API therefore has an allocation floor even
	// when Sailfish's destination and workspace are warm and caller-owned.
	assertAllocs(t, "to BigRat uint64 reused", 3, func() {
		allocationErrorSink = value64.ToBigRat(&destination, &workspace)
	})
	assertAllocs(t, "to BigRat uint256 reused", 5, func() {
		allocationErrorSink = value256.ToBigRat(&destination, &workspace)
	})
	integer64 := codec64.FromUnits(12_300_000)
	integer256 := codec256.FromUnits(uint256.Int{1_000_000_000_000_000_000})
	assertAllocs(t, "to integral BigRat uint64 reused", 0, func() {
		allocationErrorSink = integer64.ToBigRat(&destination, &workspace)
	})
	assertAllocs(t, "to integral BigRat uint256 reused", 0, func() {
		allocationErrorSink = integer256.ToBigRat(&destination, &workspace)
	})
}
