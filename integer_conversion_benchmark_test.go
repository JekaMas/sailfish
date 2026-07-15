package sailfish

import (
	"math"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
)

var benchmarkBigIntSink big.Int

// BenchmarkFixedDecimalIntegerConversionCeilings separates third-party
// conversion work and the public result-copy cost from Sailfish dispatch. It
// prevents optimizing a wrapper toward a target below its measured kernel.
func BenchmarkFixedDecimalIntegerConversionCeilings(b *testing.B) {
	units256 := uint256.Int{1, 2, 3, 4}
	big256 := units256.ToBig()

	b.Run("copy_u256", func(b *testing.B) {
		for b.Loop() {
			benchU256Sink = units256
		}
	})
	b.Run("construct_fixed_decimal_u256", func(b *testing.B) {
		for b.Loop() {
			integerConversionWideSink = FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]{
				units: units256,
			}
		}
	})
	b.Run("holiman_from_big", func(b *testing.B) {
		for b.Loop() {
			var value uint256.Int
			if value.SetFromBig(big256) {
				b.Fatal("unexpected overflow")
			}
			benchU256Sink = value
		}
	})
	b.Run("holiman_into_big_reused", func(b *testing.B) {
		destination := new(big.Int)
		units256.IntoBig(&destination)
		b.ResetTimer()
		for b.Loop() {
			units256.IntoBig(&destination)
		}
		benchmarkBigIntSink.Set(destination)
	})
}

func BenchmarkFixedDecimalIntegerConversions(b *testing.B) {
	codec64 := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	codec256 := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	big64 := new(big.Int).SetUint64(math.MaxUint64)
	units256 := uint256.Int{1, 2, 3, 4}
	big256 := units256.ToBig()
	value64 := codec64.FromUnits(math.MaxUint64)
	value256 := codec256.FromUnits(units256)

	b.Run("from_big_int/uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			allocationPriceSink, allocationErrorSink = codec64.FromBigInt(big64)
		}
	})
	b.Run("from_big_int/uint256", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			integerConversionWideSink, allocationErrorSink = codec256.FromBigInt(big256)
		}
	})
	b.Run("from_u256/uint64", func(b *testing.B) {
		b.ReportAllocs()
		units := uint256.Int{math.MaxUint64}
		for b.Loop() {
			allocationPriceSink, allocationErrorSink = codec64.FromU256(units)
		}
	})
	b.Run("from_u256/uint256", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			integerConversionWideSink, allocationErrorSink = codec256.FromU256(units256)
		}
	})
	b.Run("to_u256/uint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink = value64.ToU256()
		}
	})
	b.Run("to_u256/uint256", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchU256Sink = value256.ToU256()
		}
	})
	b.Run("to_big_int_reused/uint64", func(b *testing.B) {
		b.ReportAllocs()
		var destination big.Int
		_ = value64.ToBigInt(&destination)
		b.ResetTimer()
		for b.Loop() {
			allocationErrorSink = value64.ToBigInt(&destination)
		}
		benchmarkBigIntSink.Set(&destination)
	})
	b.Run("to_big_int_reused/uint256", func(b *testing.B) {
		b.ReportAllocs()
		var destination big.Int
		_ = value256.ToBigInt(&destination)
		b.ResetTimer()
		for b.Loop() {
			allocationErrorSink = value256.ToBigInt(&destination)
		}
		benchmarkBigIntSink.Set(&destination)
	})
	b.Run("to_big_int_fresh/uint256", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var destination big.Int
			allocationErrorSink = value256.ToBigInt(&destination)
			benchmarkBigIntSink.Set(&destination)
		}
	})
}
