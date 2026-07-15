package sailfish

import (
	"encoding"
	"errors"
	"testing"

	"github.com/goccy/go-json"
	"github.com/holiman/uint256"
)

func TestTextEncodingRoundTrip(t *testing.T) {
	t.Parallel()

	value, _ := New[PriceUint64[Fraction5]]("12.30000")
	var _ encoding.TextMarshaler = value
	var _ encoding.TextUnmarshaler = (*price5)(nil)

	text, err := value.MarshalText()
	if err != nil || string(text) != "12.30000" {
		t.Fatalf("MarshalText = %q, %v", text, err)
	}

	var decoded price5
	if err := decoded.UnmarshalText(text); err != nil || !decoded.Equal(value) {
		t.Fatalf("UnmarshalText = %q, %v", decoded.String(), err)
	}
	text[0] = '9'
	if decoded.String() != "12.30000" {
		t.Fatalf("decoded retained input bytes: %q", decoded.String())
	}
}

func TestJSONEncodingRoundTrip(t *testing.T) {
	t.Parallel()

	value, _ := New[PriceUint64[Fraction5]]("12.30000")
	encoded, err := json.Marshal(value)
	if err != nil || string(encoded) != `"12.30000"` {
		t.Fatalf("MarshalJSON = %q, %v", encoded, err)
	}

	for _, input := range []string{`"12.30000"`, `"\u0031\u0032.30000"`} {
		var decoded price5
		if err := json.Unmarshal([]byte(input), &decoded); err != nil || !decoded.Equal(value) {
			t.Fatalf("UnmarshalJSON(%s) = %q, %v", input, decoded.String(), err)
		}
	}
}

func TestJSONWideMaximumRoundTrip(t *testing.T) {
	t.Parallel()

	codec := testCodec[uint256Scale18]()
	value := codec.FromUnits(uint256.Int{
		^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0),
	})
	wire, err := value.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(wire),
		`"115792089237316195423570985008687907853269984665640564039457.584007913129639935"`; got != want {
		t.Fatalf("MarshalJSON = %q, want %q", got, want)
	}

	var decoded wide18
	if err := decoded.UnmarshalJSON(wire); err != nil {
		t.Fatal(err)
	}
	if decoded.Units() != value.Units() {
		t.Fatalf("UnmarshalJSON units = %#v, want %#v", decoded.Units(), value.Units())
	}
}

func TestJSONRejectsNonStringAndInvalidDecimal(t *testing.T) {
	t.Parallel()

	initial, _ := New[PriceUint64[Fraction5]]("1.00000")
	inputs := []string{
		`12.3`,
		`null`,
		`true`,
		`" 12.30000"`,
		`"-1.00000"`,
		`"12.3000x"`,
		`"12.30000\q"`,
		`"\u0031\u0032.3000x"`,
		`"\u0031\u0032.30000`,
	}
	for _, input := range inputs {
		value := initial
		if err := json.Unmarshal([]byte(input), &value); err == nil {
			t.Fatalf("UnmarshalJSON(%s) unexpectedly succeeded", input)
		}
		if !value.Equal(initial) || value.String() != initial.String() {
			t.Fatalf("UnmarshalJSON(%s) changed receiver", input)
		}
	}
}

func TestJSONUnmarshalEscapedDecimal(t *testing.T) {
	t.Parallel()

	var value price5
	if err := value.UnmarshalJSON([]byte(`"\u0031\u0032\u002e30000"`)); err != nil {
		t.Fatal(err)
	}
	if got, want := value.String(), "12.30000"; got != want {
		t.Fatalf("UnmarshalJSON = %q, want %q", got, want)
	}
}

func TestAppendTextAndJSON(t *testing.T) {
	t.Parallel()

	value := testCodec[PriceUint64[Fraction5]]().FromUnits(1_230_000)
	text, err := value.AppendText(make([]byte, 0, 16))
	if err != nil || string(text) != "12.30000" {
		t.Fatalf("AppendText = %q, %v", text, err)
	}
	if got := string(value.AppendJSON(make([]byte, 0, 18))); got != `"12.30000"` {
		t.Fatalf("AppendJSON = %q", got)
	}
}

func TestUnmarshalPreservesReceiverOnError(t *testing.T) {
	t.Parallel()

	value, _ := New[PriceUint64[Fraction5]]("1.00000")
	before := value
	if err := value.UnmarshalText([]byte("bad")); !errors.Is(err, ErrSyntax) {
		t.Fatalf("UnmarshalText error = %v", err)
	}
	if !value.Equal(before) || value.String() != before.String() {
		t.Fatal("UnmarshalText changed receiver on error")
	}
	if err := value.UnmarshalJSON([]byte(`"bad"`)); !errors.Is(err, ErrSyntax) {
		t.Fatalf("UnmarshalJSON error = %v, want %v", err, ErrSyntax)
	}
	if !value.Equal(before) || value.String() != before.String() {
		t.Fatal("UnmarshalJSON changed receiver on error")
	}
}
