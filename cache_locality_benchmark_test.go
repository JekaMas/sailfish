package sailfish

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/holiman/uint256"
)

type cacheAmount18 = FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]

// BenchmarkCacheLocalityUnitScans measures the representation choice that can
// affect cache residency in numeric batches. FixedDecimal carries optional retained
// text by contract; callers that need only scaled units can use FixedDecimalCodec/Units or
// Uint256FixedDecimalCodec and keep compact unit arrays instead.
func BenchmarkCacheLocalityUnitScans(b *testing.B) {
	for _, count := range [...]int{
		2_048,   // 48 KiB native / 96 KiB wide: below the performance-core L1.
		174_762, // About 4 MiB native / 8 MiB wide: below the performance-core L2.
		699_050, // About 16 MiB native / 32 MiB wide: above the performance-core L2.
	} {
		name := fmt.Sprintf("count_%d", count)
		benchmarkCacheLocalityUint64Scans(b, name, count)
		benchmarkCacheLocalityUint256Scans(b, name, count)
	}
}

// BenchmarkCacheLocalityRandomUnitScans removes the sequential prefetch
// advantage while keeping index generation outside the timed region. It
// measures whether FixedDecimal's optional text state makes a numeric working set
// materially more expensive than the raw-unit representation already exposed
// by the package.
func BenchmarkCacheLocalityRandomUnitScans(b *testing.B) {
	for _, count := range [...]int{2_048, 174_762, 699_050} {
		name := fmt.Sprintf("count_%d", count)
		order := cachePseudoRandomOrder(count)
		benchmarkCacheLocalityRandomUint64Scans(b, name, count, order)
		benchmarkCacheLocalityRandomUint256Scans(b, name, count, order)
	}
}

// BenchmarkRawUnitBoundaryCost checks whether using compact unit arrays needs
// another public parser/formatter API. The direct cases are the package's
// internal kernels; equivalent public cases must stay close enough that a
// second representation API would add no measured value.
func BenchmarkRawUnitBoundaryCost(b *testing.B) {
	const input = "123.31232"
	const units = uint64(12_331_232)
	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	var venue PriceInUint64Units[DecimalPlaces5]

	b.Run("parse/public_compact_then_units", func(b *testing.B) {
		b.ReportAllocs()
		var result uint64
		for b.Loop() {
			value, err := codec.ParseCompact(input)
			if err != nil {
				b.Fatal(err)
			}
			result = value.Units()
		}
		benchUint64Sink = result
	})

	b.Run("parse/public_units", func(b *testing.B) {
		b.ReportAllocs()
		var result uint64
		for b.Loop() {
			var err Error
			result, err = codec.ParseUnits(input)
			if err != "" {
				b.Fatal(err)
			}
		}
		benchUint64Sink = result
	})

	b.Run("parse/internal_units_kernel", func(b *testing.B) {
		b.ReportAllocs()
		var result uint64
		for b.Loop() {
			var err Error
			result, _, err = venue.unitParseString(input, 5)
			if err != "" {
				b.Fatal(err)
			}
		}
		benchUint64Sink = result
	})

	b.Run("append/public_from_units", func(b *testing.B) {
		b.ReportAllocs()
		var buffer [32]byte
		var result []byte
		for b.Loop() {
			result = codec.AppendTo(buffer[:0], codec.FromUnits(units))
		}
		benchBytesSink = result
	})

	b.Run("append/public_units", func(b *testing.B) {
		b.ReportAllocs()
		var buffer [32]byte
		var result []byte
		for b.Loop() {
			result = codec.AppendUnits(buffer[:0], units)
		}
		benchBytesSink = result
	})

	b.Run("append/internal_units_kernel", func(b *testing.B) {
		b.ReportAllocs()
		var buffer [32]byte
		var result []byte
		for b.Loop() {
			result = venue.unitAppend(buffer[:0], units, 5)
		}
		benchBytesSink = result
	})
}

func benchmarkCacheLocalityUint64Scans(b *testing.B, name string, count int) {
	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	units := make([]uint64, count)
	decimals := make([]price5, count)
	for i := range count {
		units[i] = uint64(i + 1)
		decimals[i] = codec.FromUnits(units[i])
	}

	b.Run("uint64/raw/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(count) * int64(unsafe.Sizeof(units[0])))
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for i := range units {
				sum += units[i]
			}
		}
		benchUint64Sink = sum
	})

	b.Run("uint64/decimal/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(count) * int64(unsafe.Sizeof(decimals[0])))
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for i := range decimals {
				sum += decimals[i].Units()
			}
		}
		benchUint64Sink = sum
	})
}

func benchmarkCacheLocalityUint256Scans(b *testing.B, name string, count int) {
	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	units := make([]uint256.Int, count)
	decimals := make([]cacheAmount18, count)
	for i := range count {
		units[i] = uint256.Int{uint64(i + 1), uint64(i & 7), 0, 0}
		decimals[i] = codec.FromUnits(units[i])
	}

	b.Run("uint256/raw/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(count) * int64(unsafe.Sizeof(units[0])))
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for i := range units {
				sum += units[i][0]
			}
		}
		benchUint64Sink = sum
	})

	b.Run("uint256/decimal/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(count) * int64(unsafe.Sizeof(decimals[0])))
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for i := range decimals {
				sum += decimals[i].Units()[0]
			}
		}
		benchUint64Sink = sum
	})
}

func benchmarkCacheLocalityRandomUint64Scans(b *testing.B, name string, count int, order []int) {
	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	units := make([]uint64, count)
	decimals := make([]price5, count)
	for i := range count {
		units[i] = uint64(i + 1)
		decimals[i] = codec.FromUnits(units[i])
	}

	b.Run("uint64/raw/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for _, index := range order {
				sum += units[index]
			}
		}
		benchUint64Sink = sum
	})

	b.Run("uint64/decimal/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for _, index := range order {
				sum += decimals[index].Units()
			}
		}
		benchUint64Sink = sum
	})
}

func benchmarkCacheLocalityRandomUint256Scans(b *testing.B, name string, count int, order []int) {
	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	units := make([]uint256.Int, count)
	decimals := make([]cacheAmount18, count)
	for i := range count {
		units[i] = uint256.Int{uint64(i + 1), uint64(i & 7), 0, 0}
		decimals[i] = codec.FromUnits(units[i])
	}

	b.Run("uint256/raw/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for _, index := range order {
				sum += units[index][0]
			}
		}
		benchUint64Sink = sum
	})

	b.Run("uint256/decimal/"+name, func(b *testing.B) {
		b.ReportAllocs()
		b.ReportMetric(float64(count), "values/op")
		var sum uint64
		for b.Loop() {
			sum = 0
			for _, index := range order {
				sum += decimals[index].Units()[0]
			}
		}
		benchUint64Sink = sum
	})
}

func cachePseudoRandomOrder(count int) []int {
	order := make([]int, count)
	for i := range order {
		order[i] = i
	}

	// Deterministic xorshift64* keeps benchmark fixtures reproducible. The
	// shuffle runs only during setup, outside benchmark timing.
	state := uint64(0x9e3779b97f4a7c15)
	for i := count - 1; i > 0; i-- {
		state ^= state >> 12
		state ^= state << 25
		state ^= state >> 27
		j := int((state * 0x2545f4914f6cdd1d) % uint64(i+1))
		order[i], order[j] = order[j], order[i]
	}
	return order
}
