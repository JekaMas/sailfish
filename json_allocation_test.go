package sailfish

import (
	"testing"

	"github.com/holiman/uint256"
)

func TestJSONHotPathAllocations(t *testing.T) {
	nativeCodec := testCodec[PriceUint64[Fraction5]]()
	native := nativeCodec.FromUnits(1_230_000)
	nativeWire := []byte(`"12.30000"`)
	nativeBuffer := make([]byte, 0, len(nativeWire))

	wideCodec := testCodec[uint256Scale18]()
	wide := wideCodec.FromUnits(uint256.Int{
		^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0),
	})
	wideWire := []byte(
		`"115792089237316195423570985008687907853269984665640564039457.584007913129639935"`,
	)
	wideBuffer := make([]byte, 0, len(wideWire))

	assertAllocs(t, "append JSON native", 0, func() {
		allocationBytesSink = native.AppendJSON(nativeBuffer[:0])
	})
	assertAllocs(t, "append JSON uint256", 0, func() {
		allocationBytesSink = wide.AppendJSON(wideBuffer[:0])
	})
	assertAllocs(t, "marshal JSON native owned result", 1, func() {
		allocationBytesSink, allocationErrorSink = native.MarshalJSON()
	})
	assertAllocs(t, "marshal JSON uint256 owned result", 1, func() {
		allocationBytesSink, allocationErrorSink = wide.MarshalJSON()
	})
	assertAllocs(t, "unmarshal JSON native", 0, func() {
		allocationErrorSink = benchPriceSink.UnmarshalJSON(nativeWire)
	})
	assertAllocs(t, "unmarshal JSON uint256", 0, func() {
		allocationErrorSink = benchWideSink.UnmarshalJSON(wideWire)
	})
}
