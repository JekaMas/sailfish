package sailfish

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/holiman/uint256"
)

func BenchmarkCBOR(b *testing.B) {
	nativeCodec := MustCodec[PriceUint64[Fraction5]]()
	wideCodec := MustCodec[AmountUint256[Fraction18]]()
	native := nativeCodec.FromUnits(^uint64(0))
	wide := wideCodec.FromUnits(uint256.Int{1, 2, 3, 4})
	nativeWire := native.AppendCBOR(make([]byte, 0, MaxCBORSize))
	wideWire := wide.AppendCBOR(make([]byte, 0, MaxCBORSize))
	buffer := make([]byte, 0, MaxCBORSize)

	b.Run("append/uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = native.AppendCBOR(buffer[:0])
		}
	})
	b.Run("append/uint256", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = wide.AppendCBOR(buffer[:0])
		}
	})
	b.Run("decode/uint64", func(b *testing.B) {
		var value Decimal[PriceUint64[Fraction5], uint64]
		b.ReportAllocs()
		for b.Loop() {
			_ = value.UnmarshalCBOR(nativeWire)
		}
		cborNativeSink = value
	})
	b.Run("decode/uint256", func(b *testing.B) {
		var value Decimal[AmountUint256[Fraction18], uint256.Int]
		b.ReportAllocs()
		for b.Loop() {
			_ = value.UnmarshalCBOR(wideWire)
		}
		cborWideSink = value
	})
	b.Run("marshal_owned/uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, _ = native.MarshalCBOR()
		}
	})
	b.Run("marshal_owned/uint256", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, _ = wide.MarshalCBOR()
		}
	})
}

func BenchmarkCBORToArrayIntegration(b *testing.B) {
	price := MustCodec[PriceUint64[Fraction5]]().FromUnits(12_331_232)
	amount := MustCodec[AmountUint256[Fraction18]]().FromUnits(uint256.Int{0, 1})
	value := cborQuote{Price: price, Amount: amount}
	wire, err := cbor.Marshal(value)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("marshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, _ = cbor.Marshal(value)
		}
	})
	b.Run("unmarshal", func(b *testing.B) {
		var decoded cborQuote
		b.ReportAllocs()
		for b.Loop() {
			_ = cbor.Unmarshal(wire, &decoded)
		}
		cborNativeSink = decoded.Price
		cborWideSink = decoded.Amount
	})
}

func BenchmarkCBORDispatchLayers(b *testing.B) {
	nativeCodec := MustCodec[PriceUint64[Fraction5]]()
	native := nativeCodec.FromUnits(^uint64(0))
	nativeWire := native.AppendCBOR(make([]byte, 0, MaxCBORSize))
	buffer := make([]byte, 0, MaxCBORSize)

	b.Run("append/helper_uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = appendCBORUint64(buffer[:0], ^uint64(0))
		}
	})
	b.Run("append/provider_uint64", func(b *testing.B) {
		provider := Uint64Units{}
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = provider.unitAppendCBOR(buffer[:0], ^uint64(0))
		}
	})
	b.Run("append/codec_uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = nativeCodec.AppendCBOR(buffer[:0], native)
		}
	})
	b.Run("append/decimal_uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = native.AppendCBOR(buffer[:0])
		}
	})

	b.Run("decode/helper_uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _ = parseCBORUint64(nativeWire, ^uint64(0))
		}
	})
	b.Run("decode/provider_uint64", func(b *testing.B) {
		provider := Uint64Units{}
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _ = provider.unitParseCBOR(nativeWire)
		}
	})
	b.Run("decode/codec_uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cborNativeSink, _ = nativeCodec.ParseCBOR(nativeWire)
		}
	})
	b.Run("decode/decimal_uint64", func(b *testing.B) {
		var value Decimal[PriceUint64[Fraction5], uint64]
		b.ReportAllocs()
		for b.Loop() {
			_ = value.UnmarshalCBOR(nativeWire)
		}
		cborNativeSink = value
	})
}

func BenchmarkCBORUint256Widths(b *testing.B) {
	codec := MustCodec[AmountUint256[Fraction18]]()
	runtimeCodec := MustUint256Codec(18)
	buffer := make([]byte, 0, MaxCBORSize)
	values := []struct {
		name  string
		units uint256.Int
	}{
		{name: "one_limb", units: uint256.Int{^uint64(0)}},
		{name: "two_limbs", units: uint256.Int{1, 1}},
		{name: "four_limbs", units: uint256.Int{1, 2, 3, 4}},
		{name: "maximum", units: uint256.Int{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
	}
	for _, tt := range values {
		value := codec.FromUnits(tt.units)
		wire := value.AppendCBOR(buffer[:0])
		b.Run("append/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink = value.AppendCBOR(buffer[:0])
			}
		})
		b.Run("decode/"+tt.name, func(b *testing.B) {
			var decoded Decimal[AmountUint256[Fraction18], uint256.Int]
			b.ReportAllocs()
			for b.Loop() {
				_ = decoded.UnmarshalCBOR(wire)
			}
			cborWideSink = decoded
		})
		b.Run("codec_append/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink = codec.AppendCBOR(buffer[:0], value)
			}
		})
		b.Run("codec_decode/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				cborWideSink, _ = codec.ParseCBOR(wire)
			}
		})
		b.Run("runtime_codec_append/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink = runtimeCodec.AppendCBOR(buffer[:0], tt.units)
			}
		})
		b.Run("runtime_codec_decode/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchU256Sink, _ = runtimeCodec.ParseCBOR(wire)
			}
		})
		b.Run("runtime_codec_decode_first/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchU256Sink, benchBytesSink, _ = runtimeCodec.ParseCBORFirst(wire)
			}
		})
		b.Run("runtime_codec_decode_first_into/"+tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink, _ = runtimeCodec.ParseCBORFirstInto(wire, &benchU256Sink)
			}
		})
	}
}

func BenchmarkCBORManualPositionalBar(b *testing.B) {
	value := cborBarFixture()
	buffer := make([]byte, 0, 93)
	wire := appendManualCBORBarOracle(buffer[:0], value)
	b.ReportMetric(float64(len(wire)), "B/wire")

	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ReportMetric(float64(len(wire)), "B/wire")
		for b.Loop() {
			benchBytesSink = appendManualCBORBarOracle(buffer[:0], value)
		}
	})
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		b.ReportMetric(float64(len(wire)), "B/wire")
		for b.Loop() {
			cborBarOracleSink, _ = decodeManualCBORBarOracle(wire)
		}
	})
}

var cborBarOracleSink cborBarOracle
