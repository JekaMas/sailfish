package sailfish_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/JekaMas/sailfish"
	"github.com/fxamacker/cbor/v2"
	"github.com/holiman/uint256"
)

func TestPublicAPITypeNamesExposeRepresentationAndDecimalPlaces(t *testing.T) {
	t.Parallel()

	price, err := sailfish.NewFixedDecimal[sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]]("123.31232")
	if err != nil {
		t.Fatal(err)
	}
	acceptPriceInUint64UnitsWith5DecimalPlaces(price)
	if price.Units() != 12_331_232 {
		t.Fatal(price.Units())
	}

	smallCodec := requireFixedDecimalCodec[sailfish.PriceInUint16Units[sailfish.DecimalPlaces2]](t)
	if smallCodec.FractionalDecimalPlaces() != 2 {
		t.Fatal(smallCodec.FractionalDecimalPlaces())
	}
	small, err := smallCodec.Parse("655.35")
	if err != nil {
		t.Fatal(err)
	}
	acceptSmallPrice(small)
	if small.Units() != 65_535 || smallCodec.MaxIntegerDigits() != 3 {
		t.Fatal(small.Units(), smallCodec.MaxIntegerDigits())
	}

	codec := requireFixedDecimalCodec[sailfish.AmountInUint256Units[sailfish.DecimalPlaces18]](t)
	amount, err := codec.Parse("1.000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	acceptAmountInUint256UnitsWith18DecimalPlaces(amount)
	if amount.Units() != (uint256.Int{1_000_000_000_000_000_001}) {
		t.Fatal(amount.Units())
	}

	runtimeCodec := requireUint256FixedDecimalCodec(t, 6)
	var runtimeUnits uint256.Int
	if parseErr := runtimeCodec.ParseInto("123.456789", &runtimeUnits); parseErr != "" {
		t.Fatal(parseErr)
	}
	if runtimeUnits != (uint256.Int{123_456_789}) {
		t.Fatal(runtimeUnits)
	}
}

func TestPublicIntegerConversionAPI(t *testing.T) {
	t.Parallel()

	codec := requireFixedDecimalCodec[sailfish.AmountInUint256Units[sailfish.DecimalPlaces18]](t)
	source := new(big.Int).Lsh(big.NewInt(1), 192)
	value, err := codec.FromBigInt(source)
	if err != nil {
		t.Fatal(err)
	}
	units := value.ToU256()
	fromUnits, err := codec.FromU256(units)
	if err != nil {
		t.Fatal(err)
	}
	var destination big.Int
	if err = fromUnits.ToBigInt(&destination); err != nil || destination.Cmp(source) != 0 {
		t.Fatalf("round trip = %s, %v", destination.String(), err)
	}
}

func TestPublicCBORToArrayAPI(t *testing.T) {
	t.Parallel()

	type quote struct {
		_ struct{} `cbor:",toarray"`

		Price sailfish.FixedDecimal[sailfish.PriceInUint64Units[sailfish.DecimalPlaces5], uint64]
	}

	codec := requireFixedDecimalCodec[sailfish.PriceInUint64Units[sailfish.DecimalPlaces5]](t)
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

func acceptPriceInUint64UnitsWith5DecimalPlaces(
	sailfish.FixedDecimal[sailfish.PriceInUint64Units[sailfish.DecimalPlaces5], uint64],
) {
}

func acceptSmallPrice(sailfish.FixedDecimal[sailfish.PriceInUint16Units[sailfish.DecimalPlaces2], uint16]) {
}

func acceptAmountInUint256UnitsWith18DecimalPlaces(
	sailfish.FixedDecimal[sailfish.AmountInUint256Units[sailfish.DecimalPlaces18], uint256.Int],
) {
}

func requireFixedDecimalCodec[V sailfish.FixedDecimalFormat[U], U sailfish.Unit](t *testing.T) sailfish.FixedDecimalCodec[V, U] {
	t.Helper()
	codec, err := sailfish.NewFixedDecimalCodec[V]()
	if err != nil {
		t.Fatal(err)
	}
	return codec
}

func requireUint256FixedDecimalCodec(
	t *testing.T,
	fractionalDecimalPlaces sailfish.DecimalPlaces,
) sailfish.Uint256FixedDecimalCodec {
	t.Helper()
	codec, err := sailfish.NewUint256FixedDecimalCodec(fractionalDecimalPlaces)
	if err != nil {
		t.Fatal(err)
	}
	return codec
}
