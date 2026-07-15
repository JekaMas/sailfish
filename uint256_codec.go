package sailfish

import "github.com/holiman/uint256"

// Uint256FixedDecimalCodec is the runtime-decimal-places hot-path codec for
// scaled uint256 units.
//
// Use FixedDecimalCodec with a FixedDecimalFormat when compile-time semantic
// identity and decimal places are required. Use
// Uint256FixedDecimalCodec at boundaries where trusted metadata resolves the
// fractional decimal places at runtime, such as CEX symbol decoding. Its
// methods avoid generic format
// dispatch and return Error directly so successful and rejected parses remain
// allocation-free.
//
// The zero value represents zero fractional decimal places. The number of
// decimal places is stored directly in one byte so repeated boundary
// operations do not decode constructor metadata.
type Uint256FixedDecimalCodec struct {
	fractionalDecimalPlacesValue uint8
}

// NewUint256FixedDecimalCodec validates the exact number of fractional decimal
// places once for repeated uint256 operations.
func NewUint256FixedDecimalCodec(
	fractionalDecimalPlaces DecimalPlaces,
) (Uint256FixedDecimalCodec, error) {
	if int(fractionalDecimalPlaces) > maxUint256Scale {
		return Uint256FixedDecimalCodec{}, boxedErrUnsupportedFractionalDecimalPlaces
	}
	return Uint256FixedDecimalCodec{
		fractionalDecimalPlacesValue: uint8(fractionalDecimalPlaces),
	}, nil
}

func (c Uint256FixedDecimalCodec) fractionalDecimalPlaces() int {
	return int(c.fractionalDecimalPlacesValue)
}

// FractionalDecimalPlaces returns the exact number of digits represented
// after the decimal point.
func (c Uint256FixedDecimalCodec) FractionalDecimalPlaces() DecimalPlaces {
	return DecimalPlaces(c.fractionalDecimalPlaces())
}

// MaxIntegerDigits reports the maximum integer-part digit count for the
// configured fractional decimal places.
func (c Uint256FixedDecimalCodec) MaxIntegerDigits() int {
	return unitDecimalDigits[uint256.Int]() - c.fractionalDecimalPlaces()
}

// Parse parses a strict non-negative decimal string into scaled units.
func (c Uint256FixedDecimalCodec) Parse(input string) (uint256.Int, Error) {
	value, _, err := parseUint256(input, c.fractionalDecimalPlaces())
	return value, err
}

// ParseBytes parses input without converting it to a string.
func (c Uint256FixedDecimalCodec) ParseBytes(input []byte) (uint256.Int, Error) {
	value, _, err := parseUint256(input, c.fractionalDecimalPlaces())
	return value, err
}

// ParseInto parses input into dst. It leaves dst unchanged on failure.
func (c Uint256FixedDecimalCodec) ParseInto(input string, dst *uint256.Int) Error {
	decimalPlaces := c.fractionalDecimalPlaces()
	if dst == nil {
		return ErrNilDestination
	}
	value, _, err := parseUint256(input, decimalPlaces)
	if err != "" {
		return err
	}
	*dst = value
	return ""
}

// ParseBytesInto parses input into dst without converting it to a string. It
// leaves dst unchanged on failure.
func (c Uint256FixedDecimalCodec) ParseBytesInto(input []byte, dst *uint256.Int) Error {
	decimalPlaces := c.fractionalDecimalPlaces()
	if dst == nil {
		return ErrNilDestination
	}
	value, _, err := parseUint256(input, decimalPlaces)
	if err != "" {
		return err
	}
	*dst = value
	return ""
}

// AppendTo appends canonical fixed-scale text for units. It allocates only
// when dst has insufficient capacity.
func (c Uint256FixedDecimalCodec) AppendTo(dst []byte, units uint256.Int) []byte {
	return appendUint256Decimal(dst, units, c.fractionalDecimalPlaces())
}

// Len returns the exact canonical text length for units.
func (c Uint256FixedDecimalCodec) Len(units uint256.Int) int {
	return Uint256Units{}.unitLen(units, c.fractionalDecimalPlaces())
}

// CBORLen returns the exact preferred deterministic CBOR size for units.
func (c Uint256FixedDecimalCodec) CBORLen(units uint256.Int) int {
	return Uint256Units{}.unitCBORLen(units)
}

// AppendCBOR appends the preferred deterministic CBOR encoding for units. It
// allocates only when dst has insufficient capacity.
func (c Uint256FixedDecimalCodec) AppendCBOR(dst []byte, units uint256.Int) []byte {
	return Uint256Units{}.unitAppendCBOR(dst, units)
}

// ParseCBOR decodes preferred deterministic CBOR into scaled units.
func (c Uint256FixedDecimalCodec) ParseCBOR(raw []byte) (uint256.Int, Error) {
	return Uint256Units{}.unitParseCBOR(raw)
}

// ParseCBORFirst decodes one preferred deterministic CBOR uint256 from the
// start of raw and returns the unconsumed suffix. It is intended for manual
// positional-array decoders that keep aggregate decoding allocation-free.
func (c Uint256FixedDecimalCodec) ParseCBORFirst(raw []byte) (uint256.Int, []byte, Error) {
	value, consumed, err := Uint256Units{}.unitParseCBORFirst(raw)
	if err != "" {
		return uint256.Int{}, nil, err
	}
	return value, raw[consumed:], ""
}

// ParseCBORInto decodes preferred deterministic CBOR into dst. It leaves dst
// unchanged on failure.
func (c Uint256FixedDecimalCodec) ParseCBORInto(raw []byte, dst *uint256.Int) Error {
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

// ParseCBORFirstInto decodes one preferred deterministic CBOR uint256 into
// dst and returns the unconsumed suffix. It leaves dst unchanged on failure.
func (c Uint256FixedDecimalCodec) ParseCBORFirstInto(raw []byte, dst *uint256.Int) ([]byte, Error) {
	if dst == nil {
		return nil, ErrNilDestination
	}
	value, consumed, err := Uint256Units{}.unitParseCBORFirst(raw)
	if err != "" {
		return nil, err
	}
	*dst = value
	return raw[consumed:], ""
}
