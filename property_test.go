package sailfish

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/holiman/uint256"
)

func TestUint64FormatParseProperties(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(1))
	codec := MustCodec[PriceScale9]()
	for range 20_000 {
		units := rng.Uint64()
		value := codec.FromUnits(units)
		text := value.String()
		if text != formatScaledBig(new(big.Int).SetUint64(units), 9) {
			t.Fatalf("format %d = %q", units, text)
		}
		round, err := codec.ParseCompact(text)
		if err != nil || round.Units() != units {
			t.Fatalf("round trip %d through %q = %d, %v", units, text, round.Units(), err)
		}
	}
}

func TestUint256FormatParseProperties(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(2))
	codec := MustCodec[uint256Scale18]()
	for range 5_000 {
		units := uint256.Int{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		value := codec.FromUnits(units)
		text := value.String()
		if text != formatScaledBig(uint256ToBig(units), 18) {
			t.Fatalf("format %#v = %q", units, text)
		}
		round, err := codec.ParseCompact(text)
		if err != nil || round.Units() != units {
			t.Fatalf("round trip %#v through %q = %#v, %v", units, text, round.Units(), err)
		}
	}
}

func formatScaledBig(value *big.Int, scale int) string {
	digits := value.String()
	if scale == 0 {
		return digits
	}
	if len(digits) > scale {
		point := len(digits) - scale
		return digits[:point] + "." + digits[point:]
	}
	zeros := make([]byte, scale-len(digits))
	for i := range zeros {
		zeros[i] = '0'
	}
	return "0." + string(zeros) + digits
}
