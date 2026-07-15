package sailfish

import (
	"errors"
	"strconv"
	"testing"
	"unsafe"
)

func TestFixedDecimalCodecAPIAndSizes(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	value, err := codec.Parse("123.31232")
	if err != nil || codec.FractionalDecimalPlaces() != 5 || value.Units() != 12_331_232 {
		t.Fatalf("codec = %v %d %d", err, codec.FractionalDecimalPlaces(), value.Units())
	}
	formatted := codec.FromUnits(42)
	if got := string(codec.AppendJSON(make([]byte, 0, 16), formatted)); got != `"0.00042"` {
		t.Fatalf("AppendJSON = %q", got)
	}

	if strconv.IntSize == 64 {
		if unsafe.Sizeof(price5{}) != 24 {
			t.Fatalf("uint64 FixedDecimal size = %d, want 24", unsafe.Sizeof(price5{}))
		}
		if unsafe.Sizeof(wide18{}) != 48 {
			t.Fatalf("uint256 FixedDecimal size = %d, want 48", unsafe.Sizeof(wide18{}))
		}
		if unsafe.Sizeof(codec) != 1 {
			t.Fatalf("FixedDecimalCodec size = %d, want 1", unsafe.Sizeof(codec))
		}
	}
}

func TestInvalidScaleCodec(t *testing.T) {
	t.Parallel()

	if _, err := NewFixedDecimalCodec[uint64DecimalPlaces20](); !errors.Is(err, ErrUnsupportedFractionalDecimalPlaces) {
		t.Fatalf("NewFixedDecimalCodec error = %v", err)
	}
}

func TestZeroCodecUsesCompileTimeScale(t *testing.T) {
	t.Parallel()

	var codec FixedDecimalCodec[PriceInUint64Units[DecimalPlaces5], uint64]
	value, err := codec.Parse("1.00000")
	if err != nil || codec.FractionalDecimalPlaces() != 5 || value.Units() != 100_000 {
		t.Fatalf("zero codec = %v, scale %d, units %d", err, codec.FractionalDecimalPlaces(), value.Units())
	}
}
