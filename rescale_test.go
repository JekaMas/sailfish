package sailfish

import (
	"errors"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/holiman/uint256"
)

func TestRescaleAcrossFractionalDecimalPlaces(t *testing.T) {
	t.Parallel()

	scale2 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]().FromUnits(120)
	scale3, err := Rescale[PriceInUint64Units[DecimalPlaces3]](scale2)
	if err != nil || scale3.Units() != 1_200 {
		t.Fatalf("upscale = %d, %v", scale3.Units(), err)
	}
	round, err := Rescale[PriceInUint64Units[DecimalPlaces2]](scale3)
	if err != nil || round.Units() != 120 {
		t.Fatalf("downscale = %d, %v", round.Units(), err)
	}

	inexact := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces3]]().FromUnits(1_201)
	if _, err = Rescale[PriceInUint64Units[DecimalPlaces2]](inexact); !errors.Is(err, ErrPrecision) {
		t.Fatalf("inexact downscale = %v", err)
	}
	overflow := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]().FromUnits(math.MaxUint64)
	if _, err = Rescale[PriceInUint64Units[DecimalPlaces3]](overflow); !errors.Is(err, ErrRange) {
		t.Fatalf("upscale overflow = %v", err)
	}
}

func TestRescaleWideSourceToNativeAcrossLargeScaleDelta(t *testing.T) {
	t.Parallel()

	factor := uint256.Int{123}
	var units uint256.Int
	units.Mul(&powersOf10Uint256[72], &factor)
	source := testFixedDecimalCodec[uint256DecimalPlaces77]().FromUnits(units)
	result, err := Rescale[PriceInUint64Units[DecimalPlaces5]](source)
	if err != nil || result.Units() != 123 {
		t.Fatalf("large-delta rescale = %d, %v", result.Units(), err)
	}
	units[0]++
	inexact := testFixedDecimalCodec[uint256DecimalPlaces77]().FromUnits(units)
	if _, err = Rescale[PriceInUint64Units[DecimalPlaces5]](inexact); !errors.Is(err, ErrPrecision) {
		t.Fatalf("large-delta inexact rescale = %v", err)
	}
}

func TestRescaleAndArithmeticAcrossEveryUnitBackend(t *testing.T) {
	t.Parallel()

	u8 := testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]]().FromUnits(12)
	u16, err := Rescale[PriceInUint16Units[DecimalPlaces2]](u8)
	if err != nil || u16.Units() != 120 {
		t.Fatalf("uint8 -> uint16 = %d, %v", u16.Units(), err)
	}
	u32, err := Rescale[PriceInUint32Units[DecimalPlaces3]](u16)
	if err != nil || u32.Units() != 1_200 {
		t.Fatalf("uint16 -> uint32 = %d, %v", u32.Units(), err)
	}
	u64, err := Rescale[PriceInUint64Units[DecimalPlaces4]](u32)
	if err != nil || u64.Units() != 12_000 {
		t.Fatalf("uint32 -> uint64 = %d, %v", u64.Units(), err)
	}
	u256, err := Rescale[PriceInUint256Units[DecimalPlaces20]](u8)
	wantWide := powersOf10Uint256[19]
	wantWide.Mul(&wantWide, &uint256.Int{12})
	if err != nil || u256.Units() != wantWide {
		t.Fatalf("uint8 -> uint256 = %#v, %v", u256.Units(), err)
	}
	round, err := Rescale[PriceInUint8Units[DecimalPlaces1]](u256)
	if err != nil || round.Units() != 12 {
		t.Fatalf("uint256 -> uint8 = %d, %v", round.Units(), err)
	}

	delta := testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]().FromUnits(3)
	sum, err := AddAs[PriceInUint32Units[DecimalPlaces2]](u8, delta)
	if err != nil || sum.Units() != 123 {
		t.Fatalf("cross-backend sum = %d, %v", sum.Units(), err)
	}
	difference, err := SubAs[PriceInUint64Units[DecimalPlaces2]](sum, delta)
	if err != nil || difference.Units() != 120 {
		t.Fatalf("cross-backend difference = %d, %v", difference.Units(), err)
	}

	tooLarge := testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]().FromUnits(256)
	if _, err = Rescale[PriceInUint8Units[DecimalPlaces2]](tooLarge); !errors.Is(err, ErrRange) {
		t.Fatalf("native target range = %v", err)
	}
}

func TestCrossScaleArithmeticMatchesBigIntReference(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(19))
	codec2 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]()
	codec3 := testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces3]]()
	for range 20_000 {
		leftUnits := rng.Uint64() % 10_000_000_000
		rightUnits := rng.Uint32()
		left := codec2.FromUnits(leftUnits)
		right := codec3.FromUnits(rightUnits)

		rescaled, err := Rescale[PriceInUint256Units[DecimalPlaces5]](left)
		wantRescaled := new(big.Int).Mul(new(big.Int).SetUint64(leftUnits), big.NewInt(1_000))
		gotRescaled := uint256ToBig(rescaled.ToU256())
		if err != nil || gotRescaled.Cmp(wantRescaled) != 0 {
			t.Fatalf("rescale(%d) = %s, %v; want %s", leftUnits, gotRescaled, err, wantRescaled)
		}

		sum, err := AddAs[PriceInUint256Units[DecimalPlaces5]](left, right)
		wantSum := new(big.Int).Add(
			wantRescaled,
			new(big.Int).Mul(new(big.Int).SetUint64(uint64(rightUnits)), big.NewInt(100)),
		)
		gotSum := uint256ToBig(sum.ToU256())
		if err != nil || gotSum.Cmp(wantSum) != 0 {
			t.Fatalf("sum(%d,%d) = %s, %v; want %s", leftUnits, rightUnits, gotSum, err, wantSum)
		}

		round, err := SubAs[PriceInUint256Units[DecimalPlaces5]](sum, right)
		gotRound := uint256ToBig(round.ToU256())
		if err != nil || gotRound.Cmp(wantRescaled) != 0 {
			t.Fatalf("round(%d,%d) = %s, %v; want %s", leftUnits, rightUnits, gotRound, err, wantRescaled)
		}
	}
}

func TestAddSubAsAcrossFractionalDecimalPlacesAndBackends(t *testing.T) {
	t.Parallel()

	left := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]().FromUnits(120)
	right := testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces3]]().FromUnits(3)
	sum, err := AddAs[PriceInUint64Units[DecimalPlaces3]](left, right)
	if err != nil || sum.Units() != 1_203 {
		t.Fatalf("sum = %d, %v", sum.Units(), err)
	}
	difference, err := SubAs[PriceInUint256Units[DecimalPlaces3]](sum, right)
	if err != nil || difference.Units() != (uint256.Int{1_200}) {
		t.Fatalf("difference = %#v, %v", difference.Units(), err)
	}
	if _, err = AddAs[PriceInUint64Units[DecimalPlaces2]](left, right); !errors.Is(err, ErrPrecision) {
		t.Fatalf("inexact target = %v", err)
	}
	if _, err = SubAs[PriceInUint64Units[DecimalPlaces3]](right, left); !errors.Is(err, ErrUnderflow) {
		t.Fatalf("underflow = %v", err)
	}
}

func TestCrossScaleOperationsDoNotAllocate(t *testing.T) {
	left := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces2]]().FromUnits(120)
	right := testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces3]]().FromUnits(3)
	assertAllocs(t, "rescale", 0, func() {
		allocationPriceSink, allocationErrorSink = Rescale[PriceInUint64Units[DecimalPlaces5]](left)
	})
	assertAllocs(t, "add as", 0, func() {
		allocationPriceSink, allocationErrorSink = AddAs[PriceInUint64Units[DecimalPlaces5]](left, right)
	})
	assertAllocs(t, "sub as", 0, func() {
		allocationPriceSink, allocationErrorSink = SubAs[PriceInUint64Units[DecimalPlaces5]](left, right)
	})
}
