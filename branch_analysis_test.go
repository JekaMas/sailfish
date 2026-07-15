package sailfish

import (
	"math/rand/v2"
	"strconv"
	"testing"
)

func TestDecimalDigits64MatchesStandardLibrary(t *testing.T) {
	boundaries := []uint64{0, 1, 9, 10, 99, 100, ^uint64(0)}
	for _, power := range powersOf10Uint64 {
		boundaries = append(boundaries, power)
		if power > 0 {
			boundaries = append(boundaries, power-1)
		}
		if power < ^uint64(0) {
			boundaries = append(boundaries, power+1)
		}
	}
	for _, value := range boundaries {
		if got, want := decimalDigits64(value), len(strconv.FormatUint(value, 10)); got != want {
			t.Fatalf("decimalDigits64(%d) = %d, want %d", value, got, want)
		}
	}

	rng := rand.New(rand.NewPCG(1, 2))
	for range 100_000 {
		value := rng.Uint64()
		if got, want := decimalDigits64(value), len(strconv.FormatUint(value, 10)); got != want {
			t.Fatalf("decimalDigits64(%d) = %d, want %d", value, got, want)
		}
	}
}
