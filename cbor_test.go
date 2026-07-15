package sailfish

import (
	"encoding/hex"
	"errors"
	"math"
	"math/rand"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/holiman/uint256"
)

type cborQuote struct {
	_ struct{} `cbor:",toarray"`

	Price  FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64]
	Amount FixedDecimal[AmountInUint256Units[DecimalPlaces18], uint256.Int]
}

func TestCBORNativePreferredSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value uint64
		wire  string
	}{
		{name: "zero", value: 0, wire: "00"},
		{name: "inline max", value: 23, wire: "17"},
		{name: "uint8 min", value: 24, wire: "1818"},
		{name: "uint8 max", value: math.MaxUint8, wire: "18ff"},
		{name: "uint16 min", value: math.MaxUint8 + 1, wire: "190100"},
		{name: "uint16 max", value: math.MaxUint16, wire: "19ffff"},
		{name: "uint32 min", value: math.MaxUint16 + 1, wire: "1a00010000"},
		{name: "uint32 max", value: math.MaxUint32, wire: "1affffffff"},
		{name: "uint64 min", value: math.MaxUint32 + 1, wire: "1b0000000100000000"},
		{name: "uint64 max", value: math.MaxUint64, wire: "1bffffffffffffffff"},
	}

	codec := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := codec.FromUnits(tt.value)
			if got := hex.EncodeToString(value.AppendCBOR(nil)); got != tt.wire {
				t.Fatalf("AppendCBOR = %s, want %s", got, tt.wire)
			}
			if got := value.CBORLen(); got != len(tt.wire)/2 {
				t.Fatalf("CBORLen = %d, want %d", got, len(tt.wire)/2)
			}

			wire, err := hex.DecodeString(tt.wire)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := codec.ParseCBOR(wire)
			if err != nil || decoded.Units() != tt.value {
				t.Fatalf("ParseCBOR = %d, %v, want %d", decoded.Units(), err, tt.value)
			}
		})
	}
}

func TestCBORNarrowUnitsEnforceRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		decode func([]byte) error
	}{
		{
			name: "uint8",
			decode: func(raw []byte) error {
				var value FixedDecimal[PriceInUint8Units[DecimalPlaces1], uint8]
				return value.UnmarshalCBOR(raw)
			},
		},
		{
			name: "uint16",
			decode: func(raw []byte) error {
				var value FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]
				return value.UnmarshalCBOR(raw)
			},
		},
		{
			name: "uint32",
			decode: func(raw []byte) error {
				var value FixedDecimal[PriceInUint32Units[DecimalPlaces5], uint32]
				return value.UnmarshalCBOR(raw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw []byte
			switch tt.name {
			case "uint8":
				raw = []byte{0x19, 0x01, 0x00}
			case "uint16":
				raw = []byte{0x1a, 0x00, 0x01, 0x00, 0x00}
			default:
				raw = []byte{0x1b, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
			}
			if err := tt.decode(raw); !errors.Is(err, ErrRange) {
				t.Fatalf("error = %v, want %v", err, ErrRange)
			}
		})
	}
}

func TestCBORUint256PreferredSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		units uint256.Int
		wire  string
	}{
		{name: "zero", units: uint256.Int{}, wire: "00"},
		{name: "uint64 max", units: uint256.Int{math.MaxUint64}, wire: "1bffffffffffffffff"},
		{name: "two to 64", units: uint256.Int{0, 1}, wire: "c249010000000000000000"},
		{
			name:  "maximum",
			units: uint256.Int{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64},
			wire:  "c25820ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		},
	}

	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := codec.FromUnits(tt.units)
			if got := hex.EncodeToString(value.AppendCBOR(nil)); got != tt.wire {
				t.Fatalf("AppendCBOR = %s, want %s", got, tt.wire)
			}
			wire, err := hex.DecodeString(tt.wire)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := codec.ParseCBOR(wire)
			if err != nil || decoded.Units() != tt.units {
				t.Fatalf("ParseCBOR = %#v, %v, want %#v", decoded.Units(), err, tt.units)
			}
		})
	}
}

func TestCBORRoundTripsEveryUint256MagnitudeWidth(t *testing.T) {
	t.Parallel()

	codec := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]()
	for byteLen := 9; byteLen <= 32; byteLen++ {
		var magnitude [32]byte
		magnitude[len(magnitude)-byteLen] = 1
		for i := len(magnitude) - byteLen + 1; i < len(magnitude); i++ {
			magnitude[i] = byte(i*37 + byteLen)
		}
		var units uint256.Int
		units.SetBytes(magnitude[len(magnitude)-byteLen:])
		value := codec.FromUnits(units)
		var buffer [MaxCBORSize]byte
		wire := value.AppendCBOR(buffer[:0])
		decoded, err := codec.ParseCBOR(wire)
		if err != nil || decoded.Units() != units {
			t.Fatalf("byte length %d: %#v -> %x -> %#v, %v", byteLen, units, wire, decoded.Units(), err)
		}
	}
}

func TestCBORRoundTripsNativeWidthCrossProducts(t *testing.T) {
	t.Parallel()

	for units := 0; units <= math.MaxUint8; units++ {
		assertCBORRoundTrip(t, testFixedDecimalCodec[PriceInUint8Units[DecimalPlaces2]](), uint8(units))
	}

	rng := rand.New(rand.NewSource(0xcb0_2026))
	for range 10_000 {
		assertCBORRoundTrip(t, testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces4]](), uint16(rng.Uint32()))
		assertCBORRoundTrip(t, testFixedDecimalCodec[PriceInUint32Units[DecimalPlaces9]](), rng.Uint32())
		assertCBORRoundTrip(t, testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces19]](), rng.Uint64())
	}
}

func assertCBORRoundTrip[V FixedDecimalFormat[U], U Unit](t *testing.T, codec FixedDecimalCodec[V, U], units U) {
	t.Helper()

	value := codec.FromUnits(units)
	var buffer [MaxCBORSize]byte
	wire := codec.AppendCBOR(buffer[:0], value)
	if got := codec.CBORLen(value); got != len(wire) {
		t.Fatalf("CBORLen(%v) = %d, want %d", units, got, len(wire))
	}
	decoded, err := codec.ParseCBOR(wire)
	if err != nil || decoded.Units() != units {
		t.Fatalf("%v -> %x -> %v, %v", units, wire, decoded.Units(), err)
	}
}

func TestCBORRejectsNonDeterministicOrMalformedInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		wire []byte
		err  error
	}{
		{name: "empty", wire: nil, err: ErrCBORSyntax},
		{name: "negative", wire: []byte{0x20}, err: ErrCBORSyntax},
		{name: "text", wire: []byte{0x61, '1'}, err: ErrCBORSyntax},
		{name: "array", wire: []byte{0x81, 0x01}, err: ErrCBORSyntax},
		{name: "indefinite", wire: []byte{0x5f, 0xff}, err: ErrCBORSyntax},
		{name: "reserved additional info", wire: []byte{0x1c}, err: ErrCBORSyntax},
		{name: "truncated", wire: []byte{0x19, 0x01}, err: ErrCBORSyntax},
		{name: "trailing", wire: []byte{0x01, 0x00}, err: ErrCBORSyntax},
		{name: "long uint8", wire: []byte{0x18, 0x17}, err: ErrCBORNonDeterministic},
		{name: "long uint16", wire: []byte{0x19, 0x00, 0xff}, err: ErrCBORNonDeterministic},
		{name: "long uint32", wire: []byte{0x1a, 0x00, 0x00, 0xff, 0xff}, err: ErrCBORNonDeterministic},
		{name: "long uint64", wire: []byte{0x1b, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff}, err: ErrCBORNonDeterministic},
		{name: "small bignum", wire: []byte{0xc2, 0x41, 0x01}, err: ErrCBORNonDeterministic},
		{name: "empty bignum", wire: []byte{0xc2, 0x40}, err: ErrCBORNonDeterministic},
		{name: "truncated bignum tag", wire: []byte{0xc2}, err: ErrCBORSyntax},
		{name: "wrong bignum payload", wire: []byte{0xc2, 0x61, '1'}, err: ErrCBORSyntax},
		{name: "truncated bignum payload", wire: []byte{0xc2, 0x49, 0x01}, err: ErrCBORSyntax},
		{name: "unsupported bignum length width", wire: []byte{0xc2, 0x59, 0, 9}, err: ErrCBORSyntax},
		{name: "leading zero bignum", wire: []byte{0xc2, 0x49, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: ErrCBORNonDeterministic},
		{name: "long bytes length", wire: append([]byte{0xc2, 0x58, 0x09}, make([]byte, 9)...), err: ErrCBORNonDeterministic},
		{name: "oversized bignum", wire: append([]byte{0xc2, 0x58, 0x21, 0x01}, make([]byte, 32)...), err: ErrRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]().FromUnits(uint256.Int{42})
			before := value
			if err := value.UnmarshalCBOR(tt.wire); !errors.Is(err, tt.err) {
				t.Fatalf("UnmarshalCBOR error = %v, want %v", err, tt.err)
			}
			if !value.Equal(before) || value.HasRepresentation() != before.HasRepresentation() {
				t.Fatal("UnmarshalCBOR changed receiver on error")
			}
		})
	}
}

func TestCBORRejectsUnsupportedFormatScale(t *testing.T) {
	t.Parallel()

	invalid := FixedDecimal[uint64DecimalPlaces20, uint64]{}
	if _, err := invalid.MarshalCBOR(); !errors.Is(err, ErrUnsupportedFractionalDecimalPlaces) {
		t.Fatalf("MarshalCBOR error = %v, want %v", err, ErrUnsupportedFractionalDecimalPlaces)
	}
	if err := invalid.UnmarshalCBOR([]byte{0}); !errors.Is(err, ErrUnsupportedFractionalDecimalPlaces) {
		t.Fatalf("UnmarshalCBOR error = %v, want %v", err, ErrUnsupportedFractionalDecimalPlaces)
	}
}

func TestCBORToArrayInteroperabilityIsCompact(t *testing.T) {
	t.Parallel()

	price := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces5]]().FromUnits(12_331_232)
	amount := testFixedDecimalCodec[AmountInUint256Units[DecimalPlaces18]]().FromUnits(uint256.Int{0, 1})
	original := cborQuote{Price: price, Amount: amount}

	wire, err := cbor.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	const want = "821a00bc28e0c249010000000000000000"
	if got := hex.EncodeToString(wire); got != want {
		t.Fatalf("toarray wire = %s, want %s", got, want)
	}

	var decoded cborQuote
	if err = cbor.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Price.Equal(price) || !decoded.Amount.Equal(amount) {
		t.Fatalf("decoded = %#v", decoded)
	}
	if decoded.Price.HasRepresentation() || decoded.Amount.HasRepresentation() {
		t.Fatal("CBOR decode retained non-numeric representation state")
	}
}

func TestCBORMarshalInterfacesUseCompactScalar(t *testing.T) {
	t.Parallel()

	value := testFixedDecimalCodec[PriceInUint16Units[DecimalPlaces2]]().FromUnits(65_535)
	wire, err := value.MarshalCBOR()
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(wire); got != "19ffff" {
		t.Fatalf("MarshalCBOR = %s", got)
	}

	var decoded FixedDecimal[PriceInUint16Units[DecimalPlaces2], uint16]
	if err = decoded.UnmarshalCBOR(wire); err != nil {
		t.Fatal(err)
	}
	if decoded.Units() != 65_535 {
		t.Fatalf("decoded units = %d", decoded.Units())
	}
}

func TestCBORCarriesUnitsWhileTypeCarriesScale(t *testing.T) {
	t.Parallel()

	one := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces1]]().FromUnits(123)
	nine := testFixedDecimalCodec[PriceInUint64Units[DecimalPlaces9]]().FromUnits(123)
	if got, want := hex.EncodeToString(one.AppendCBOR(nil)), hex.EncodeToString(nine.AppendCBOR(nil)); got != want {
		t.Fatalf("equal units encoded differently: %s != %s", got, want)
	}
	if one.String() == nine.String() {
		t.Fatalf("different typed scales formatted equally: %q", one.String())
	}
}

func TestUint256RuntimeCodecCBORRoundTrip(t *testing.T) {
	t.Parallel()

	codec := testUint256FixedDecimalCodec(18)
	units := uint256.Int{1, 2, 3, 4}
	var buffer [MaxCBORSize]byte
	wire := codec.AppendCBOR(buffer[:0], units)
	if codec.CBORLen(units) != len(wire) {
		t.Fatalf("CBORLen = %d, wire length = %d", codec.CBORLen(units), len(wire))
	}
	decoded, err := codec.ParseCBOR(wire)
	if err != "" || decoded != units {
		t.Fatalf("ParseCBOR = %#v, %v", decoded, err)
	}
	var into uint256.Int
	if err = codec.ParseCBORInto(wire, &into); err != "" || into != units {
		t.Fatalf("ParseCBORInto = %#v, %v", into, err)
	}
	if err = codec.ParseCBORInto(wire, nil); err != ErrNilDestination {
		t.Fatalf("ParseCBORInto nil error = %v", err)
	}
}
