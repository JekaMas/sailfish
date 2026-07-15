package sailfish

import (
	"errors"
	"testing"
)

type testAsset struct {
	Chain uint32
	Token string
}

func TestDenominatedArithmeticRequiresMatchingIdentity(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[AmountInUint64Units[DecimalPlaces6]]()
	usdc := testAsset{Chain: 1, Token: "USDC"}
	usdt := testAsset{Chain: 1, Token: "USDT"}
	left := NewDenominated(usdc, codec.FromUnits(1_000_000))
	right := NewDenominated(usdc, codec.FromUnits(250_000))

	sum, err := left.Add(right)
	if err != nil || sum.Denomination() != usdc || sum.Decimal().Units() != 1_250_000 {
		t.Fatalf("sum = %#v, %v", sum, err)
	}
	difference, err := sum.Sub(right)
	if err != nil || difference.Decimal().Units() != 1_000_000 {
		t.Fatalf("difference = %#v, %v", difference, err)
	}
	if _, err = left.Add(NewDenominated(usdt, right.Decimal())); !errors.Is(err, ErrDenominationMismatch) {
		t.Fatalf("mismatch = %v", err)
	}
	if _, err = left.Compare(NewDenominated(usdt, right.Decimal())); !errors.Is(err, ErrDenominationMismatch) {
		t.Fatalf("compare mismatch = %v", err)
	}
}

func TestDenominatedHotOperationsDoNotAllocate(t *testing.T) {
	codec := testFixedDecimalCodec[AmountInUint64Units[DecimalPlaces6]]()
	asset := uint32(7)
	left := NewDenominated(asset, codec.FromUnits(1_000_000))
	right := NewDenominated(asset, codec.FromUnits(250_000))
	var sink Denominated[uint32, AmountInUint64Units[DecimalPlaces6], uint64]
	assertAllocs(t, "denominated add", 0, func() {
		sink, allocationErrorSink = left.Add(right)
	})
	assertAllocs(t, "denominated sub", 0, func() {
		sink, allocationErrorSink = left.Sub(right)
	})
	_ = sink
}

func TestDenominatedCrossScaleArithmeticPreservesIdentity(t *testing.T) {
	t.Parallel()

	asset := testAsset{Chain: 1, Token: "USDC"}
	otherAsset := testAsset{Chain: 1, Token: "USDT"}
	left := NewDenominated(
		asset,
		testFixedDecimalCodec[AmountInUint64Units[DecimalPlaces2]]().FromUnits(120),
	)
	right := NewDenominated(
		asset,
		testFixedDecimalCodec[AmountInUint32Units[DecimalPlaces3]]().FromUnits(3),
	)

	rescaled, err := RescaleDenominated[AmountInUint64Units[DecimalPlaces5]](left)
	if err != nil || rescaled.Denomination() != asset || rescaled.Decimal().Units() != 120_000 {
		t.Fatalf("rescale = %#v, %v", rescaled, err)
	}
	sum, err := AddDenominatedAs[AmountInUint64Units[DecimalPlaces3]](left, right)
	if err != nil || sum.Denomination() != asset || sum.Decimal().Units() != 1_203 {
		t.Fatalf("sum = %#v, %v", sum, err)
	}
	difference, err := SubDenominatedAs[AmountInUint256Units[DecimalPlaces3]](sum, right)
	if err != nil || difference.Denomination() != asset || difference.Decimal().Units()[0] != 1_200 {
		t.Fatalf("difference = %#v, %v", difference, err)
	}

	mismatch := NewDenominated(otherAsset, right.Decimal())
	if _, err = AddDenominatedAs[AmountInUint64Units[DecimalPlaces3]](left, mismatch); !errors.Is(err, ErrDenominationMismatch) {
		t.Fatalf("add mismatch = %v", err)
	}
	if _, err = SubDenominatedAs[AmountInUint64Units[DecimalPlaces3]](left, mismatch); !errors.Is(err, ErrDenominationMismatch) {
		t.Fatalf("sub mismatch = %v", err)
	}
}

func TestDenominatedCrossScaleOperationsDoNotAllocate(t *testing.T) {
	asset := uint32(7)
	left := NewDenominated(
		asset,
		testFixedDecimalCodec[AmountInUint64Units[DecimalPlaces2]]().FromUnits(120),
	)
	right := NewDenominated(
		asset,
		testFixedDecimalCodec[AmountInUint32Units[DecimalPlaces3]]().FromUnits(3),
	)
	var sink Denominated[uint32, AmountInUint64Units[DecimalPlaces5], uint64]

	assertAllocs(t, "denominated rescale", 0, func() {
		sink, allocationErrorSink = RescaleDenominated[AmountInUint64Units[DecimalPlaces5]](left)
	})
	assertAllocs(t, "denominated add as", 0, func() {
		sink, allocationErrorSink = AddDenominatedAs[AmountInUint64Units[DecimalPlaces5]](left, right)
	})
	assertAllocs(t, "denominated sub as", 0, func() {
		sink, allocationErrorSink = SubDenominatedAs[AmountInUint64Units[DecimalPlaces5]](left, right)
	})
	_ = sink
}
