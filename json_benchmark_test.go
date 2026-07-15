package sailfish

import (
	"testing"

	json "github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

var jsonErrorSink error

// BenchmarkJSONHotPaths separates caller-buffer encoding, required owned
// output, direct decoding, and reflective/interface integration.
func BenchmarkJSONHotPaths(b *testing.B) {
	nativeCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	nativeRetained, _ := nativeCodec.Parse("12.30000")
	nativeFormatted := nativeCodec.FromUnits(1_230_000)

	wideCodec := testFixedDecimalCodec[uint256DecimalPlaces18]()
	wideUnits := uint256.Int{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}
	wideFormatted := wideCodec.FromUnits(wideUnits)
	wideRetained, _ := wideCodec.Parse(
		"115792089237316195423570985008687907853269984665640564039457.584007913129639935",
	)

	nativeWire := []byte(`"12.30000"`)
	escapedNativeWire := []byte(`"\u0031\u0032.30000"`)
	wideWire := []byte(
		`"115792089237316195423570985008687907853269984665640564039457.584007913129639935"`,
	)
	nativeBuffer := make([]byte, 0, len(nativeWire))
	wideBuffer := make([]byte, 0, len(wideWire))

	benchmarkJSONAppend(b, "native_retained", nativeRetained, nativeBuffer)
	benchmarkJSONAppend(b, "native_formatted", nativeFormatted, nativeBuffer)
	benchmarkJSONAppend(b, "wide_retained", wideRetained, wideBuffer)
	benchmarkJSONAppend(b, "wide_formatted", wideFormatted, wideBuffer)

	benchmarkJSONMarshal(b, "native_retained", nativeRetained)
	benchmarkJSONMarshal(b, "native_formatted", nativeFormatted)
	benchmarkJSONMarshal(b, "wide_retained", wideRetained)
	benchmarkJSONMarshal(b, "wide_formatted", wideFormatted)

	b.Run("marshal_go_json/native_retained", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, jsonErrorSink = json.Marshal(nativeRetained)
		}
	})
	b.Run("marshal_go_json/wide_formatted", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, jsonErrorSink = json.Marshal(wideFormatted)
		}
	})

	benchmarkJSONUnmarshal(b, "native_canonical", nativeWire, &benchPriceSink)
	benchmarkJSONUnmarshal(b, "native_escaped", escapedNativeWire, &benchPriceSink)
	benchmarkJSONUnmarshal(b, "wide_canonical", wideWire, &benchWideSink)

	b.Run("unmarshal_go_json/native_canonical", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			jsonErrorSink = json.Unmarshal(nativeWire, &benchPriceSink)
		}
	})
	b.Run("unmarshal_go_json/wide_canonical", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			jsonErrorSink = json.Unmarshal(wideWire, &benchWideSink)
		}
	})
}

func benchmarkJSONAppend[V FixedDecimalFormat[U], U Unit](
	b *testing.B,
	name string,
	value FixedDecimal[V, U],
	buffer []byte,
) {
	b.Run("append/"+name, func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = value.AppendJSON(buffer[:0])
		}
	})
}

func benchmarkJSONMarshal[V FixedDecimalFormat[U], U Unit](
	b *testing.B,
	name string,
	value FixedDecimal[V, U],
) {
	b.Run("marshal_direct/"+name, func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, jsonErrorSink = value.MarshalJSON()
		}
	})
}

func benchmarkJSONUnmarshal[V FixedDecimalFormat[U], U Unit](
	b *testing.B,
	name string,
	wire []byte,
	dst *FixedDecimal[V, U],
) {
	b.Run("unmarshal_direct/"+name, func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			jsonErrorSink = dst.UnmarshalJSON(wire)
		}
	})
}
