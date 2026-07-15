package sailfish

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/holiman/uint256"
)

func TestParseUint64ChunkDenseRuns(t *testing.T) {
	t.Parallel()

	const digits = "1234567890123456789"
	for length := 1; length <= len(digits); length++ {
		input := digits[:length]
		want, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		got, parseErr := parseUint64Chunk(input, 0, len(input))
		if parseErr != "" || got != want {
			t.Fatalf("string length %d = %d, %v; want %d", length, got, parseErr, want)
		}
		got, parseErr = parseUint64Chunk([]byte(input), 0, len(input))
		if parseErr != "" || got != want {
			t.Fatalf("bytes length %d = %d, %v; want %d", length, got, parseErr, want)
		}
		if length >= 8 {
			got, parseErr = parseUint64DenseChunk(input, 0, len(input))
			if parseErr != "" || got != want {
				t.Fatalf("dense string length %d = %d, %v; want %d", length, got, parseErr, want)
			}
			got, parseErr = parseUint64DenseChunk([]byte(input), 0, len(input))
			if parseErr != "" || got != want {
				t.Fatalf("dense bytes length %d = %d, %v; want %d", length, got, parseErr, want)
			}
		}
	}
}

func TestParseUint64ChunkRejectsEveryInvalidPosition(t *testing.T) {
	t.Parallel()

	for length := 1; length <= 19; length++ {
		for invalidAt := range length {
			raw := []byte("1234567890123456789"[:length])
			raw[invalidAt] = 'x'
			if _, err := parseUint64Chunk(raw, 0, len(raw)); err != ErrSyntax {
				t.Fatalf("length %d invalid position %d error = %v", length, invalidAt, err)
			}
			if _, err := parseUint64Chunk(string(raw), 0, len(raw)); err != ErrSyntax {
				t.Fatalf("string length %d invalid position %d error = %v", length, invalidAt, err)
			}
			if length >= 8 {
				if _, err := parseUint64DenseChunk(raw, 0, len(raw)); err != ErrSyntax {
					t.Fatalf("dense length %d invalid position %d error = %v", length, invalidAt, err)
				}
				if _, err := parseUint64DenseChunk(string(raw), 0, len(raw)); err != ErrSyntax {
					t.Fatalf("dense string length %d invalid position %d error = %v", length, invalidAt, err)
				}
			}
		}
	}
}

func TestParseUint64KnownDotSWAR(t *testing.T) {
	t.Parallel()

	const digits = "1234567890123456"
	for _, count := range [...]int{8, 16} {
		want, err := strconv.ParseUint(digits[:count], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		for dot := 1; dot < count; dot++ {
			input := digits[:dot] + "." + digits[dot:count]
			scale := count - dot
			for _, raw := range []decimalTestInput{
				{name: "string", parse: func() (uint64, Error) { value, _, parseErr := parseUint64(input, scale); return value, parseErr }},
				{name: "bytes", parse: func() (uint64, Error) {
					value, _, parseErr := parseUint64([]byte(input), scale)
					return value, parseErr
				}},
			} {
				got, parseErr := raw.parse()
				if parseErr != "" || got != want {
					t.Fatalf("%s digits=%d dot=%d = (%d, %v), want %d", raw.name, count, dot, got, parseErr, want)
				}
			}

			for invalidAt := range len(input) {
				if invalidAt == dot {
					continue
				}
				invalid := []byte(input)
				invalid[invalidAt] = 'x'
				if _, _, parseErr := parseUint64(invalid, scale); parseErr != ErrSyntax {
					t.Fatalf("digits=%d dot=%d invalid=%d error=%v", count, dot, invalidAt, parseErr)
				}
			}
		}
	}
}

type decimalTestInput struct {
	name  string
	parse func() (uint64, Error)
}

func TestErrorsAreTypedConstants(t *testing.T) {
	t.Parallel()

	const copied Error = ErrSyntax
	var _ error = copied

	_, err := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("")
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

			got, err := NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]](tt.input)
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
		if _, err := NewFixedDecimal[PriceInUint64Units[DecimalPlaces1]](input); !errors.Is(err, ErrSyntax) {
			t.Fatalf("byte 0x%02x accepted by string parser: %v", b, err)
		}
		if _, err := NewFixedDecimalFromBytes[PriceInUint64Units[DecimalPlaces1]]([]byte(input)); !errors.Is(err, ErrSyntax) {
			t.Fatalf("byte 0x%02x accepted by byte parser: %v", b, err)
		}
	}
}

func TestUint64ScaleBoundaries(t *testing.T) {
	t.Parallel()

	max0, err := NewFixedDecimal[uint64DecimalPlaces0]("18446744073709551615")
	if err != nil || max0.Units() != math.MaxUint64 {
		t.Fatalf("scale 0 max: units=%d err=%v", max0.Units(), err)
	}
	max19, err := NewFixedDecimal[uint64DecimalPlaces19]("1.8446744073709551615")
	if err != nil || max19.Units() != math.MaxUint64 {
		t.Fatalf("scale 19 max: units=%d err=%v", max19.Units(), err)
	}
	if _, err := NewFixedDecimal[uint64DecimalPlaces20]("1"); !errors.Is(err, ErrUnsupportedFractionalDecimalPlaces) {
		t.Fatalf("scale 20 error = %v, want %v", err, ErrUnsupportedFractionalDecimalPlaces)
	}
}

func TestUint256ParseBoundaries(t *testing.T) {
	t.Parallel()

	max, err := NewFixedDecimal[uint256DecimalPlaces0](maxUint256Decimal)
	if err != nil {
		t.Fatal(err)
	}
	want := uint256.Int{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64}
	if max.Units() != want || max.String() != maxUint256Decimal {
		t.Fatalf("max = %#v %q", max.Units(), max.String())
	}
	if _, err := NewFixedDecimal[uint256DecimalPlaces0](maxUint256PlusOne); !errors.Is(err, ErrRange) {
		t.Fatalf("max+1 error = %v, want %v", err, ErrRange)
	}
	if _, err := NewFixedDecimal[uint256DecimalPlaces0](strings.Repeat("9", 200)); !errors.Is(err, ErrRange) {
		t.Fatalf("long overflow error = %v, want %v", err, ErrRange)
	}

	point := len(maxUint256Decimal) - 18
	text18 := maxUint256Decimal[:point] + "." + maxUint256Decimal[point:]
	scaled, err := NewFixedDecimal[uint256DecimalPlaces18](text18)
	if err != nil || scaled.String() != text18 {
		t.Fatalf("scale 18 max = %v %q", err, scaled.String())
	}

	one, err := NewFixedDecimal[uint256DecimalPlaces77]("1")
	if err != nil || one.String() != "1.00000000000000000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("scale 77 one = %v %q", err, one.String())
	}
	if _, err := NewFixedDecimal[uint256DecimalPlaces78]("1"); !errors.Is(err, ErrUnsupportedFractionalDecimalPlaces) {
		t.Fatalf("scale 78 error = %v, want %v", err, ErrUnsupportedFractionalDecimalPlaces)
	}
}

func TestStringAndBytesParsersAgree(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"0", "1", "1.2", "001.20000", "184467440737095.51615",
		".5", "5.", "-1", "+1", "1e2", "1 2", "1.123456",
	}
	for _, input := range inputs {
		fromString, stringErr := NewCompactFixedDecimal[PriceInUint64Units[DecimalPlaces5]](input)
		fromBytes, bytesErr := NewFixedDecimalFromBytes[PriceInUint64Units[DecimalPlaces5]]([]byte(input))
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

func TestPriceInUint64UnitsDecimalPlaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
	}{
		{name: "scale1", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces1]]("1"))},
		{name: "scale2", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces2]]("1"))},
		{name: "scale3", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces3]]("1"))},
		{name: "scale4", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces4]]("1"))},
		{name: "scale5", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces5]]("1"))},
		{name: "scale6", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces6]]("1"))},
		{name: "scale7", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces7]]("1"))},
		{name: "scale8", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces8]]("1"))},
		{name: "scale9", got: mustText(NewFixedDecimal[PriceInUint64Units[DecimalPlaces9]]("1"))},
	}
	for i, tt := range tests {
		want := "1." + strings.Repeat("0", i+1)
		if tt.got != want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, want)
		}
	}
}

func mustText[V FixedDecimalFormat[U], U Unit](d FixedDecimal[V, U], err error) string {
	if err != nil {
		return err.Error()
	}
	return d.String()
}
