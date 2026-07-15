package sailfish

import (
	"testing"

	"github.com/holiman/uint256"
)

// BenchmarkOptimizationParseLengths exposes the dense digit runs targeted by
// scalar and SWAR parser candidates without codec or ownership overhead.
func BenchmarkOptimizationParseLengths(b *testing.B) {
	native := []struct {
		name  string
		input string
	}{
		{name: "digits_1", input: "7"},
		{name: "digits_2", input: "42"},
		{name: "digits_7", input: "1234567"},
		{name: "digits_8", input: "12345678"},
		{name: "digits_9", input: "123456789"},
		{name: "digits_16", input: "1234567890123456"},
		{name: "digits_19", input: "1234567890123456789"},
		{name: "digits_20", input: "18446744073709551615"},
	}
	for _, tt := range native {
		b.Run("uint64/string/"+tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(tt.input)))
			b.ReportAllocs()
			for b.Loop() {
				benchUint64Sink, _ = parseUint64Digits(tt.input, 0, len(tt.input))
			}
		})
		inputBytes := []byte(tt.input)
		b.Run("uint64/bytes/"+tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(inputBytes)))
			b.ReportAllocs()
			for b.Loop() {
				benchUint64Sink, _ = parseUint64Digits(inputBytes, 0, len(inputBytes))
			}
		})
	}

	wide := []struct {
		name  string
		input string
	}{
		{name: "digits_19", input: "1234567890123456789"},
		{name: "digits_20", input: "12345678901234567890"},
		{name: "digits_38", input: "12345678901234567890123456789012345678"},
		{name: "digits_39", input: "123456789012345678901234567890123456789"},
		{name: "digits_57", input: "123456789012345678901234567890123456789012345678901234567"},
		{name: "digits_58", input: "1234567890123456789012345678901234567890123456789012345678"},
		{name: "digits_77", input: "12345678901234567890123456789012345678901234567890123456789012345678901234567"},
		{name: "max_78", input: maxUint256Decimal},
	}
	for _, tt := range wide {
		b.Run("uint256/string/"+tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(tt.input)))
			b.ReportAllocs()
			for b.Loop() {
				benchU256Sink, _ = parseUint256Digits(tt.input, 0, len(tt.input))
			}
		})
		inputBytes := []byte(tt.input)
		b.Run("uint256/bytes/"+tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(inputBytes)))
			b.ReportAllocs()
			for b.Loop() {
				benchU256Sink, _ = parseUint256Digits(inputBytes, 0, len(inputBytes))
			}
		})
	}
}

func BenchmarkOptimizationCanonicalParse(b *testing.B) {
	cases := []struct {
		name  string
		input string
		scale int
	}{
		{name: "digits_8_scale_5", input: "123.45678", scale: 5},
		{name: "digits_16_scale_5", input: "12345678901.23456", scale: 5},
		{name: "digits_16_scale_9", input: "1234567.890123456", scale: 9},
		{name: "scale_2", input: "12345678901234567.89", scale: 2},
		{name: "scale_5", input: "12345678901234.56789", scale: 5},
		{name: "scale_9", input: "1234567890.123456789", scale: 9},
		{name: "scale_18", input: "1.234567890123456789", scale: 18},
	}
	for _, tt := range cases {
		b.Run("string/"+tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(tt.input)))
			b.ReportAllocs()
			for b.Loop() {
				benchUint64Sink, _, _ = parseUint64(tt.input, tt.scale)
			}
		})
		inputBytes := []byte(tt.input)
		b.Run("bytes/"+tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(inputBytes)))
			b.ReportAllocs()
			for b.Loop() {
				benchUint64Sink, _, _ = parseUint64(inputBytes, tt.scale)
			}
		})
	}
}

func BenchmarkOptimizationParseErrors(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "invalid_first", input: "x234567890123456789"},
		{name: "invalid_middle", input: "123456789x123456789"},
		{name: "invalid_last", input: "123456789012345678x"},
		{name: "overflow", input: "18446744073709551616"},
	}
	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(tt.input)))
			b.ReportAllocs()
			for b.Loop() {
				benchUint64Sink, _ = parseUint64Digits(tt.input, 0, len(tt.input))
			}
		})
	}
}

func BenchmarkOptimizationBatchParse(b *testing.B) {
	inputs := [...]string{
		"0.00001",
		"1.25000",
		"123.31232",
		"999999.99999",
		"42.00000",
		"0.12500",
		"25000.50000",
		"18446744073709.55161",
	}
	for _, size := range []int{1, 8, 64, 256} {
		name := map[int]string{1: "items_1", 8: "items_8", 64: "items_64", 256: "items_256"}[size]
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size * len(inputs[2])))
			for b.Loop() {
				var sum uint64
				for i := range size {
					value, _, _ := parseUint64(inputs[i&7], 5)
					sum += value
				}
				benchUint64Sink = sum
			}
		})
	}
}

func BenchmarkOptimizationFormatNative(b *testing.B) {
	cases := []struct {
		name  string
		units uint64
		scale int
	}{
		{name: "digits_1_scale_0", units: 7, scale: 0},
		{name: "digits_2_scale_0", units: 42, scale: 0},
		{name: "digits_8_scale_0", units: 12_345_678, scale: 0},
		{name: "digits_9_scale_5", units: 123_456_789, scale: 5},
		{name: "digits_16_scale_9", units: 1_234_567_890_123_456, scale: 9},
		{name: "digits_19_scale_18", units: 1_234_567_890_123_456_789, scale: 18},
		{name: "digits_20_scale_0", units: ^uint64(0), scale: 0},
		{name: "below_scale", units: 123, scale: 9},
	}
	buffer := make([]byte, 0, maxUint64TextLen)
	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink = appendUint64Decimal(buffer[:0], tt.units, tt.scale)
			}
		})
	}
}

// BenchmarkOptimizationFormatReverseSWAR covers both selected packed widths
// and the adjacent pair-table widths protected from dispatch regressions.
func BenchmarkOptimizationFormatReverseSWAR(b *testing.B) {
	cases := []struct {
		name  string
		units uint64
		scale int
	}{
		{name: "digits_2_scale_1", units: 77, scale: 1},
		{name: "digits_3_scale_2", units: 777, scale: 2},
		{name: "digits_4_scale_2", units: 7_777, scale: 2},
		{name: "digits_5_scale_2", units: 77_777, scale: 2},
		{name: "digits_6_scale_5", units: 777_777, scale: 5},
		{name: "digits_7_scale_5", units: 7_777_777, scale: 5},
		{name: "digits_8_scale_1", units: 77_777_777, scale: 1},
		{name: "digits_8_scale_5", units: 77_777_777, scale: 5},
		{name: "digits_8_scale_7", units: 77_777_777, scale: 7},
		{name: "digits_9_scale_5_protected", units: 777_777_777, scale: 5},
		{name: "digits_11_scale_5_protected", units: 77_777_777_777, scale: 5},
		{name: "digits_12_scale_9", units: 777_777_777_777, scale: 9},
		{name: "digits_13_scale_9", units: 7_777_777_777_777, scale: 9},
		{name: "digits_14_scale_9", units: 77_777_777_777_777, scale: 9},
		{name: "digits_15_scale_9", units: 777_777_777_777_777, scale: 9},
		{name: "digits_16_scale_9", units: 7_777_777_777_777_777, scale: 9},
		{name: "digits_17_scale_9", units: 77_777_777_777_777_777, scale: 9},
		{name: "digits_18_scale_9", units: 777_777_777_777_777_777, scale: 9},
		{name: "digits_19_scale_18", units: 7_777_777_777_777_777_777, scale: 18},
		{name: "digits_20_scale_18", units: ^uint64(0), scale: 18},
	}
	var buffer [maxUint64TextLen]byte
	for _, test := range cases {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink = appendUint64Decimal(buffer[:0], test.units, test.scale)
			}
		})
	}
}

func BenchmarkOptimizationFormatWide(b *testing.B) {
	cases := []struct {
		name  string
		units uint256.Int
		scale int
	}{
		{name: "one_limb_scale_18", units: uint256.Int{123_456_789_012_345_678}, scale: 18},
		{name: "two_limbs_scale_5", units: uint256.Int{1, 1}, scale: 5},
		{name: "three_limbs_scale_18", units: uint256.Int{1, 2, 3}, scale: 18},
		{name: "four_limbs_scale_0", units: uint256.Int{1, 2, 3, 4}, scale: 0},
		{name: "four_limbs_scale_18", units: uint256.Int{1, 2, 3, 4}, scale: 18},
		{name: "maximum_scale_18", units: uint256.Int{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}, scale: 18},
	}
	buffer := make([]byte, 0, maxUint256TextLen)
	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchBytesSink = appendUint256Decimal(buffer[:0], tt.units, tt.scale)
			}
		})
	}
}
