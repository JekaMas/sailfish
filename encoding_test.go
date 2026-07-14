package sailfish

import (
	"encoding"
	"errors"
	"testing"

	"github.com/goccy/go-json"
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

func TestJSONRejectsNonStringAndInvalidDecimal(t *testing.T) {
	t.Parallel()

	for _, input := range []string{`12.3`, `null`, `true`, `" 12.30000"`, `"-1.00000"`} {
		var value price5
		if err := json.Unmarshal([]byte(input), &value); err == nil {
			t.Fatalf("UnmarshalJSON(%s) unexpectedly succeeded", input)
		}
	}
}

func TestAppendTextAndJSON(t *testing.T) {
	t.Parallel()

	value := MustCodec[PriceUint64[Fraction5]]().FromUnits(1_230_000)
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
}
