package sailfish

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/holiman/uint256"
)

func TestUint256MulSmallAddMatchesBigInt(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(3))
	multipliers := []uint64{1, 10, 100, 1_000_000, 1_000_000_000_000_000_000, uint256ChunkBase}
	modulus := new(big.Int).Lsh(big.NewInt(1), 256)

	for range 20_000 {
		value := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		multiplier := multipliers[rng.Intn(len(multipliers))]
		add := rng.Uint64()
		if multiplier != 0 {
			add %= multiplier
		}

		got, overflow := uint256MulSmallAdd(value, multiplier, add)
		wantBig := new(big.Int).Mul(uint256ToBig(value), new(big.Int).SetUint64(multiplier))
		wantBig.Add(wantBig, new(big.Int).SetUint64(add))
		wantOverflow := wantBig.BitLen() > 256
		wantBig.Mod(wantBig, modulus)
		want, conversionOverflow := uint256.FromBig(wantBig)
		if conversionOverflow {
			t.Fatal("modulo result does not fit uint256")
		}
		if overflow != wantOverflow || got != *want {
			t.Fatalf(
				"mul-add(%#v, %d, %d) = %#v/%v, want %#v/%v",
				value, multiplier, add, got, overflow, *want, wantOverflow,
			)
		}
	}
}

func TestUint256DivMod64MatchesBigInt(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(4))
	for range 20_000 {
		value := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		divisor := rng.Uint64()
		if divisor == 0 {
			divisor = 1
		}

		gotQuotient, gotRemainder := uint256DivMod64(value, divisor)
		wantQuotient, wantRemainder := new(big.Int).QuoRem(
			uint256ToBig(value),
			new(big.Int).SetUint64(divisor),
			new(big.Int),
		)
		want, overflow := uint256.FromBig(wantQuotient)
		if overflow {
			t.Fatal("quotient does not fit uint256")
		}
		if gotQuotient != *want || gotRemainder != wantRemainder.Uint64() {
			t.Fatalf(
				"divmod(%#v, %d) = %#v/%d, want %#v/%d",
				value, divisor, gotQuotient, gotRemainder, *want, wantRemainder.Uint64(),
			)
		}
	}
}

func TestUint256AddSubMatchesBigInt(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(5))
	modulus := new(big.Int).Lsh(big.NewInt(1), 256)
	for range 20_000 {
		a := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		b := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		left := MustCodec[uint256Scale0]().FromUnits(a)
		right := MustCodec[uint256Scale0]().FromUnits(b)

		gotSum, overflow := left.AddOverflow(right)
		wantSumBig := new(big.Int).Add(uint256ToBig(a), uint256ToBig(b))
		wantOverflow := wantSumBig.BitLen() > 256
		wantSumBig.Mod(wantSumBig, modulus)
		wantSum, _ := uint256.FromBig(wantSumBig)
		if overflow != wantOverflow || gotSum.Units() != *wantSum {
			t.Fatalf("add %#v + %#v = %#v/%v, want %#v/%v", a, b, gotSum.Units(), overflow, *wantSum, wantOverflow)
		}

		gotDifference, underflow := left.SubUnderflow(right)
		wantDifferenceBig := new(big.Int).Sub(uint256ToBig(a), uint256ToBig(b))
		wantUnderflow := wantDifferenceBig.Sign() < 0
		wantDifferenceBig.Mod(wantDifferenceBig, modulus)
		wantDifference, _ := uint256.FromBig(wantDifferenceBig)
		if underflow != wantUnderflow || gotDifference.Units() != *wantDifference {
			t.Fatalf("sub %#v - %#v = %#v/%v, want %#v/%v", a, b, gotDifference.Units(), underflow, *wantDifference, wantUnderflow)
		}
	}
}
