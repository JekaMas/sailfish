package sailfish

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/holiman/uint256"
)

func TestDecimalStructLayoutIsPaddingMinimal(t *testing.T) {
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
		{name: "uint8", size: unsafe.Sizeof(Decimal[PriceUint8[Fraction1], uint8]{}), align: unsafe.Alignof(Decimal[PriceUint8[Fraction1], uint8]{}), unitsOffset: unsafe.Offsetof((Decimal[PriceUint8[Fraction1], uint8]{}).units), representationOffset: unsafe.Offsetof((Decimal[PriceUint8[Fraction1], uint8]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint16", size: unsafe.Sizeof(Decimal[PriceUint16[Fraction2], uint16]{}), align: unsafe.Alignof(Decimal[PriceUint16[Fraction2], uint16]{}), unitsOffset: unsafe.Offsetof((Decimal[PriceUint16[Fraction2], uint16]{}).units), representationOffset: unsafe.Offsetof((Decimal[PriceUint16[Fraction2], uint16]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint32", size: unsafe.Sizeof(Decimal[PriceUint32[Fraction5], uint32]{}), align: unsafe.Alignof(Decimal[PriceUint32[Fraction5], uint32]{}), unitsOffset: unsafe.Offsetof((Decimal[PriceUint32[Fraction5], uint32]{}).units), representationOffset: unsafe.Offsetof((Decimal[PriceUint32[Fraction5], uint32]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint64", size: unsafe.Sizeof(Decimal[PriceUint64[Fraction5], uint64]{}), align: unsafe.Alignof(Decimal[PriceUint64[Fraction5], uint64]{}), unitsOffset: unsafe.Offsetof((Decimal[PriceUint64[Fraction5], uint64]{}).units), representationOffset: unsafe.Offsetof((Decimal[PriceUint64[Fraction5], uint64]{}).representation), wantSize: 24, wantRepresentation: 8},
		{name: "uint256", size: unsafe.Sizeof(Decimal[AmountUint256[Fraction18], uint256.Int]{}), align: unsafe.Alignof(Decimal[AmountUint256[Fraction18], uint256.Int]{}), unitsOffset: unsafe.Offsetof((Decimal[AmountUint256[Fraction18], uint256.Int]{}).units), representationOffset: unsafe.Offsetof((Decimal[AmountUint256[Fraction18], uint256.Int]{}).representation), wantSize: 48, wantRepresentation: 32},
	}

	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.align != 8 || tt.unitsOffset != 0 || tt.representationOffset != tt.wantRepresentation {
			t.Errorf("%s layout: size=%d align=%d units=%d representation=%d; want size=%d align=8 units=0 representation=%d", tt.name, tt.size, tt.align, tt.unitsOffset, tt.representationOffset, tt.wantSize, tt.wantRepresentation)
		}
	}
}

func TestCodecAndFormatLayoutsRemainMinimal(t *testing.T) {
	t.Parallel()

	if size := unsafe.Sizeof(testCodec[PriceUint64[Fraction5]]()); size != 1 {
		t.Fatalf("Codec size = %d, want 1", size)
	}
	if align := unsafe.Alignof(testCodec[PriceUint64[Fraction5]]()); align != 1 {
		t.Fatalf("Codec alignment = %d, want 1", align)
	}
	if size := unsafe.Sizeof(PriceUint64[Fraction5]{}); size != 0 {
		t.Fatalf("format size = %d, want 0", size)
	}
	if size := unsafe.Sizeof(testUint256Codec(18)); size != 1 {
		t.Fatalf("Uint256Codec size = %d, want 1", size)
	}
}
