package sailfish

import (
	"errors"
	"math"
	"testing"

	"github.com/holiman/uint256"
)

func TestUint64Arithmetic(t *testing.T) {
	t.Parallel()

	base, _ := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("1.20000")
	delta, _ := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("0.00001")

	sum, err := base.Add(delta)
	if err != nil || sum.String() != "1.20001" {
		t.Fatalf("sum = %q, %v", sum.String(), err)
	}
	difference, err := sum.Sub(delta)
	if err != nil || !difference.Equal(base) {
		t.Fatalf("difference = %q, %v", difference.String(), err)
	}

	max, _ := NewFixedDecimalFromUnits[PriceInUint64Units[DecimalPlaces5]](uint64(math.MaxUint64))
	one := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]().FromUnits(1)
	wrapped, overflow := max.AddOverflow(one)
	if !overflow || !wrapped.IsZero() {
		t.Fatalf("wrapped=%d overflow=%v", wrapped.Units(), overflow)
	}
	before := max
	if !max.AddAssign(one) || !max.Equal(before) {
		t.Fatal("overflowing AddAssign changed receiver")
	}

	zero := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]().FromUnits(0)
	if _, err := zero.Sub(one); !errors.Is(err, ErrUnderflow) {
		t.Fatalf("underflow error = %v", err)
	}
}

func TestUint256Arithmetic(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[uint256DecimalPlaces0]()
	max := codec.FromUnits(uint256.Int{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64})
	one := codec.FromUnits(uint256.Int{1})
	wrapped, overflow := max.AddOverflow(one)
	if !overflow || !wrapped.IsZero() {
		t.Fatalf("wrapped=%#v overflow=%v", wrapped.Units(), overflow)
	}

	two, err := one.Add(one)
	if err != nil || two.Units() != (uint256.Int{2}) {
		t.Fatalf("one + one = %#v, %v", two.Units(), err)
	}
	if underflow := one.SubAssign(two); !underflow || one.Units() != (uint256.Int{1}) {
		t.Fatalf("underflow=%v receiver=%#v", underflow, one.Units())
	}
}
