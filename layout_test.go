package sailfish

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/holiman/uint256"
)

func TestFixedDecimalStructLayoutIsPaddingMinimal(t *testing.T) {
	t.Parallel()

	if strconv.IntSize != 64 {
		t.Skip("64-bit layout assertion")
	}

	tests := []struct {
		name                 string
		size                 uintptr
		align                uintptr
		unitsOffset          uintptr
		representationOffset uintptr
		wantSize             uintptr
		wantRepresentation   uintptr
	}{
		{name: "uint8", size: unsafe.Sizeof(FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]{}), align: unsafe.Alignof(FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]{}), unitsOffset: unsafe.Offsetof((FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]{}).units), representationOffset: unsafe.Offsetof((FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint16", size: unsafe.Sizeof(FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]{}), align: unsafe.Alignof(FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]{}), unitsOffset: unsafe.Offsetof((FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]{}).units), representationOffset: unsafe.Offsetof((FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint32", size: unsafe.Sizeof(FixedDecimal[PriceInUint32Units[DecimalPlaces5], uint32]{}), align: unsafe.Alignof(FixedDecimal[PriceInUint32Units[DecimalPlaces5], uint32]{}), unitsOffset: unsafe.Offsetof((FixedDecimal[PriceInUint32Units[DecimalPlaces5], uint32]{}).units), representationOffset: unsafe.Offsetof((FixedDecimal[PriceInUint32Units[DecimalPlaces5], uint32]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint64", size: unsafe.Sizeof(FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64]{}), align: unsafe.Alignof(FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64]{}), unitsOffset: unsafe.Offsetof((FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64]{}).units), representationOffset: unsafe.Offsetof((FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint256", size: unsafe.Sizeof(FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]{}), align: unsafe.Alignof(FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]{}), unitsOffset: unsafe.Offsetof((FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]{}).units), representationOffset: unsafe.Offsetof((FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]{}).representation), wantSize: 48, wantRepresentation: 32},
	}

	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.align != 8 || tt.unitsOffset != 0 || tt.representationOffset != tt.wantRepresentation {
			t.Errorf("%s layout: size=%d align=%d units=%d representation=%d; want size=%d align=8 units=0 representation=%d", tt.name, tt.size, tt.align, tt.unitsOffset, tt.representationOffset, tt.wantSize, tt.wantRepresentation)
		}
	}
}

func TestFixedDecimalCodecAndFormatLayoutsRemainMinimal(t *testing.T) {
	t.Parallel()

	if size := unsafe.Sizeof(testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()); size != 1 {
		t.Fatalf("FixedDecimalCodec size = %d, want 1", size)
	}
	if align := unsafe.Alignof(testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()); align != 1 {
		t.Fatalf("FixedDecimalCodec alignment = %d, want 1", align)
	}
	if size := unsafe.Sizeof(PriceInUint64Units[DecimalPlaces5]{}); size != 0 {
		t.Fatalf("format size = %d, want 0", size)
	}
	if size := unsafe.Sizeof(testUint256FixedDecimalCodec(18)); size != 1 {
		t.Fatalf("Uint256FixedDecimalCodec size = %d, want 1", size)
	}
}
