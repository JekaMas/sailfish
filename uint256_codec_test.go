package sailfish

import (
	"math/rand/v2"
	"testing"

	"github.com/holiman/uint256"
)

func TestUint256CodecRuntimeScaleAPI(t *testing.T) {
	t.Parallel()

	codec, err := NewUint256Codec(6)
	if err != nil {
		t.Fatal(err)
	}
	if codec.Scale() != 6 {
		t.Fatalf("Scale = %d", codec.Scale())
	}

	value, parseErr := codec.Parse("123.456789")
	if parseErr != "" || value != (uint256.Int{123_456_789}) {
		t.Fatalf("Parse = %#v, %v", value, parseErr)
	}
	fromBytes, parseErr := codec.ParseBytes([]byte("123.456789"))
	if parseErr != "" || fromBytes != value {
		t.Fatalf("ParseBytes = %#v, %v", fromBytes, parseErr)
	}

	buffer := make([]byte, 0, codec.Len(value))
	if got := string(codec.AppendTo(buffer, value)); got != "123.456789" {
		t.Fatalf("AppendTo = %q", got)
	}
}

func TestUint256CodecParseIntoPreservesDestinationOnError(t *testing.T) {
	t.Parallel()

	codec := MustUint256Codec(6)
	destination := uint256.Int{9, 8, 7, 6}
	want := destination
	if err := codec.ParseInto("bad", &destination); err != ErrSyntax {
		t.Fatalf("ParseInto error = %v", err)
	}
	if destination != want {
		t.Fatalf("destination changed to %#v", destination)
	}
	if err := codec.ParseBytesInto([]byte("bad"), &destination); err != ErrSyntax {
		t.Fatalf("ParseBytesInto error = %v", err)
	}
	if destination != want {
		t.Fatalf("destination changed to %#v", destination)
	}
	if err := codec.ParseInto("1.000000", nil); err != ErrNilDestination {
		t.Fatalf("nil destination error = %v", err)
	}
}

func TestUint256CodecScaleValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewUint256Codec(77); err != nil {
		t.Fatalf("scale 77: %v", err)
	}
	if _, err := NewUint256Codec(78); err != ErrScale {
		t.Fatalf("scale 78 error = %v", err)
	}

	defer func() {
		if recovered := recover(); recovered != ErrUninitializedCodec {
			t.Fatalf("zero codec panic = %v", recovered)
		}
	}()
	var zero Uint256Codec
	_ = zero.Scale()
}

func TestUint256CodecRoundTripAllScalesAndLimbs(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(0x51a1f15, 0xc0dec))
	for scale := 0; scale <= maxUint256Scale; scale++ {
		codec := MustUint256Codec(Notion(scale))
		for sample := 0; sample < 256; sample++ {
			want := uint256.Int{
				random.Uint64(), random.Uint64(), random.Uint64(), random.Uint64(),
			}
			text := codec.AppendTo(make([]byte, 0, maxUint256TextLen), want)
			got, err := codec.ParseBytes(text)
			if err != "" || got != want {
				t.Fatalf("scale=%d value=%#v text=%q: got=%#v err=%v", scale, want, text, got, err)
			}
		}
	}
}

func TestUint256CodecAllocations(t *testing.T) {
	codec := MustUint256Codec(6)
	buffer := make([]byte, 0, 32)
	var value uint256.Int

	assertAllocs(t, "runtime codec parse", 0, func() {
		value, _ = codec.Parse("123.456789")
	})
	assertAllocs(t, "runtime codec parse into", 0, func() {
		_ = codec.ParseInto("123.456789", &value)
	})
	assertAllocs(t, "runtime codec append", 0, func() {
		allocationBytesSink = codec.AppendTo(buffer[:0], value)
	})
}
