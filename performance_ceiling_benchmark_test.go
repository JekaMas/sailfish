package sailfish

import (
	"math/bits"
	"testing"

	"github.com/holiman/uint256"
)

// BenchmarkPerformanceCeilings pairs each public hot operation with the
// narrowest internal kernel that performs equivalent numeric or ownership
// work. The kernel is a measured implementation ceiling, not a claim about
// nominal hardware cycles. Public-path optimization targets are derived from
// these same-binary pairs.
func BenchmarkPerformanceCeilings(b *testing.B) {
	benchmarkParseCeilings(b)
	benchmarkFormatCeilings(b)
	benchmarkRetainedCeilings(b)
	benchmarkCompareCeilings(b)
	benchmarkArithmeticCeilings(b)
	benchmarkCBORCeilings(b)
}

func benchmarkParseCeilings(b *testing.B) {
	const input = "123.456789"
	inputBytes := []byte(input)
	codec := testUint256FixedDecimalCodec(6)

	b.Run("parse/runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _ = codec.Parse(input)
		}
	})
	b.Run("parse/runtime_codec_into", func(b *testing.B) {
		var value uint256.Int
		b.ReportAllocs()
		for b.Loop() {
			_ = codec.ParseInto(input, &value)
		}
		benchU256Sink = value
	})
	b.Run("parse/runtime_codec_bytes", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _ = codec.ParseBytes(inputBytes)
		}
	})
	b.Run("parse/runtime_codec_bytes_into", func(b *testing.B) {
		var value uint256.Int
		b.ReportAllocs()
		for b.Loop() {
			_ = codec.ParseBytesInto(inputBytes, &value)
		}
		benchU256Sink = value
	})
	b.Run("parse/complete_uint256_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _, _ = parseUint256(input, 6)
		}
	})
	b.Run("parse/canonical_digit_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _ = parseUint64WithDot(input, 3)
		}
	})
}

func benchmarkFormatCeilings(b *testing.B) {
	codec := testUint256FixedDecimalCodec(6)
	units := uint256.Int{123_456_789}
	buffer := make([]byte, 0, 32)

	b.Run("format/runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], units)
		}
	})
	b.Run("format/complete_uint64_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = appendUint64Decimal(buffer[:0], units[0], 6)
		}
	})
	b.Run("format/fixed_digit_kernel", func(b *testing.B) {
		out := buffer[:10]
		b.ReportAllocs()
		for b.Loop() {
			fillFixed64(out, units[0])
		}
		benchBytesSink = out
	})
	b.Run("format/runtime_codec_len", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = codec.Len(units)
		}
	})
	b.Run("format/complete_uint256_len", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = (Uint256Units{}).unitLen(units, 6)
		}
	})
}

func benchmarkRetainedCeilings(b *testing.B) {
	nativeCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	native, _ := nativeCodec.Parse("123.31232")
	wideCodec := testFixedDecimalCodec[uint256DecimalPlaces18]()
	wide, _ := wideCodec.Parse("115792089237316195423570985008687907853269984665640564039.457584007913129639")
	buffer := make([]byte, 0, 96)

	b.Run("retained/native_public", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = nativeCodec.AppendTo(buffer[:0], native)
		}
	})
	b.Run("retained/native_copy_ceiling", func(b *testing.B) {
		text := native.representation
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = append(buffer[:0], text...)
		}
	})
	b.Run("retained/wide_public", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = wideCodec.AppendTo(buffer[:0], wide)
		}
	})
	b.Run("retained/wide_copy_ceiling", func(b *testing.B) {
		text := wide.representation
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = append(buffer[:0], text...)
		}
	})
}

func benchmarkCompareCeilings(b *testing.B) {
	nativeCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	nativeA := nativeCodec.FromUnits(12_331_232)
	nativeB := nativeCodec.FromUnits(12_331_233)
	wideCodec := testFixedDecimalCodec[uint256DecimalPlaces18]()
	wideA := wideCodec.FromUnits(uint256.Int{1, 2, 3, 4})
	wideB := wideCodec.FromUnits(uint256.Int{2, 2, 3, 4})

	b.Run("compare/native_public", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = nativeA.Compare(nativeB)
		}
	})
	b.Run("compare/native_integer_ceiling", func(b *testing.B) {
		a, c := nativeA.units, nativeB.units
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = compareUint64Ceiling(a, c)
		}
	})
	b.Run("compare/wide_public", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = wideA.Compare(wideB)
		}
	})
	b.Run("compare/wide_integer_ceiling", func(b *testing.B) {
		a, c := wideA.units, wideB.units
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = compareUint256Ceiling(a, c)
		}
	})
}

func compareUint64Ceiling(a, b uint64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareUint256Ceiling(a, b uint256.Int) int {
	for i := 3; i >= 0; i-- {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func benchmarkArithmeticCeilings(b *testing.B) {
	nativeCodec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	nativeBase := nativeCodec.FromUnits(12_300_000)
	nativeDelta := nativeCodec.FromUnits(1)
	wideCodec := testFixedDecimalCodec[uint256DecimalPlaces18]()
	wideBase := wideCodec.FromUnits(uint256.Int{1, 2, 3, 4})
	wideDelta := wideCodec.FromUnits(uint256.Int{1})

	b.Run("add/native_public", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value := nativeBase
			benchBoolSink = value.AddAssign(nativeDelta)
			benchPriceSink = value
		}
	})
	b.Run("add/native_integer_ceiling", func(b *testing.B) {
		a, c := nativeBase.units, nativeDelta.units
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _ = bits.Add64(a, c, 0)
		}
	})
	b.Run("add/wide_public", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value := wideBase
			benchBoolSink = value.AddAssign(wideDelta)
			benchWideSink = value
		}
	})
	b.Run("add/wide_integer_ceiling", func(b *testing.B) {
		a, c := wideBase.units, wideDelta.units
		b.ReportAllocs()
		for b.Loop() {
			var result uint256.Int
			var carry uint64
			result[0], carry = bits.Add64(a[0], c[0], 0)
			result[1], carry = bits.Add64(a[1], c[1], carry)
			result[2], carry = bits.Add64(a[2], c[2], carry)
			result[3], _ = bits.Add64(a[3], c[3], carry)
			benchU256Sink = result
		}
	})
}

func benchmarkCBORCeilings(b *testing.B) {
	codec := testUint256FixedDecimalCodec(6)
	units := uint256.Int{123_456_789}
	buffer := make([]byte, 0, MaxCBORSize)
	raw := codec.AppendCBOR(buffer[:0], units)

	b.Run("cbor/append_runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendCBOR(buffer[:0], units)
		}
	})
	b.Run("cbor/len_runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = codec.CBORLen(units)
		}
	})
	b.Run("cbor/len_complete_uint256_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = (Uint256Units{}).unitCBORLen(units)
		}
	})
	b.Run("cbor/append_integer_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = appendCBORUint64(buffer[:0], units[0])
		}
	})
	b.Run("cbor/append_complete_uint256_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = (Uint256Units{}).unitAppendCBOR(buffer[:0], units)
		}
	})
	b.Run("cbor/decode_runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _ = codec.ParseCBOR(raw)
		}
	})
	b.Run("cbor/decode_integer_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _ = parseCBORUint64(raw, ^uint64(0))
		}
	})
	b.Run("cbor/decode_complete_uint256_kernel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _ = (Uint256Units{}).unitParseCBOR(raw)
		}
	})
	b.Run("cbor/decode_first_runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, benchBytesSink, _ = codec.ParseCBORFirst(raw)
		}
	})
	b.Run("cbor/decode_first_complete_uint256_path", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, consumed, err := (Uint256Units{}).unitParseCBORFirst(raw)
			if err != "" {
				benchBytesSink = nil
				continue
			}
			benchU256Sink = value
			benchBytesSink = raw[consumed:]
		}
	})
	b.Run("cbor/decode_into_runtime_codec", func(b *testing.B) {
		var value uint256.Int
		b.ReportAllocs()
		for b.Loop() {
			_ = codec.ParseCBORInto(raw, &value)
		}
		benchU256Sink = value
	})
	b.Run("cbor/decode_into_complete_uint256_path", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			decoded, err := (Uint256Units{}).unitParseCBOR(raw)
			if err == "" {
				benchU256Sink = decoded
			}
		}
	})
	b.Run("cbor/decode_first_into_runtime_codec", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink, _ = codec.ParseCBORFirstInto(raw, &benchU256Sink)
		}
	})
	b.Run("cbor/decode_first_into_complete_uint256_path", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			decoded, consumed, err := (Uint256Units{}).unitParseCBORFirst(raw)
			if err != "" {
				benchBytesSink = nil
				continue
			}
			benchU256Sink = decoded
			benchBytesSink = raw[consumed:]
		}
	})
}
