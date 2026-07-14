package sailfish

import (
	"encoding/binary"
	"math"
	"math/bits"

	"github.com/holiman/uint256"
)

// MaxCBORSize is the maximum preferred CBOR encoding size of one Decimal.
// It is tag 2, a one-byte length argument, and a 32-byte uint256 magnitude.
const MaxCBORSize = 35

const (
	cborUnsignedAdditionalUint8  = 24
	cborUnsignedAdditionalUint16 = 25
	cborUnsignedAdditionalUint32 = 26
	cborUnsignedAdditionalUint64 = 27
	cborPositiveBignumTag        = 0xc2
	cborByteStringOneByteLength  = 0x58
)

// CBORLen returns the exact size of the preferred CBOR encoding. Decimal is
// encoded as its scaled unsigned integer. Scale and retained source text are
// type/cache metadata and are intentionally absent from the wire format.
func (d Decimal[V, U]) CBORLen() int {
	var venue V
	return venue.unitCBORLen(d.units)
}

// AppendCBOR appends the preferred deterministic CBOR encoding. It allocates
// only when dst has insufficient capacity. When Decimal is a field in a
// cbor:",toarray" struct, the result is a scalar array element rather than a
// redundant nested one-element array.
func (d Decimal[V, U]) AppendCBOR(dst []byte) []byte {
	var venue V
	return venue.unitAppendCBOR(dst, d.units)
}

// MarshalCBOR implements the fxamacker/cbor Marshaler contract. The returned
// owned slice necessarily allocates once; use AppendCBOR on hot paths.
func (d Decimal[V, U]) MarshalCBOR() ([]byte, error) {
	if _, err := checkedScale[V, U](); err != "" {
		return nil, boxedError(err)
	}
	var venue V
	out := make([]byte, 0, venue.unitCBORLen(d.units))
	return venue.unitAppendCBOR(out, d.units), nil
}

// UnmarshalCBOR implements the fxamacker/cbor Unmarshaler contract. It accepts
// only RFC 8949 preferred deterministic unsigned encodings and leaves d
// unchanged on failure. Successful decode clears retained text because CBOR
// carries numeric units only.
func (d *Decimal[V, U]) UnmarshalCBOR(raw []byte) error {
	if _, err := checkedScale[V, U](); err != "" {
		return boxedError(err)
	}
	var venue V
	units, err := venue.unitParseCBOR(raw)
	if err != "" {
		return boxedError(err)
	}
	d.units = units
	d.representation = ""
	return nil
}

// CBORLen returns the exact preferred CBOR size after validating the codec.
func (c Codec[V, U]) CBORLen(d Decimal[V, U]) int {
	_ = c.scale()
	var venue V
	return venue.unitCBORLen(d.units)
}

// AppendCBOR appends preferred deterministic CBOR after validating the codec.
func (c Codec[V, U]) AppendCBOR(dst []byte, d Decimal[V, U]) []byte {
	_ = c.scale()
	var venue V
	return venue.unitAppendCBOR(dst, d.units)
}

// ParseCBOR decodes preferred deterministic CBOR without retaining raw input.
func (c Codec[V, U]) ParseCBOR(raw []byte) (Decimal[V, U], error) {
	_ = c.scale()
	var venue V
	units, err := venue.unitParseCBOR(raw)
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	return Decimal[V, U]{units: units}, nil
}

// ParseCBORFirst decodes one preferred deterministic CBOR decimal from the
// start of raw and returns the unconsumed suffix. It is the typed hot-path
// decoder for decimal fields inside manually encoded positional arrays.
// ParseCBOR remains the strict whole-item API.
func (c Codec[V, U]) ParseCBORFirst(raw []byte) (Decimal[V, U], []byte, error) {
	_ = c.scale()
	var venue V
	units, consumed, err := venue.unitParseCBORFirst(raw)
	if err != "" {
		return Decimal[V, U]{}, nil, boxedError(err)
	}
	return Decimal[V, U]{units: units}, raw[consumed:], nil
}

func (Uint8Units) unitCBORLen(units uint8) int {
	return cborUint64Len(uint64(units))
}

func (Uint8Units) unitAppendCBOR(dst []byte, units uint8) []byte {
	return appendCBORUint64(dst, uint64(units))
}

func (Uint8Units) unitParseCBOR(raw []byte) (uint8, Error) {
	value, err := parseCBORUint64(raw, math.MaxUint8)
	return uint8(value), err
}

func (Uint8Units) unitParseCBORFirst(raw []byte) (uint8, int, Error) {
	value, consumed, err := parseCBORUint64First(raw, math.MaxUint8)
	return uint8(value), consumed, err
}

func (Uint16Units) unitCBORLen(units uint16) int {
	return cborUint64Len(uint64(units))
}

func (Uint16Units) unitAppendCBOR(dst []byte, units uint16) []byte {
	return appendCBORUint64(dst, uint64(units))
}

func (Uint16Units) unitParseCBOR(raw []byte) (uint16, Error) {
	value, err := parseCBORUint64(raw, math.MaxUint16)
	return uint16(value), err
}

func (Uint16Units) unitParseCBORFirst(raw []byte) (uint16, int, Error) {
	value, consumed, err := parseCBORUint64First(raw, math.MaxUint16)
	return uint16(value), consumed, err
}

func (Uint32Units) unitCBORLen(units uint32) int {
	return cborUint64Len(uint64(units))
}

func (Uint32Units) unitAppendCBOR(dst []byte, units uint32) []byte {
	return appendCBORUint64(dst, uint64(units))
}

func (Uint32Units) unitParseCBOR(raw []byte) (uint32, Error) {
	value, err := parseCBORUint64(raw, math.MaxUint32)
	return uint32(value), err
}

func (Uint32Units) unitParseCBORFirst(raw []byte) (uint32, int, Error) {
	value, consumed, err := parseCBORUint64First(raw, math.MaxUint32)
	return uint32(value), consumed, err
}

func (Uint64Units) unitCBORLen(units uint64) int {
	return cborUint64Len(units)
}

func (Uint64Units) unitAppendCBOR(dst []byte, units uint64) []byte {
	return appendCBORUint64(dst, units)
}

func (Uint64Units) unitParseCBOR(raw []byte) (uint64, Error) {
	return parseCBORUint64(raw, math.MaxUint64)
}

func (Uint64Units) unitParseCBORFirst(raw []byte) (uint64, int, Error) {
	return parseCBORUint64First(raw, math.MaxUint64)
}

func (Uint256Units) unitCBORLen(units uint256.Int) int {
	if units[1]|units[2]|units[3] == 0 {
		return cborUint64Len(units[0])
	}
	byteLen := uint256ByteLen(units)
	return 1 + cborByteStringHeaderLen(byteLen) + byteLen
}

func (Uint256Units) unitAppendCBOR(dst []byte, units uint256.Int) []byte {
	if units[1]|units[2]|units[3] == 0 {
		return appendCBORUint64(dst, units[0])
	}

	byteLen := uint256ByteLen(units)
	dst = append(dst, cborPositiveBignumTag)
	if byteLen <= 23 {
		dst = append(dst, byte(0x40+byteLen))
	} else {
		dst = append(dst, cborByteStringOneByteLength, byte(byteLen))
	}

	var encoded [32]byte
	binary.BigEndian.PutUint64(encoded[0:8], units[3])
	binary.BigEndian.PutUint64(encoded[8:16], units[2])
	binary.BigEndian.PutUint64(encoded[16:24], units[1])
	binary.BigEndian.PutUint64(encoded[24:32], units[0])
	return append(dst, encoded[len(encoded)-byteLen:]...)
}

func (Uint256Units) unitParseCBOR(raw []byte) (uint256.Int, Error) {
	if len(raw) == 0 {
		return uint256.Int{}, ErrCBORSyntax
	}
	if raw[0] != cborPositiveBignumTag {
		value, err := parseCBORUint64(raw, math.MaxUint64)
		return uint256.Int{value}, err
	}
	if len(raw) < 2 {
		return uint256.Int{}, ErrCBORSyntax
	}

	byteLen, headerLen, err := parsePreferredCBORByteStringHeader(raw[1:])
	if err != "" {
		return uint256.Int{}, err
	}
	if byteLen <= 8 {
		return uint256.Int{}, ErrCBORNonDeterministic
	}
	if byteLen > 32 {
		return uint256.Int{}, ErrRange
	}
	begin := 1 + headerLen
	if len(raw) != begin+byteLen {
		return uint256.Int{}, ErrCBORSyntax
	}
	if raw[begin] == 0 {
		return uint256.Int{}, ErrCBORNonDeterministic
	}

	var value uint256.Int
	value.SetBytes(raw[begin:])
	return value, ""
}

func (Uint256Units) unitParseCBORFirst(raw []byte) (uint256.Int, int, Error) {
	if len(raw) == 0 {
		return uint256.Int{}, 0, ErrCBORSyntax
	}
	if raw[0] != cborPositiveBignumTag {
		value, consumed, err := parseCBORUint64First(raw, math.MaxUint64)
		return uint256.Int{value}, consumed, err
	}
	if len(raw) < 2 {
		return uint256.Int{}, 0, ErrCBORSyntax
	}

	byteLen, headerLen, err := parsePreferredCBORByteStringHeader(raw[1:])
	if err != "" {
		return uint256.Int{}, 0, err
	}
	if byteLen <= 8 {
		return uint256.Int{}, 0, ErrCBORNonDeterministic
	}
	if byteLen > 32 {
		return uint256.Int{}, 0, ErrRange
	}
	begin := 1 + headerLen
	consumed := begin + byteLen
	if len(raw) < consumed {
		return uint256.Int{}, 0, ErrCBORSyntax
	}
	if raw[begin] == 0 {
		return uint256.Int{}, 0, ErrCBORNonDeterministic
	}

	var value uint256.Int
	value.SetBytes(raw[begin:consumed])
	return value, consumed, ""
}

func cborUint64Len(value uint64) int {
	switch {
	case value <= 23:
		return 1
	case value <= math.MaxUint8:
		return 2
	case value <= math.MaxUint16:
		return 3
	case value <= math.MaxUint32:
		return 5
	default:
		return 9
	}
}

func appendCBORUint64(dst []byte, value uint64) []byte {
	switch {
	case value <= 23:
		return append(dst, byte(value))
	case value <= math.MaxUint8:
		return append(dst, cborUnsignedAdditionalUint8, byte(value))
	case value <= math.MaxUint16:
		dst = append(dst, cborUnsignedAdditionalUint16, 0, 0)
		binary.BigEndian.PutUint16(dst[len(dst)-2:], uint16(value))
		return dst
	case value <= math.MaxUint32:
		dst = append(dst, cborUnsignedAdditionalUint32, 0, 0, 0, 0)
		binary.BigEndian.PutUint32(dst[len(dst)-4:], uint32(value))
		return dst
	default:
		dst = append(dst, cborUnsignedAdditionalUint64, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(dst[len(dst)-8:], value)
		return dst
	}
}

func parseCBORUint64(raw []byte, maxValue uint64) (uint64, Error) {
	if len(raw) == 0 || raw[0]>>5 != 0 {
		return 0, ErrCBORSyntax
	}

	additional := raw[0] & 0x1f
	var value uint64
	var expectedLen int
	switch {
	case additional <= 23:
		value = uint64(additional)
		expectedLen = 1
	case additional == cborUnsignedAdditionalUint8:
		expectedLen = 2
		if len(raw) < expectedLen {
			return 0, ErrCBORSyntax
		}
		value = uint64(raw[1])
		if value <= 23 {
			return 0, ErrCBORNonDeterministic
		}
	case additional == cborUnsignedAdditionalUint16:
		expectedLen = 3
		if len(raw) < expectedLen {
			return 0, ErrCBORSyntax
		}
		value = uint64(binary.BigEndian.Uint16(raw[1:]))
		if value <= math.MaxUint8 {
			return 0, ErrCBORNonDeterministic
		}
	case additional == cborUnsignedAdditionalUint32:
		expectedLen = 5
		if len(raw) < expectedLen {
			return 0, ErrCBORSyntax
		}
		value = uint64(binary.BigEndian.Uint32(raw[1:]))
		if value <= math.MaxUint16 {
			return 0, ErrCBORNonDeterministic
		}
	case additional == cborUnsignedAdditionalUint64:
		expectedLen = 9
		if len(raw) < expectedLen {
			return 0, ErrCBORSyntax
		}
		value = binary.BigEndian.Uint64(raw[1:])
		if value <= math.MaxUint32 {
			return 0, ErrCBORNonDeterministic
		}
	default:
		return 0, ErrCBORSyntax
	}

	if len(raw) != expectedLen {
		return 0, ErrCBORSyntax
	}
	if value > maxValue {
		return 0, ErrRange
	}
	return value, ""
}

func parseCBORUint64First(raw []byte, maxValue uint64) (uint64, int, Error) {
	if len(raw) == 0 || raw[0]>>5 != 0 {
		return 0, 0, ErrCBORSyntax
	}

	additional := raw[0] & 0x1f
	var value uint64
	var expectedLen int
	switch {
	case additional <= 23:
		value = uint64(additional)
		expectedLen = 1
	case additional == cborUnsignedAdditionalUint8:
		expectedLen = 2
		if len(raw) < expectedLen {
			return 0, 0, ErrCBORSyntax
		}
		value = uint64(raw[1])
		if value <= 23 {
			return 0, 0, ErrCBORNonDeterministic
		}
	case additional == cborUnsignedAdditionalUint16:
		expectedLen = 3
		if len(raw) < expectedLen {
			return 0, 0, ErrCBORSyntax
		}
		value = uint64(binary.BigEndian.Uint16(raw[1:]))
		if value <= math.MaxUint8 {
			return 0, 0, ErrCBORNonDeterministic
		}
	case additional == cborUnsignedAdditionalUint32:
		expectedLen = 5
		if len(raw) < expectedLen {
			return 0, 0, ErrCBORSyntax
		}
		value = uint64(binary.BigEndian.Uint32(raw[1:]))
		if value <= math.MaxUint16 {
			return 0, 0, ErrCBORNonDeterministic
		}
	case additional == cborUnsignedAdditionalUint64:
		expectedLen = 9
		if len(raw) < expectedLen {
			return 0, 0, ErrCBORSyntax
		}
		value = binary.BigEndian.Uint64(raw[1:])
		if value <= math.MaxUint32 {
			return 0, 0, ErrCBORNonDeterministic
		}
	default:
		return 0, 0, ErrCBORSyntax
	}

	if value > maxValue {
		return 0, 0, ErrRange
	}
	return value, expectedLen, ""
}

func uint256ByteLen(value uint256.Int) int {
	for limb := 3; limb >= 0; limb-- {
		if value[limb] != 0 {
			return limb*8 + (bits.Len64(value[limb])+7)/8
		}
	}
	return 0
}

func cborByteStringHeaderLen(byteLen int) int {
	if byteLen <= 23 {
		return 1
	}
	return 2
}

func parsePreferredCBORByteStringHeader(raw []byte) (byteLen, headerLen int, err Error) {
	if len(raw) == 0 || raw[0]>>5 != 2 {
		return 0, 0, ErrCBORSyntax
	}
	additional := raw[0] & 0x1f
	switch {
	case additional <= 23:
		return int(additional), 1, ""
	case additional == cborUnsignedAdditionalUint8:
		if len(raw) < 2 {
			return 0, 0, ErrCBORSyntax
		}
		if raw[1] <= 23 {
			return 0, 0, ErrCBORNonDeterministic
		}
		return int(raw[1]), 2, ""
	default:
		return 0, 0, ErrCBORSyntax
	}
}
