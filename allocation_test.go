package sailfish

import (
	"testing"

	"github.com/holiman/uint256"
)

var (
	allocationPriceSink  price5
	allocationUint8Sink  FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]
	allocationUint16Sink FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]
	allocationUint32Sink FixedDecimal[PriceInUint32Units[DecimalPlaces5], uint32]
	allocationWideSink   wide18
	allocationBytesSink  []byte
	allocationStringSink string
	allocationIntSink    int
	allocationBoolSink   bool
	allocationErrorSink  error
)

func TestHotPathAllocations(t *testing.T) {
	priceCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	wideCodec := testFixedDecimalCodec[uint256DecimalPlaces18]()
	priceBytes := []byte("123.31232")
	appendBuffer := make([]byte, 0, 96)
	priceRetained, _ := priceCodec.Parse("123.31232")
	priceFromUnits := priceCodec.FromUnits(12_331_232)
	priceDelta := priceCodec.FromUnits(1)
	wideFromUnits := wideCodec.FromUnits(uint256.Int{1, 2, 3, 4})
	otherScale := testFixedDecimalCodec[uint256DecimalPlaces37]().FromUnits(uint256.Int{5, 6, 7, 8})
	uint8Codec := testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces1]]()
	uint16Codec := testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]()
	uint32Codec := testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces5]]()

	assertAllocs(t, "parse canonical uint64", 0, func() {
		allocationPriceSink, _ = priceCodec.Parse("123.31232")
	})
	assertAllocs(t, "parse compact uint64", 0, func() {
		allocationPriceSink, _ = priceCodec.ParseCompact("123.31232")
	})
	assertAllocs(t, "parse bytes uint64", 0, func() {
		allocationPriceSink, _ = priceCodec.ParseBytes(priceBytes)
	})
	assertAllocs(t, "parse compact uint8", 0, func() {
		allocationUint8Sink, _ = uint8Codec.ParseCompact("25.5")
	})
	assertAllocs(t, "parse compact uint16", 0, func() {
		allocationUint16Sink, _ = uint16Codec.ParseCompact("655.35")
	})
	assertAllocs(t, "parse compact uint32", 0, func() {
		allocationUint32Sink, _ = uint32Codec.ParseCompact("42949.67295")
	})
	assertAllocs(t, "reject invalid uint64", 0, func() {
		allocationPriceSink, allocationErrorSink = priceCodec.ParseCompact("123x31232")
	})
	assertAllocs(t, "reject precision uint64", 0, func() {
		allocationPriceSink, allocationErrorSink = priceCodec.ParseCompact("1.123456")
	})
	assertAllocs(t, "reject range uint64", 0, func() {
		allocationPriceSink, allocationErrorSink = priceCodec.ParseCompact("184467440737095.51616")
	})
	assertAllocs(t, "reject scale", 0, func() {
		_, allocationErrorSink = NewFixedDecimalCodec[uint64DecimalPlaces20]()
	})
	assertAllocs(t, "append uint64", 0, func() {
		allocationBytesSink = priceCodec.AppendTo(appendBuffer[:0], priceFromUnits)
	})
	uint8Value := uint8Codec.FromUnits(255)
	uint16Value := uint16Codec.FromUnits(65_535)
	uint32Value := uint32Codec.FromUnits(4_294_967_295)
	assertAllocs(t, "append uint8", 0, func() {
		allocationBytesSink = uint8Codec.AppendTo(appendBuffer[:0], uint8Value)
	})
	assertAllocs(t, "append uint16", 0, func() {
		allocationBytesSink = uint16Codec.AppendTo(appendBuffer[:0], uint16Value)
	})
	assertAllocs(t, "append uint32", 0, func() {
		allocationBytesSink = uint32Codec.AppendTo(appendBuffer[:0], uint32Value)
	})
	assertAllocs(t, "retained string", 0, func() {
		allocationStringSink = priceRetained.String()
	})
	assertAllocs(t, "add assign uint64", 0, func() {
		value := priceFromUnits
		allocationBoolSink = value.AddAssign(priceDelta)
		allocationPriceSink = value
	})
	assertAllocs(t, "parse uint256", 0, func() {
		allocationWideSink, _ = wideCodec.Parse("12345678901234567890.123456789012345678")
	})
	assertAllocs(t, "append uint256", 0, func() {
		allocationBytesSink = wideCodec.AppendTo(appendBuffer[:0], wideFromUnits)
	})
	assertAllocs(t, "cross-scale compare", 0, func() {
		allocationIntSink = Compare(wideFromUnits, otherScale)
	})
	assertAllocs(t, "formatted string", 1, func() {
		allocationStringSink = priceFromUnits.String()
	})
	maxPrice := priceCodec.FromUnits(^uint64(0))
	assertAllocs(t, "addition overflow error", 0, func() {
		allocationPriceSink, allocationErrorSink = maxPrice.Add(priceDelta)
	})
	zeroPrice := priceCodec.FromUnits(0)
	assertAllocs(t, "subtraction underflow error", 0, func() {
		allocationPriceSink, allocationErrorSink = zeroPrice.Sub(priceDelta)
	})
}

func assertAllocs(t *testing.T, name string, want float64, fn func()) {
	t.Helper()
	got := testing.AllocsPerRun(1_000, fn)
	if got != want {
		t.Errorf("%s allocations = %.2f, want %.2f", name, got, want)
	}
}
