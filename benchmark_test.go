package sailfish

import (
	"strconv"
	"strings"
	"testing"

	"github.com/holiman/uint256"
)

var (
	benchPriceSink  price5
	benchWideSink   wide18
	benchBytesSink  []byte
	benchStringSink string
	benchIntSink    int
	benchBoolSink   bool
	benchUint64Sink uint64
	benchU256Sink   uint256.Int
)

func BenchmarkCodecParsePriceFractions(b *testing.B) {
	benchmarkParsePrice[PriceUint64[Fraction1]](b, "scale_1", "123456789012345678.9")
	benchmarkParsePrice[PriceUint64[Fraction2]](b, "scale_2", "12345678901234567.89")
	benchmarkParsePrice[PriceUint64[Fraction3]](b, "scale_3", "1234567890123456.789")
	benchmarkParsePrice[PriceUint64[Fraction4]](b, "scale_4", "123456789012345.6789")
	benchmarkParsePrice[PriceUint64[Fraction5]](b, "scale_5", "12345678901234.56789")
	benchmarkParsePrice[PriceUint64[Fraction6]](b, "scale_6", "1234567890123.456789")
	benchmarkParsePrice[PriceUint64[Fraction7]](b, "scale_7", "123456789012.3456789")
	benchmarkParsePrice[PriceUint64[Fraction8]](b, "scale_8", "12345678901.23456789")
	benchmarkParsePrice[PriceUint64[Fraction9]](b, "scale_9", "1234567890.123456789")
}

type benchmarkConcreteFraction5 struct{ Uint64Units }

func (benchmarkConcreteFraction5) NotionScale() Notion { return 5 }

// BenchmarkScaleMetadataDispatch measures the explicit generic format against
// a test-local concrete venue. Codec caches scale metadata, while one-shot
// constructors and direct Decimal methods resolve it on every call.
func BenchmarkScaleMetadataDispatch(b *testing.B) {
	b.Run("new/explicit_format", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := New[PriceUint64[Fraction5]]("123.31232")
			_ = value
		}
	})
	b.Run("new/test_concrete", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := New[benchmarkConcreteFraction5]("123.31232")
			_ = value
		}
	})

	explicitCodec := MustCodec[PriceUint64[Fraction5]]()
	concreteCodec := MustCodec[benchmarkConcreteFraction5]()
	b.Run("codec_parse/explicit_format", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := explicitCodec.Parse("123.31232")
			_ = value
		}
	})
	b.Run("codec_parse/test_concrete", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := concreteCodec.Parse("123.31232")
			_ = value
		}
	})

	explicitValue := explicitCodec.FromUnits(12_331_232)
	concreteValue := concreteCodec.FromUnits(12_331_232)
	buffer := make([]byte, 0, 32)
	b.Run("decimal_append/explicit_format", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := explicitValue.AppendTo(buffer[:0])
			_ = out
		}
	})
	b.Run("decimal_append/test_concrete", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := concreteValue.AppendTo(buffer[:0])
			_ = out
		}
	})
}

// BenchmarkGenericBackendDispatch preserves the rejected design comparison.
// Selecting parse operations through a generic backend type switch is
// measurably slower than embedding a concrete zero-sized provider in a format.
func BenchmarkGenericBackendDispatch(b *testing.B) {
	b.Run("uint64/concrete_provider", func(b *testing.B) {
		provider := Uint64Units{}
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _, _ = provider.unitParseString("123.31232", 5)
		}
	})
	b.Run("uint64/generic_function", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _, _ = benchmarkGenericBackendParse[uint64]("123.31232", 5)
		}
	})
	b.Run("uint64/generic_method", func(b *testing.B) {
		provider := benchmarkGenericBackend[uint64]{}
		b.ReportAllocs()
		for b.Loop() {
			benchUint64Sink, _, _ = provider.parse("123.31232", 5)
		}
	})

	b.Run("uint256/concrete_provider", func(b *testing.B) {
		provider := Uint256Units{}
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _, _ = provider.unitParseString("123.456789", 6)
		}
	})
	b.Run("uint256/generic_function", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _, _ = benchmarkGenericBackendParse[uint256.Int]("123.456789", 6)
		}
	})
	b.Run("uint256/generic_method", func(b *testing.B) {
		provider := benchmarkGenericBackend[uint256.Int]{}
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _, _ = provider.parse("123.456789", 6)
		}
	})
}

type benchmarkGenericBackend[U Unit] struct{}

func (benchmarkGenericBackend[U]) parse(input string, scale int) (U, bool, Error) {
	return benchmarkGenericBackendParse[U](input, scale)
}

func benchmarkGenericBackendParse[U Unit](input string, scale int) (U, bool, Error) {
	var zero U
	switch any(zero).(type) {
	case uint64:
		value, canonical, err := parseUint64(input, scale)
		return any(value).(U), canonical, err
	case uint256.Int:
		value, canonical, err := parseUint256(input, scale)
		return any(value).(U), canonical, err
	default:
		panic("unsupported benchmark backend")
	}
}

func BenchmarkExplicitUnitWidths(b *testing.B) {
	benchmarkUnitWidth(b, "uint8", MustCodec[PriceUint8[Fraction1]](), uint8(255), "25.5")
	benchmarkUnitWidth(b, "uint16", MustCodec[PriceUint16[Fraction1]](), uint16(255), "25.5")
	benchmarkUnitWidth(b, "uint32", MustCodec[PriceUint32[Fraction1]](), uint32(255), "25.5")
	benchmarkUnitWidth(b, "uint64", MustCodec[PriceUint64[Fraction1]](), uint64(255), "25.5")
}

func benchmarkUnitWidth[V Venue[U], U Unit](
	b *testing.B,
	name string,
	codec Codec[V, U],
	units U,
	input string,
) {
	b.Helper()
	b.Run("parse/"+name, func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.ParseCompact(input)
			_ = value
		}
	})
	value := codec.FromUnits(units)
	buffer := make([]byte, 0, 32)
	b.Run("append/"+name, func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			out := codec.AppendTo(buffer[:0], value)
			_ = out
		}
	})
}

func benchmarkParsePrice[V Venue[uint64]](b *testing.B, name, input string) {
	b.Helper()
	codec := MustCodec[V]()
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.Parse(input)
			benchUint64Sink = value.Units()
		}
	})
}

func BenchmarkCodecParse(b *testing.B) {
	b.Run("uint64/canonical", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		b.ReportAllocs()
		for b.Loop() {
			benchPriceSink, _ = codec.Parse("123.31232")
		}
	})
	b.Run("uint64/compact", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		b.ReportAllocs()
		for b.Loop() {
			benchPriceSink, _ = codec.ParseCompact("123.31232")
		}
	})
	b.Run("uint64/bytes", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		input := []byte("123.31232")
		b.ReportAllocs()
		for b.Loop() {
			benchPriceSink, _ = codec.ParseBytes(input)
		}
	})
	b.Run("uint64/noncanonical", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		b.ReportAllocs()
		for b.Loop() {
			benchPriceSink, _ = codec.ParseCompact("00123.31")
		}
	})
	b.Run("uint64/invalid", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		b.ReportAllocs()
		for b.Loop() {
			benchPriceSink, _ = codec.ParseCompact("123x31232")
		}
	})
	b.Run("uint256/canonical", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		b.ReportAllocs()
		for b.Loop() {
			benchWideSink, _ = codec.Parse("12345678901234567890.123456789012345678")
		}
	})
	b.Run("uint256/max", func(b *testing.B) {
		codec := MustCodec[uint256Scale0]()
		b.ReportAllocs()
		for b.Loop() {
			_, _ = codec.Parse(maxUint256Decimal)
		}
	})
}

func BenchmarkReferenceStrconvSplitUint64(b *testing.B) {
	const input = "123.31232"
	b.ReportAllocs()
	for b.Loop() {
		point := strings.IndexByte(input, '.')
		whole, _ := strconv.ParseUint(input[:point], 10, 64)
		fraction, _ := strconv.ParseUint(input[point+1:], 10, 64)
		benchUint64Sink = whole*100_000 + fraction
	}
}

func BenchmarkAppendTo(b *testing.B) {
	b.Run("uint64/retained", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		value, _ := codec.Parse("123.31232")
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("uint64/formatted", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		value := codec.FromUnits(12_331_232)
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("uint256/formatted", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		value := codec.FromUnits(uint256.Int{1, 2, 3, 4})
		buffer := make([]byte, 0, 96)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
}

// BenchmarkUint256MarketHotPaths separates ordinary venue-sized values from
// full-width uint256 values. Most CEX prices and amounts occupy one limb even
// when the durable representation uses uint256.Int; the wide cases protect the
// worst-case contract without letting them obscure the dominant workload.
func BenchmarkUint256MarketHotPaths(b *testing.B) {
	b.Run("parse_string/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale6]()
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.ParseCompact("123.456789")
			benchU256Sink = value.Units()
		}
	})
	b.Run("parse_internal/cex_scale6_one_limb", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _, _ := parseUint256("123.456789", 6)
			benchU256Sink = value
		}
	})
	b.Run("parse_runtime_codec/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustUint256Codec(6)
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink, _ = codec.Parse("123.456789")
		}
	})
	b.Run("parse_into_runtime_codec/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustUint256Codec(6)
		var value uint256.Int
		b.ReportAllocs()
		for b.Loop() {
			_ = codec.ParseInto("123.456789", &value)
		}
		benchU256Sink = value
	})
	b.Run("parse_uint64_internal/cex_scale6_one_limb", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _, _ := parseUint64("123.456789", 6)
			benchUint64Sink = value
		}
	})
	b.Run("parse_uint64_with_dot_internal/cex_scale6_one_limb", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			value, _ := parseUint64WithDot("123.456789", 3)
			benchUint64Sink = value
		}
	})
	b.Run("parse_bytes/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale6]()
		input := []byte("123.456789")
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.ParseBytes(input)
			benchU256Sink = value.Units()
		}
	})
	b.Run("parse_string/scale18_one_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.ParseCompact("0.123456789012345678")
			benchU256Sink = value.Units()
		}
	})
	b.Run("parse_string/scale18_two_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.ParseCompact("12345678901234567890.123456789012345678")
			benchU256Sink = value.Units()
		}
	})
	b.Run("parse_string/scale0_four_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale0]()
		b.ReportAllocs()
		for b.Loop() {
			value, _ := codec.ParseCompact(maxUint256Decimal)
			benchU256Sink = value.Units()
		}
	})

	b.Run("append/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale6]()
		value := codec.FromUnits(uint256.Int{123_456_789})
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append_retained/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale6]()
		value, err := codec.Parse("123.456789")
		if err != nil {
			b.Fatal(err)
		}
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append_internal/cex_scale6_one_limb", func(b *testing.B) {
		value := uint256.Int{123_456_789}
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = appendUint256Decimal(buffer[:0], value, 6)
		}
	})
	b.Run("append_runtime_codec/cex_scale6_one_limb", func(b *testing.B) {
		codec := MustUint256Codec(6)
		value := uint256.Int{123_456_789}
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append_uint64_internal/cex_scale6_one_limb", func(b *testing.B) {
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = appendUint64Decimal(buffer[:0], 123_456_789, 6)
		}
	})
	b.Run("fill_uint64_internal/cex_scale6_one_limb", func(b *testing.B) {
		var buffer [9]byte
		b.ReportAllocs()
		for b.Loop() {
			fillUnsigned64(buffer[:], 123_456_789)
			benchBytesSink = buffer[:]
		}
	})
	b.Run("append/scale18_one_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		value := codec.FromUnits(uint256.Int{123_456_789_012_345_678})
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append/scale18_two_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		value := codec.FromUnits(uint256.Int{1, 1})
		buffer := make([]byte, 0, 64)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append/scale18_four_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		value := codec.FromUnits(uint256.Int{1, 2, 3, 4})
		buffer := make([]byte, 0, 96)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append_runtime_codec/scale18_four_limb", func(b *testing.B) {
		codec := MustUint256Codec(18)
		value := uint256.Int{1, 2, 3, 4}
		buffer := make([]byte, 0, 96)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("append_retained/scale18_four_limb", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		units := uint256.Int{1, 2, 3, 4}
		text := string(appendUint256Decimal(nil, units, 18))
		value, err := codec.Parse(text)
		if err != nil {
			b.Fatal(err)
		}
		buffer := make([]byte, 0, 96)
		b.ReportAllocs()
		for b.Loop() {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
}

func BenchmarkString(b *testing.B) {
	b.Run("uint64/retained", func(b *testing.B) {
		value, _ := New[PriceUint64[Fraction5]]("123.31232")
		b.ReportAllocs()
		for b.Loop() {
			benchStringSink = value.String()
		}
	})
	b.Run("uint64/formatted", func(b *testing.B) {
		value := MustCodec[PriceUint64[Fraction5]]().FromUnits(12_331_232)
		b.ReportAllocs()
		for b.Loop() {
			benchStringSink = value.String()
		}
	})
}

func BenchmarkCompare(b *testing.B) {
	b.Run("uint64/same-scale", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		a := codec.FromUnits(12_331_232)
		c := codec.FromUnits(12_331_233)
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = a.Compare(c)
		}
	})
	b.Run("uint256/same-scale", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		a := codec.FromUnits(uint256.Int{1, 2, 3, 4})
		c := codec.FromUnits(uint256.Int{2, 2, 3, 4})
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = a.Compare(c)
		}
	})
	b.Run("cross-scale", func(b *testing.B) {
		a := MustCodec[PriceUint64[Fraction5]]().FromUnits(12_331_232)
		c := MustCodec[uint256Scale18]().FromUnits(uint256.Int{12_331_232_000_000_000_000})
		b.ReportAllocs()
		for b.Loop() {
			benchIntSink = Compare(a, c)
		}
	})
}

func BenchmarkAddAssign(b *testing.B) {
	b.Run("uint64", func(b *testing.B) {
		codec := MustCodec[PriceUint64[Fraction5]]()
		base := codec.FromUnits(12_300_000)
		delta := codec.FromUnits(1)
		b.ReportAllocs()
		for b.Loop() {
			value := base
			benchBoolSink = value.AddAssign(delta)
			benchPriceSink = value
		}
	})
	b.Run("uint256", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		base := codec.FromUnits(uint256.Int{1, 2, 3, 4})
		delta := codec.FromUnits(uint256.Int{1})
		b.ReportAllocs()
		for b.Loop() {
			value := base
			benchBoolSink = value.AddAssign(delta)
			benchWideSink = value
		}
	})
}
