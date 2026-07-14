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
)

func BenchmarkCodecParsePriceScales(b *testing.B) {
	benchmarkParsePrice[PriceScale1](b, "scale_1", "123456789012345678.9")
	benchmarkParsePrice[PriceScale2](b, "scale_2", "12345678901234567.89")
	benchmarkParsePrice[PriceScale3](b, "scale_3", "1234567890123456.789")
	benchmarkParsePrice[PriceScale4](b, "scale_4", "123456789012345.6789")
	benchmarkParsePrice[PriceScale5](b, "scale_5", "12345678901234.56789")
	benchmarkParsePrice[PriceScale6](b, "scale_6", "1234567890123.456789")
	benchmarkParsePrice[PriceScale7](b, "scale_7", "123456789012.3456789")
	benchmarkParsePrice[PriceScale8](b, "scale_8", "12345678901.23456789")
	benchmarkParsePrice[PriceScale9](b, "scale_9", "1234567890.123456789")
}

func benchmarkParsePrice[V Venue[uint64]](b *testing.B, name, input string) {
	b.Helper()
	codec := MustCodec[V]()
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			value, _ := codec.Parse(input)
			benchUint64Sink = value.Units()
		}
	})
}

func BenchmarkCodecParse(b *testing.B) {
	b.Run("uint64/canonical", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchPriceSink, _ = codec.Parse("123.31232")
		}
	})
	b.Run("uint64/compact", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchPriceSink, _ = codec.ParseCompact("123.31232")
		}
	})
	b.Run("uint64/bytes", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		input := []byte("123.31232")
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchPriceSink, _ = codec.ParseBytes(input)
		}
	})
	b.Run("uint64/noncanonical", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchPriceSink, _ = codec.ParseCompact("00123.31")
		}
	})
	b.Run("uint64/invalid", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchPriceSink, _ = codec.ParseCompact("123x31232")
		}
	})
	b.Run("uint256/canonical", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchWideSink, _ = codec.Parse("12345678901234567890.123456789012345678")
		}
	})
	b.Run("uint256/max", func(b *testing.B) {
		codec := MustCodec[uint256Scale0]()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = codec.Parse(maxUint256Decimal)
		}
	})
}

func BenchmarkReferenceStrconvSplitUint64(b *testing.B) {
	const input = "123.31232"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		point := strings.IndexByte(input, '.')
		whole, _ := strconv.ParseUint(input[:point], 10, 64)
		fraction, _ := strconv.ParseUint(input[point+1:], 10, 64)
		benchUint64Sink = whole*100_000 + fraction
	}
}

func BenchmarkAppendTo(b *testing.B) {
	b.Run("uint64/retained", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		value, _ := codec.Parse("123.31232")
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("uint64/formatted", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		value := codec.FromUnits(12_331_232)
		buffer := make([]byte, 0, 32)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
	b.Run("uint256/formatted", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		value := codec.FromUnits(uint256.Int{1, 2, 3, 4})
		buffer := make([]byte, 0, 96)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchBytesSink = codec.AppendTo(buffer[:0], value)
		}
	})
}

func BenchmarkString(b *testing.B) {
	b.Run("uint64/retained", func(b *testing.B) {
		value, _ := New[PriceScale5]("123.31232")
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchStringSink = value.String()
		}
	})
	b.Run("uint64/formatted", func(b *testing.B) {
		value := MustCodec[PriceScale5]().FromUnits(12_331_232)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchStringSink = value.String()
		}
	})
}

func BenchmarkCompare(b *testing.B) {
	b.Run("uint64/same-scale", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		a := codec.FromUnits(12_331_232)
		c := codec.FromUnits(12_331_233)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchIntSink = a.Compare(c)
		}
	})
	b.Run("uint256/same-scale", func(b *testing.B) {
		codec := MustCodec[uint256Scale18]()
		a := codec.FromUnits(uint256.Int{1, 2, 3, 4})
		c := codec.FromUnits(uint256.Int{2, 2, 3, 4})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchIntSink = a.Compare(c)
		}
	})
	b.Run("cross-scale", func(b *testing.B) {
		a := MustCodec[PriceScale5]().FromUnits(12_331_232)
		c := MustCodec[uint256Scale18]().FromUnits(uint256.Int{12_331_232_000_000_000_000})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchIntSink = Compare(a, c)
		}
	})
}

func BenchmarkAddAssign(b *testing.B) {
	b.Run("uint64", func(b *testing.B) {
		codec := MustCodec[PriceScale5]()
		base := codec.FromUnits(12_300_000)
		delta := codec.FromUnits(1)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
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
		for i := 0; i < b.N; i++ {
			value := base
			benchBoolSink = value.AddAssign(delta)
			benchWideSink = value
		}
	})
}
