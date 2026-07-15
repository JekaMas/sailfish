package sailfish

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/holiman/uint256"
)

func TestSameVenueCompare(t *testing.T) {
	t.Parallel()

	a, _ := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("1.20000")
	b, _ := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("1.20001")
	if a.Compare(b) != -1 || b.Cmp(a) != 1 || !a.Less(b) || a.Equal(b) {
		t.Fatal("same-venue comparison contract failed")
	}
}

func TestCrossScaleAndBackendCompare(t *testing.T) {
	t.Parallel()

	a, _ := NewFixedDecimal[PriceInUint64Units[DecimalPlaces2]]("1.20")
	b, _ := NewFixedDecimal[uint256DecimalPlaces18]("1.200000000000000000")
	if Compare(a, b) != 0 || Compare(b, a) != 0 {
		t.Fatal("cross-backend equal values differ")
	}

	less, _ := NewFixedDecimal[PriceInUint64Units[DecimalPlaces2]]("0.01")
	more, _ := NewFixedDecimal[uint256DecimalPlaces18]("0.010000000000000001")
	if Compare(less, more) != -1 || Compare(more, less) != 1 {
		t.Fatal("cross-backend ordering failed")
	}
}

func TestCrossScaleCompareMatchesBigIntReference(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(0x5a11f15))
	for range 10_000 {
		aUnits := rng.Uint64()
		bUnits := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		a, _ := NewFixedDecimalFromUnits[uint64DecimalPlaces19](aUnits)
		b, _ := NewFixedDecimalFromUnits[uint256DecimalPlaces37](bUnits)

		got := Compare(a, b)
		want := compareScaledBig(
			new(big.Int).SetUint64(aUnits),
			19,
			uint256ToBig(bUnits),
			37,
		)
		if got != want {
			t.Fatalf("Compare(%d, %#v) = %d, want %d", aUnits, bUnits, got, want)
		}
	}
}

func compareScaledBig(a *big.Int, aScale int, b *big.Int, bScale int) int {
	left := new(big.Int).Mul(new(big.Int).Set(a), pow10Big(bScale))
	right := new(big.Int).Mul(new(big.Int).Set(b), pow10Big(aScale))
	return left.Cmp(right)
}

func pow10Big(scale int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
}

func uint256ToBig(value uint256.Int) *big.Int {
	var out big.Int
	for i := 3; i >= 0; i-- {
		out.Lsh(&out, 64)
		if value[i] != 0 {
			out.Add(&out, new(big.Int).SetUint64(value[i]))
		}
	}
	return &out
}
