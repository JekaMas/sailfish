package sailfish

import (
	"errors"
	"strconv"
	"testing"
	"unsafe"
)

func TestCodecAPIAndSizes(t *testing.T) {
	t.Parallel()

	codec := MustCodec[PriceUint64[Fraction5]]()
	value, err := codec.Parse("123.31232")
	if err != nil || codec.Scale() != 5 || value.Units() != 12_331_232 {
		t.Fatalf("codec = %v %d %d", err, codec.Scale(), value.Units())
	}
	formatted := codec.FromUnits(42)
	if got := string(codec.AppendJSON(make([]byte, 0, 16), formatted)); got != `"0.00042"` {
		t.Fatalf("AppendJSON = %q", got)
	}

	if strconv.IntSize == 64 {
		if unsafe.Sizeof(price5{}) != 24 {
			t.Fatalf("uint64 Decimal size = %d, want 24", unsafe.Sizeof(price5{}))
		}
		if unsafe.Sizeof(wide18{}) != 48 {
			t.Fatalf("uint256 Decimal size = %d, want 48", unsafe.Sizeof(wide18{}))
		}
		if unsafe.Sizeof(codec) != 1 {
			t.Fatalf("Codec size = %d, want 1", unsafe.Sizeof(codec))
		}
	}
}

func TestInvalidScaleCodec(t *testing.T) {
	t.Parallel()

	if _, err := NewCodec[uint64Scale20](); !errors.Is(err, ErrScale) {
		t.Fatalf("NewCodec error = %v", err)
	}
}

func TestUninitializedCodecPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if got := recover(); got != ErrUninitializedCodec {
			t.Fatalf("panic = %#v, want %#v", got, ErrUninitializedCodec)
		}
	}()
	var codec Codec[PriceUint64[Fraction5], uint64]
	_, _ = codec.Parse("1.00000")
}
