package sailfish

import (
	"errors"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	json "github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

func TestExplicitUnitWidthsKeepDecimalPlacesIndependent(t *testing.T) {
	t.Parallel()

	u8, err := NewFixedDecimal[PriceInUint8Units[DecimalPlaces1]]("25.5")
	if err != nil || u8.Units() != 255 || u8.String() != "25.5" {
		t.Fatalf("uint8 price = %v %d %q", err, u8.Units(), u8.String())
	}
	if _, err = NewFixedDecimal[PriceInUint8Units[DecimalPlaces1]]("25.6"); !errors.Is(err, ErrRange) {
		t.Fatalf("uint8 range error = %v", err)
	}

	u16, err := NewFixedDecimal[PriceInUint16Units[DecimalPlaces2]]("655.35")
	if err != nil || u16.Units() != 65_535 || u16.String() != "655.35" {
		t.Fatalf("uint16 price = %v %d %q", err, u16.Units(), u16.String())
	}
	if _, err = NewFixedDecimal[PriceInUint16Units[DecimalPlaces2]]("655.36"); !errors.Is(err, ErrRange) {
		t.Fatalf("uint16 range error = %v", err)
	}

	u32, err := NewFixedDecimal[PriceInUint32Units[DecimalPlaces5]]("42949.67295")
	if err != nil || u32.Units() != 4_294_967_295 || u32.String() != "42949.67295" {
		t.Fatalf("uint32 price = %v %d %q", err, u32.Units(), u32.String())
	}
	if _, err = NewFixedDecimal[PriceInUint32Units[DecimalPlaces5]]("42949.67296"); !errors.Is(err, ErrRange) {
		t.Fatalf("uint32 range error = %v", err)
	}

	u64, err := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("184467440737095.51615")
	if err != nil || u64.Units() != ^uint64(0) {
		t.Fatalf("uint64 price = %v %d", err, u64.Units())
	}
}

func TestUnitWidthRejectsUnsupportedDecimalPlaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{name: "uint8 scale 3", err: codecError[PriceInUint8Units[DecimalPlaces3]]()},
		{name: "uint16 scale 5", err: codecError[PriceInUint16Units[DecimalPlaces5]]()},
		{name: "uint32 scale 10", err: codecError[PriceInUint32Units[DecimalPlaces10]]()},
		{name: "uint64 scale 20", err: codecError[PriceInUint64Units[DecimalPlaces20]]()},
	}
	for _, tt := range tests {
		if !errors.Is(tt.err, ErrUnsupportedFractionalDecimalPlaces) {
			t.Errorf("%s error = %v, want %v", tt.name, tt.err, ErrUnsupportedFractionalDecimalPlaces)
		}
	}
}

func TestPriceAndAmountFormatsAreDistinct(t *testing.T) {
	t.Parallel()

	priceType := reflect.TypeFor[PriceInUint256Units[DecimalPlaces18]]()
	amountType := reflect.TypeFor[AmountInUint256Units[DecimalPlaces18]]()
	if priceType == amountType {
		t.Fatalf("price and amount formats share type %v", priceType)
	}

	genericAmount, err := NewFixedDecimal[AmountInUint32Units[DecimalPlaces2]]("123.45")
	if err != nil || genericAmount.Units() != 12_345 {
		t.Fatalf("generic amount = %v %d", err, genericAmount.Units())
	}

	amount, err := NewFixedDecimal[AmountInUint256Units[DecimalPlaces18]]("1.000000000000000001")
	if err != nil || amount.Units() != (uint256.Int{1_000_000_000_000_000_001}) {
		t.Fatalf("amount = %v %#v", err, amount.Units())
	}
	acceptAmountInUint256UnitsWith18DecimalPlaces(amount)
}

func TestNarrowUnitEncodingRoundTrips(t *testing.T) {
	t.Parallel()

	type payload struct {
		Price FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16] `json:"price"`
	}
	original := payload{Price: testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]().FromUnits(65_535)}
	wire, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(wire) != `{"price":"655.35"}` {
		t.Fatalf("wire = %s", wire)
	}
	var decoded payload
	if err = json.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Price.Units() != original.Price.Units() {
		t.Fatalf("decoded units = %d, want %d", decoded.Price.Units(), original.Price.Units())
	}
}

func TestFixedDecimalCodecReportsBackendIntegerCapacity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int
		want int
	}{
		{name: "uint8 scale1", got: testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]]().MaxIntegerDigits(), want: 2},
		{name: "uint16 scale2", got: testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]().MaxIntegerDigits(), want: 3},
		{name: "uint32 scale5", got: testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces5]]().MaxIntegerDigits(), want: 5},
		{name: "uint64 scale9", got: testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]]().MaxIntegerDigits(), want: 11},
		{name: "uint256 scale18", got: testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]().MaxIntegerDigits(), want: 60},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s MaxIntegerDigits = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestNarrowUnitArithmeticOverflow(t *testing.T) {
	t.Parallel()

	assertNarrowArithmetic(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]](), uint8(255))
	assertNarrowArithmetic(t, testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]](), uint16(65_535))
	assertNarrowArithmetic(t, testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces5]](), ^uint32(0))
}

func assertNarrowArithmetic[V FixedDecimalFormat[U], U NativeUnit](t *testing.T, codec FixedDecimalCodec[V, U], maxUnits U) {
	t.Helper()

	max := codec.FromUnits(maxUnits)
	one := codec.FromUnits(U(1))
	wrapped, overflow := max.AddOverflow(one)
	if !overflow || !wrapped.IsZero() {
		t.Fatalf("wrapped=%v overflow=%v", wrapped.Units(), overflow)
	}
	if difference, underflow := wrapped.SubUnderflow(one); !underflow || difference.Units() != maxUnits {
		t.Fatalf("difference=%v underflow=%v", difference.Units(), underflow)
	}
}

func TestNarrowUnitRoundTrips(t *testing.T) {
	t.Parallel()

	for value := 0; value <= 255; value++ {
		assertUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces0]](), uint8(value))
		assertUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]](), uint8(value))
		assertUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces2]](), uint8(value))
	}

	rng := rand.New(rand.NewSource(0x51a1f15))
	for range 10_000 {
		assertUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces4]](), uint16(rng.Uint32()))
		assertUnitRoundTrip(t, testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces9]](), rng.Uint32())
	}
}

func assertUnitRoundTrip[V FixedDecimalFormat[U], U Unit](t *testing.T, codec FixedDecimalCodec[V, U], units U) {
	t.Helper()

	var buffer [maxUint256TextLen]byte
	text := codec.AppendTo(buffer[:0], codec.FromUnits(units))
	parsedString, err := codec.ParseCompact(string(text))
	if err != nil || parsedString.Units() != units {
		t.Fatalf("string round trip %q: units=%v err=%v, want %v", text, parsedString.Units(), err, units)
	}
	parsedBytes, err := codec.ParseBytes(text)
	if err != nil || parsedBytes.Units() != units {
		t.Fatalf("byte round trip %q: units=%v err=%v, want %v", text, parsedBytes.Units(), err, units)
	}
}

func TestNarrowAndWideFormatsCompareExactly(t *testing.T) {
	t.Parallel()

	narrow, err := NewFixedDecimal[PriceInUint8Units[DecimalPlaces1]]("1.2")
	if err != nil {
		t.Fatal(err)
	}
	wide, err := NewFixedDecimal[AmountInUint256Units[DecimalPlaces18]]("1.200000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if Compare(narrow, wide) != 0 || Compare(wide, narrow) != 0 {
		t.Fatal("narrow and wide equal values compare differently")
	}
}

func TestNarrowUnitsDoNotShrinkDecimalWithRetainedString(t *testing.T) {
	t.Parallel()

	if strconv.IntSize != 64 {
		t.Skip("64-bit layout assertion")
	}
	u8Size := unsafe.Sizeof(FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]{})
	u16Size := unsafe.Sizeof(FixedDecimal[PriceInUint16Units[DecimalPlaces1], uint16]{})
	u32Size := unsafe.Sizeof(FixedDecimal[PriceInUint32Units[DecimalPlaces1], uint32]{})
	u64Size := unsafe.Sizeof(FixedDecimal[PriceInUint64Units[DecimalPlaces1], uint64]{})
	if u8Size != 24 || u16Size != 24 || u32Size != 24 || u64Size != 24 {
		t.Fatalf("FixedDecimal sizes = %d/%d/%d/%d, want 24 each", u8Size, u16Size, u32Size, u64Size)
	}
	if formatSize := unsafe.Sizeof(PriceInUint16Units[DecimalPlaces2]{}); formatSize != 0 {
		t.Fatalf("generic format size = %d, want 0", formatSize)
	}
	if codecSize := unsafe.Sizeof(testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]()); codecSize != 1 {
		t.Fatalf("generic codec size = %d, want 1", codecSize)
	}
}

func codecError[V FixedDecimalFormat[U], U Unit]() error {
	_, err := NewFixedDecimalCodec[V]()
	return err
}

func acceptAmountInUint256UnitsWith18DecimalPlaces(FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]) {
}
