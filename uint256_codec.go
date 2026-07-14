package sailfish

import "github.com/holiman/uint256"

// Uint256Codec is the runtime-scale hot-path codec for scaled uint256 units.
//
// Use Codec with a Venue when compile-time venue identity is required. Use
// Uint256Codec at boundaries where trusted metadata resolves the scale at
// runtime, such as CEX symbol decoding. Its methods avoid generic venue
// dispatch and return Error directly so successful and rejected parses remain
// allocation-free.
//
// The one-byte scalePlusOne encoding reserves zero for an uninitialized codec.
type Uint256Codec struct {
	scalePlusOne uint8
}

// NewUint256Codec validates scale once for repeated uint256 operations.
func NewUint256Codec(scale Notion) (Uint256Codec, error) {
	if int(scale) > maxUint256Scale {
		return Uint256Codec{}, boxedErrScale
	}
	return Uint256Codec{scalePlusOne: uint8(scale + 1)}, nil
}

// MustUint256Codec is NewUint256Codec with panic-on-invalid configuration.
func MustUint256Codec(scale Notion) Uint256Codec {
	codec, err := NewUint256Codec(scale)
	if err != nil {
		panic(err)
	}
	return codec
}

func (c Uint256Codec) scale() int {
	if c.scalePlusOne == 0 {
		panic(boxedErrUninitializedCodec)
	}
	return int(c.scalePlusOne - 1)
}

// Scale returns the configured number of fractional decimal digits.
func (c Uint256Codec) Scale() Notion { return Notion(c.scale()) }

// MaxIntegerDigits reports the maximum integer-part digit count at this scale.
func (c Uint256Codec) MaxIntegerDigits() int {
	return unitDecimalDigits[uint256.Int]() - c.scale()
}

// Parse parses a strict non-negative decimal string into scaled units.
func (c Uint256Codec) Parse(input string) (uint256.Int, Error) {
	value, _, err := parseUint256(input, c.scale())
	return value, err
}

// ParseBytes parses input without converting it to a string.
func (c Uint256Codec) ParseBytes(input []byte) (uint256.Int, Error) {
	value, _, err := parseUint256(input, c.scale())
	return value, err
}

// ParseInto parses input into dst. It leaves dst unchanged on failure.
func (c Uint256Codec) ParseInto(input string, dst *uint256.Int) Error {
	scale := c.scale()
	if dst == nil {
		return ErrNilDestination
	}
	value, _, err := parseUint256(input, scale)
	if err != "" {
		return err
	}
	*dst = value
	return ""
}

// ParseBytesInto parses input into dst without converting it to a string. It
// leaves dst unchanged on failure.
func (c Uint256Codec) ParseBytesInto(input []byte, dst *uint256.Int) Error {
	scale := c.scale()
	if dst == nil {
		return ErrNilDestination
	}
	value, _, err := parseUint256(input, scale)
	if err != "" {
		return err
	}
	*dst = value
	return ""
}

// AppendTo appends canonical fixed-scale text for units. It allocates only
// when dst has insufficient capacity.
func (c Uint256Codec) AppendTo(dst []byte, units uint256.Int) []byte {
	return appendUint256Decimal(dst, units, c.scale())
}

// Len returns the exact canonical text length for units.
func (c Uint256Codec) Len(units uint256.Int) int {
	return Uint256Units{}.unitLen(units, c.scale())
}

// CBORLen returns the exact preferred deterministic CBOR size for units.
func (c Uint256Codec) CBORLen(units uint256.Int) int {
	_ = c.scale()
	return Uint256Units{}.unitCBORLen(units)
}

// AppendCBOR appends the preferred deterministic CBOR encoding for units. It
// allocates only when dst has insufficient capacity.
func (c Uint256Codec) AppendCBOR(dst []byte, units uint256.Int) []byte {
	_ = c.scale()
	return Uint256Units{}.unitAppendCBOR(dst, units)
}

// ParseCBOR decodes preferred deterministic CBOR into scaled units.
func (c Uint256Codec) ParseCBOR(raw []byte) (uint256.Int, Error) {
	_ = c.scale()
	return Uint256Units{}.unitParseCBOR(raw)
}

// ParseCBORInto decodes preferred deterministic CBOR into dst. It leaves dst
// unchanged on failure.
func (c Uint256Codec) ParseCBORInto(raw []byte, dst *uint256.Int) Error {
	_ = c.scale()
	if dst == nil {
		return ErrNilDestination
	}
	value, err := Uint256Units{}.unitParseCBOR(raw)
	if err != "" {
		return err
	}
	*dst = value
	return ""
}
