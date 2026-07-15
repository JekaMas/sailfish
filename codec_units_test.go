package sailfish

import (
	"testing"

	"github.com/holiman/uint256"
)

func TestFixedDecimalCodecRawUnitOperations(t *testing.T) {
	t.Parallel()

	priceCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	priceUnits, err := priceCodec.ParseUnits("123.31232")
	if err != "" {
		t.Fatalf("ParseUnits: %s", err)
	}
	if priceUnits != 12_331_232 {
		t.Fatalf("ParseUnits = %d, want 12331232", priceUnits)
	}

	priceBytesUnits, err := priceCodec.ParseUnitsBytes([]byte("123.31232"))
	if err != "" {
		t.Fatalf("ParseUnitsBytes: %s", err)
	}
	if priceBytesUnits != priceUnits {
		t.Fatalf("ParseUnitsBytes = %d, want %d", priceBytesUnits, priceUnits)
	}

	var priceBuffer [32]byte
	if got := string(priceCodec.AppendUnits(priceBuffer[:0], priceUnits)); got != "123.31232" {
		t.Fatalf("AppendUnits = %q, want %q", got, "123.31232")
	}
	if got := priceCodec.UnitsLen(priceUnits); got != len("123.31232") {
		t.Fatalf("UnitsLen = %d, want %d", got, len("123.31232"))
	}

	amountCodec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	amountUnits, err := amountCodec.ParseUnits("18446744073709551616.000000000000000001")
	if err != "" {
		t.Fatalf("wide ParseUnits: %s", err)
	}
	wantAmount := uint256.Int{1, 1_000_000_000_000_000_000}
	if amountUnits != wantAmount {
		t.Fatalf("wide ParseUnits = %v, want %v", amountUnits, wantAmount)
	}
}

func TestFixedDecimalCodecRawUnitOperationsRejectInvalidInput(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	for _, input := range []string{"", " 1.00000", "+1.00000", "1e0", "1.000000"} {
		if _, err := codec.ParseUnits(input); err == "" {
			t.Fatalf("ParseUnits(%q) unexpectedly succeeded", input)
		}
	}
	if _, err := codec.ParseUnitsBytes([]byte("1.000000")); err == "" {
		t.Fatal("ParseUnitsBytes with excess precision unexpectedly succeeded")
	}
}

func TestFixedDecimalCodecRawUnitOperationsDoNotAllocate(t *testing.T) {
	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	var buffer [32]byte
	allocs := testing.AllocsPerRun(1_000, func() {
		units, err := codec.ParseUnits("123.31232")
		if err != "" {
			t.Fatalf("ParseUnits: %s", err)
		}
		benchBytesSink = codec.AppendUnits(buffer[:0], units)
		benchUint64Sink = uint64(codec.UnitsLen(units))
	})
	if allocs != 0 {
		t.Fatalf("raw unit operations allocated %.2f times per run", allocs)
	}
}
