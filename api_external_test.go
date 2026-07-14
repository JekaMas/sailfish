package sailfish_test

import (
	"encoding/hex"
	"testing"

	"github.com/JekaMas/sailfish"
	"github.com/fxamacker/cbor/v2"
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

	smallCodec := requireCodec[sailfish.PriceUint16[sailfish.Fraction2]](t)
	small, err := smallCodec.Parse("655.35")
	if err != nil {
		t.Fatal(err)
	}
	acceptSmallPrice(small)
	if small.Units() != 65_535 || smallCodec.MaxIntegerDigits() != 3 {
		t.Fatal(small.Units(), smallCodec.MaxIntegerDigits())
	}

	codec := requireCodec[sailfish.AmountUint256[sailfish.Fraction18]](t)
	amount, err := codec.Parse("1.000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	acceptAmountUint256Fraction18(amount)
	if amount.Units() != (uint256.Int{1_000_000_000_000_000_001}) {
		t.Fatal(amount.Units())
	}

	runtimeCodec := requireUint256Codec(t, 6)
	var runtimeUnits uint256.Int
	if parseErr := runtimeCodec.ParseInto("123.456789", &runtimeUnits); parseErr != "" {
		t.Fatal(parseErr)
	}
	if runtimeUnits != (uint256.Int{123_456_789}) {
		t.Fatal(runtimeUnits)
	}
}

func TestPublicCBORToArrayAPI(t *testing.T) {
	t.Parallel()

	type quote struct {
		_ struct{} `cbor:",toarray"`

		Price sailfish.Decimal[sailfish.PriceUint64[sailfish.Fraction5], uint64]
	}

	codec := requireCodec[sailfish.PriceUint64[sailfish.Fraction5]](t)
	original := quote{Price: codec.FromUnits(12_331_232)}
	wire, err := cbor.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(wire); got != "811a00bc28e0" {
		t.Fatalf("wire = %s", got)
	}
	var decoded quote
	if err = cbor.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Price.Equal(original.Price) {
		t.Fatal(decoded.Price.Units())
	}

	first, rest, err := codec.ParseCBORFirst(wire[1:])
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 0 || !first.Equal(original.Price) {
		t.Fatalf("manual decode = %d, rest %x", first.Units(), rest)
	}
}

func acceptPriceUint64Fraction5(sailfish.Decimal[sailfish.PriceUint64[sailfish.Fraction5], uint64]) {}

func acceptSmallPrice(sailfish.Decimal[sailfish.PriceUint16[sailfish.Fraction2], uint16]) {}

func acceptAmountUint256Fraction18(
	sailfish.Decimal[sailfish.AmountUint256[sailfish.Fraction18], uint256.Int],
) {
}

func requireCodec[V sailfish.Venue[U], U sailfish.Unit](t *testing.T) sailfish.Codec[V, U] {
	t.Helper()
	codec, err := sailfish.NewCodec[V]()
	if err != nil {
		t.Fatal(err)
	}
	return codec
}

func requireUint256Codec(t *testing.T, scale sailfish.Notion) sailfish.Uint256Codec {
	t.Helper()
	codec, err := sailfish.NewUint256Codec(scale)
	if err != nil {
		t.Fatal(err)
	}
	return codec
}
