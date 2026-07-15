package sailfish

import (
	"testing"
)

var branchAnalysisDigitsSink int

// BenchmarkDecimalDigitsDistributions protects fixed-width predictable,
// mixed-width, and market-shaped inputs. The accepted production algorithm is
// branchless with respect to decimal width; PERFORMANCE.md records the removed
// comparison-tree baseline.
func BenchmarkDecimalDigitsDistributions(b *testing.B) {
	distributions := []struct {
		name   string
		values []uint64
	}{
		{
			name: "predictable_8_digits",
			values: []uint64{
				12_331_232, 98_765_432, 10_000_001, 55_555_555,
				42_424_242, 87_654_321, 11_111_111, 76_543_210,
			},
		},
		{
			name: "mixed_widths",
			values: []uint64{
				0, 9, 10, 99, 100, 999, 1_000, 9_999,
				10_000, 99_999, 1_000_000, 99_999_999,
				1_000_000_000, 99_999_999_999, 1_000_000_000_000,
				99_999_999_999_999, 1_000_000_000_000_000,
				99_999_999_999_999_999, 1_000_000_000_000_000_000,
				^uint64(0),
			},
		},
		{
			name: "market_shaped",
			values: []uint64{
				1, 12, 1_234, 12_331_232, 99_999_999, 123_456_789,
				1_000_000_000, 12_345_678_901, 999_999_999_999,
				12_345_678_901_234, 999_999_999_999_999,
				12_345_678_901_234_567, 999_999_999_999_999_999,
			},
		},
	}

	for _, distribution := range distributions {
		b.Run(distribution.name, func(b *testing.B) {
			index := 0
			b.ReportAllocs()
			for b.Loop() {
				branchAnalysisDigitsSink = decimalDigits64(distribution.values[index])
				index++
				if index == len(distribution.values) {
					index = 0
				}
			}
		})
	}
}
