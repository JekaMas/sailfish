package sailfish_test

import (
	"testing"

	"github.com/JekaMas/sailfish"
	"github.com/holiman/uint256"
)

func TestPublicAPITypeInference(t *testing.T) {
	t.Parallel()

	price, err := sailfish.New[sailfish.PriceUint64[sailfish.Fraction5]]("123.31232")
	if err != nil {
		t.Fatal(err)
	}
	acceptPriceUint64Fraction5(price)
	if price.Units() != 12_331_232 {
		t.Fatal(price.Units())
	}

	smallCodec := sailfish.MustCodec[sailfish.PriceUint16[sailfish.Fraction2]]()
	small, err := smallCodec.Parse("655.35")
	if err != nil {
		t.Fatal(err)
	}
	acceptSmallPrice(small)
	if small.Units() != 65_535 || smallCodec.MaxIntegerDigits() != 3 {
		t.Fatal(small.Units(), smallCodec.MaxIntegerDigits())
	}

	codec := sailfish.MustCodec[sailfish.AmountUint256[sailfish.Fraction18]]()
	amount, err := codec.Parse("1.000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	acceptAmountUint256Fraction18(amount)
	if amount.Units() != (uint256.Int{1_000_000_000_000_000_001}) {
		t.Fatal(amount.Units())
	}

	runtimeCodec := sailfish.MustUint256Codec(6)
	var runtimeUnits uint256.Int
	if parseErr := runtimeCodec.ParseInto("123.456789", &runtimeUnits); parseErr != "" {
		t.Fatal(parseErr)
	}
	if runtimeUnits != (uint256.Int{123_456_789}) {
		t.Fatal(runtimeUnits)
	}
}

func acceptPriceUint64Fraction5(sailfish.Decimal[sailfish.PriceUint64[sailfish.Fraction5], uint64]) {}

func acceptSmallPrice(sailfish.Decimal[sailfish.PriceUint16[sailfish.Fraction2], uint16]) {}

func acceptAmountUint256Fraction18(
	sailfish.Decimal[sailfish.AmountUint256[sailfish.Fraction18], uint256.Int],
) {
}
