package sailfish

import (
	"testing"

	"github.com/holiman/uint256"
)

var (
	cborNativeSink Decimal[PriceUint64[Fraction5], uint64]
	cborWideSink   Decimal[AmountUint256[Fraction18], uint256.Int]
)

func TestCBORHotPathAllocations(t *testing.T) {
	nativeCodec := testCodec[PriceUint64[Fraction5]]()
	wideCodec := testCodec[AmountUint256[Fraction18]]()
	native := nativeCodec.FromUnits(^uint64(0))
	wide := wideCodec.FromUnits(uint256.Int{1, 2, 3, 4})
	nativeWire := native.AppendCBOR(make([]byte, 0, MaxCBORSize))
	wideWire := wide.AppendCBOR(make([]byte, 0, MaxCBORSize))
	buffer := make([]byte, 0, MaxCBORSize)

	assertAllocs(t, "append CBOR uint64", 0, func() {
		allocationBytesSink = native.AppendCBOR(buffer[:0])
	})
	assertAllocs(t, "append CBOR uint256", 0, func() {
		allocationBytesSink = wide.AppendCBOR(buffer[:0])
	})
	assertAllocs(t, "codec append CBOR uint64", 0, func() {
		allocationBytesSink = nativeCodec.AppendCBOR(buffer[:0], native)
	})
	assertAllocs(t, "unmarshal CBOR uint64", 0, func() {
		allocationErrorSink = cborNativeSink.UnmarshalCBOR(nativeWire)
	})
	assertAllocs(t, "unmarshal CBOR uint256", 0, func() {
		allocationErrorSink = cborWideSink.UnmarshalCBOR(wideWire)
	})
	assertAllocs(t, "codec parse CBOR uint64", 0, func() {
		cborNativeSink, allocationErrorSink = nativeCodec.ParseCBOR(nativeWire)
	})
	assertAllocs(t, "codec parse CBOR uint256", 0, func() {
		cborWideSink, allocationErrorSink = wideCodec.ParseCBOR(wideWire)
	})
	runtimeCodec := testUint256Codec(18)
	assertAllocs(t, "runtime codec append CBOR uint256", 0, func() {
		allocationBytesSink = runtimeCodec.AppendCBOR(buffer[:0], wide.Units())
	})
	assertAllocs(t, "runtime codec parse CBOR uint256", 0, func() {
		benchU256Sink, _ = runtimeCodec.ParseCBOR(wideWire)
	})
	assertAllocs(t, "runtime codec parse into CBOR uint256", 0, func() {
		_ = runtimeCodec.ParseCBORInto(wideWire, &benchU256Sink)
	})
	assertAllocs(t, "codec parse first CBOR uint64", 0, func() {
		cborNativeSink, allocationBytesSink, allocationErrorSink = nativeCodec.ParseCBORFirst(nativeWire)
	})
	assertAllocs(t, "codec parse first CBOR uint256", 0, func() {
		cborWideSink, allocationBytesSink, allocationErrorSink = wideCodec.ParseCBORFirst(wideWire)
	})
	assertAllocs(t, "runtime codec parse first CBOR uint256", 0, func() {
		benchU256Sink, allocationBytesSink, allocationErrorSink = runtimeCodec.ParseCBORFirst(wideWire)
	})
	assertAllocs(t, "runtime codec parse first into CBOR uint256", 0, func() {
		allocationBytesSink, allocationErrorSink = runtimeCodec.ParseCBORFirstInto(wideWire, &benchU256Sink)
	})
	assertAllocs(t, "marshal CBOR owned uint64", 1, func() {
		allocationBytesSink, allocationErrorSink = native.MarshalCBOR()
	})
	assertAllocs(t, "marshal CBOR owned uint256", 1, func() {
		allocationBytesSink, allocationErrorSink = wide.MarshalCBOR()
	})

	bar := cborBarFixture()
	barBuffer := make([]byte, 0, 93)
	barWire := appendManualCBORBarOracle(barBuffer[:0], bar)
	assertAllocs(t, "manual positional bar encode", 0, func() {
		allocationBytesSink = appendManualCBORBarOracle(barBuffer[:0], bar)
	})
	assertAllocs(t, "manual positional bar decode", 1, func() {
		cborBarOracleSink, allocationErrorSink = decodeManualCBORBarOracle(barWire)
	})
}
