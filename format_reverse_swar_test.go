package sailfish

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"strconv"
	"testing"
)

func TestPackedASCII8MatchesReference(t *testing.T) {
	t.Parallel()

	values := []uint32{0, 1, 9, 10, 99, 100, 9_999, 10_000, 99_999_999}
	rng := rand.New(rand.NewPCG(5, 6))
	for range 100_000 {
		values = append(values, rng.Uint32N(100_000_000))
	}

	for _, value := range values {
		var got [8]byte
		binary.LittleEndian.PutUint64(got[:], packedASCII8(value))

		var want [8]byte
		text := strconv.AppendUint(nil, uint64(value), 10)
		for i := 0; i < len(want)-len(text); i++ {
			want[i] = '0'
		}
		copy(want[len(want)-len(text):], text)
		if got != want {
			t.Fatalf("packedASCII8(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestReverseSWARFormattingPreservesCallerTail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		units uint64
		scale int
	}{
		{units: 77_777, scale: 2},
		{units: 777_777, scale: 5},
		{units: 7_777_777, scale: 5},
		{units: 77_777_777, scale: 5},
		{units: 77_777_777_777_777, scale: 9},
		{units: 7_777_777_777_777_777, scale: 9},
		{units: ^uint64(0), scale: 18},
	}
	for _, test := range tests {
		var storage [64]byte
		for i := range storage {
			storage[i] = 0xa5
		}
		copy(storage[:], "pre")

		got := appendUint64Decimal(storage[:3], test.units, test.scale)
		wantText := formatScaledBigUint64(test.units, test.scale)
		if gotText := string(got[3:]); gotText != wantText {
			t.Fatalf("format(%d, %d) = %q, want %q", test.units, test.scale, gotText, wantText)
		}
		if !bytes.Equal(storage[len(got):], bytes.Repeat([]byte{0xa5}, len(storage)-len(got))) {
			t.Fatalf("format(%d, %d) modified capacity beyond returned slice", test.units, test.scale)
		}
	}
}

func formatScaledBigUint64(value uint64, scale int) string {
	digits := strconv.FormatUint(value, 10)
	if scale == 0 {
		return digits
	}
	if len(digits) > scale {
		point := len(digits) - scale
		return digits[:point] + "." + digits[point:]
	}
	return "0." + string(bytes.Repeat([]byte{'0'}, scale-len(digits))) + digits
}
