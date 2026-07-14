package sailfish

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/holiman/uint256"
)

func TestErrorsAreTypedConstants(t *testing.T) {
	t.Parallel()

	const copied Error = ErrSyntax
	var _ error = copied

	_, err := New[PriceUint64[Fraction5]]("")
	if !errors.Is(err, ErrSyntax) {
		t.Fatalf("errors.Is(%v, ErrSyntax) = false", err)
	}
}

func TestUint64ParseCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantUnits uint64
		wantText  string
		wantErr   Error
	}{
		{name: "zero integer", input: "0", wantText: "0.00000"},
		{name: "zero fraction", input: "0.0", wantText: "0.00000"},
		{name: "canonical", input: "123.31232", wantUnits: 12_331_232, wantText: "123.31232"},
		{name: "pad fraction", input: "12.3", wantUnits: 1_230_000, wantText: "12.30000"},
		{name: "leading zero normalization", input: "00012.3", wantUnits: 1_230_000, wantText: "12.30000"},
		{name: "max", input: "184467440737095.51615", wantUnits: math.MaxUint64, wantText: "184467440737095.51615"},
		{name: "empty", input: "", wantErr: ErrSyntax},
		{name: "leading dot", input: ".1", wantErr: ErrSyntax},
		{name: "trailing dot", input: "1.", wantErr: ErrSyntax},
		{name: "precision", input: "1.123456", wantErr: ErrPrecision},
		{name: "negative", input: "-1.00000", wantErr: ErrSyntax},
		{name: "plus", input: "+1.00000", wantErr: ErrSyntax},
		{name: "whitespace", input: " 1.00000", wantErr: ErrSyntax},
		{name: "exponent", input: "1e2", wantErr: ErrSyntax},
		{name: "invalid byte", input: "1x.00000", wantErr: ErrSyntax},
		{name: "second dot", input: "1.2.000", wantErr: ErrSyntax},
		{name: "range", input: "184467440737095.51616", wantErr: ErrRange},
		{name: "accumulator width overflow", input: "184467440737095516160", wantErr: ErrRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := New[PriceUint64[Fraction5]](tt.input)
			if tt.wantErr != "" {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("New(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("New(%q): %v", tt.input, err)
			}
			if got.Units() != tt.wantUnits {
				t.Fatalf("New(%q) units = %d, want %d", tt.input, got.Units(), tt.wantUnits)
			}
			if got.String() != tt.wantText {
				t.Fatalf("New(%q) text = %q, want %q", tt.input, got.String(), tt.wantText)
			}
		})
	}
}

func TestParserRejectsEveryNonDigitByte(t *testing.T) {
	t.Parallel()

	for raw := 0; raw <= 255; raw++ {
		b := byte(raw)
		if b >= '0' && b <= '9' {
			continue
		}

		input := string([]byte{'1', '.', b})
		if _, err := New[PriceUint64[Fraction1]](input); !errors.Is(err, ErrSyntax) {
			t.Fatalf("byte 0x%02x accepted by string parser: %v", b, err)
		}
		if _, err := NewBytes[PriceUint64[Fraction1]]([]byte(input)); !errors.Is(err, ErrSyntax) {
			t.Fatalf("byte 0x%02x accepted by byte parser: %v", b, err)
		}
	}
}

func TestUint64ScaleBoundaries(t *testing.T) {
	t.Parallel()

	max0, err := New[uint64Scale0]("18446744073709551615")
	if err != nil || max0.Units() != math.MaxUint64 {
		t.Fatalf("scale 0 max: units=%d err=%v", max0.Units(), err)
	}
	max19, err := New[uint64Scale19]("1.8446744073709551615")
	if err != nil || max19.Units() != math.MaxUint64 {
		t.Fatalf("scale 19 max: units=%d err=%v", max19.Units(), err)
	}
	if _, err := New[uint64Scale20]("1"); !errors.Is(err, ErrScale) {
		t.Fatalf("scale 20 error = %v, want %v", err, ErrScale)
	}
}

func TestUint256ParseBoundaries(t *testing.T) {
	t.Parallel()

	max, err := New[uint256Scale0](maxUint256Decimal)
	if err != nil {
		t.Fatal(err)
	}
	want := uint256.Int{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64}
	if max.Units() != want || max.String() != maxUint256Decimal {
		t.Fatalf("max = %#v %q", max.Units(), max.String())
	}
	if _, err := New[uint256Scale0](maxUint256PlusOne); !errors.Is(err, ErrRange) {
		t.Fatalf("max+1 error = %v, want %v", err, ErrRange)
	}
	if _, err := New[uint256Scale0](strings.Repeat("9", 200)); !errors.Is(err, ErrRange) {
		t.Fatalf("long overflow error = %v, want %v", err, ErrRange)
	}

	point := len(maxUint256Decimal) - 18
	text18 := maxUint256Decimal[:point] + "." + maxUint256Decimal[point:]
	scaled, err := New[uint256Scale18](text18)
	if err != nil || scaled.String() != text18 {
		t.Fatalf("scale 18 max = %v %q", err, scaled.String())
	}

	one, err := New[uint256Scale77]("1")
	if err != nil || one.String() != "1.00000000000000000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("scale 77 one = %v %q", err, one.String())
	}
	if _, err := New[uint256Scale78]("1"); !errors.Is(err, ErrScale) {
		t.Fatalf("scale 78 error = %v, want %v", err, ErrScale)
	}
}

func TestStringAndBytesParsersAgree(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"0", "1", "1.2", "001.20000", "184467440737095.51615",
		".5", "5.", "-1", "+1", "1e2", "1 2", "1.123456",
	}
	for _, input := range inputs {
		fromString, stringErr := NewCompact[PriceUint64[Fraction5]](input)
		fromBytes, bytesErr := NewBytes[PriceUint64[Fraction5]]([]byte(input))
		if !sameError(stringErr, bytesErr) {
			t.Fatalf("%q errors differ: string=%v bytes=%v", input, stringErr, bytesErr)
		}
		if stringErr == nil && !fromString.Equal(fromBytes) {
			t.Fatalf("%q values differ: %d != %d", input, fromString.Units(), fromBytes.Units())
		}
	}
}

func sameError(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Error() == b.Error()
}

func TestPriceUint64Fractions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
	}{
		{name: "scale1", got: mustText(New[PriceUint64[Fraction1]]("1"))},
		{name: "scale2", got: mustText(New[PriceUint64[Fraction2]]("1"))},
		{name: "scale3", got: mustText(New[PriceUint64[Fraction3]]("1"))},
		{name: "scale4", got: mustText(New[PriceUint64[Fraction4]]("1"))},
		{name: "scale5", got: mustText(New[PriceUint64[Fraction5]]("1"))},
		{name: "scale6", got: mustText(New[PriceUint64[Fraction6]]("1"))},
		{name: "scale7", got: mustText(New[PriceUint64[Fraction7]]("1"))},
		{name: "scale8", got: mustText(New[PriceUint64[Fraction8]]("1"))},
		{name: "scale9", got: mustText(New[PriceUint64[Fraction9]]("1"))},
	}
	for i, tt := range tests {
		want := "1." + strings.Repeat("0", i+1)
		if tt.got != want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, want)
		}
	}
}

func mustText[V Venue[U], U Unit](d Decimal[V, U], err error) string {
	if err != nil {
		panic(err)
	}
	return d.String()
}
