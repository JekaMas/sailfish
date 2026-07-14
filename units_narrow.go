package sailfish

const (
	maxUint8Scale    = 2
	maxUint8TextLen  = 4
	maxUint16Scale   = 4
	maxUint16TextLen = 6
	maxUint32Scale   = 9
	maxUint32TextLen = 11
)

// Uint8Units, Uint16Units, and Uint32Units are zero-sized unit providers.
// Embed one in a custom venue, or use the PriceUint* and AmountUint* formats.
type Uint8Units struct{}

func (Uint8Units) unitParseString(s string, scale int) (uint8, bool, Error) {
	return parseUint8(s, scale)
}

func (Uint8Units) unitParseBytes(b []byte, scale int) (uint8, bool, Error) {
	return parseUint8(b, scale)
}

func (Uint8Units) unitAppend(dst []byte, units uint8, scale int) []byte {
	return appendUint64Decimal(dst, uint64(units), scale)
}

func (Uint8Units) unitString(units uint8, scale int) string {
	var buf [maxUint8TextLen]byte
	out := appendUint64Decimal(buf[:0], uint64(units), scale)
	return string(out)
}

func (Uint8Units) unitLen(units uint8, scale int) int {
	return scaledTextLen(decimalDigits64(uint64(units)), scale)
}

func (Uint8Units) unitCompare(a, b uint8) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

type Uint16Units struct{}

func (Uint16Units) unitParseString(s string, scale int) (uint16, bool, Error) {
	return parseUint16(s, scale)
}

func (Uint16Units) unitParseBytes(b []byte, scale int) (uint16, bool, Error) {
	return parseUint16(b, scale)
}

func (Uint16Units) unitAppend(dst []byte, units uint16, scale int) []byte {
	return appendUint64Decimal(dst, uint64(units), scale)
}

func (Uint16Units) unitString(units uint16, scale int) string {
	var buf [maxUint16TextLen]byte
	out := appendUint64Decimal(buf[:0], uint64(units), scale)
	return string(out)
}

func (Uint16Units) unitLen(units uint16, scale int) int {
	return scaledTextLen(decimalDigits64(uint64(units)), scale)
}

func (Uint16Units) unitCompare(a, b uint16) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

type Uint32Units struct{}

func (Uint32Units) unitParseString(s string, scale int) (uint32, bool, Error) {
	return parseUint32(s, scale)
}

func (Uint32Units) unitParseBytes(b []byte, scale int) (uint32, bool, Error) {
	return parseUint32(b, scale)
}

func (Uint32Units) unitAppend(dst []byte, units uint32, scale int) []byte {
	return appendUint64Decimal(dst, uint64(units), scale)
}

func (Uint32Units) unitString(units uint32, scale int) string {
	var buf [maxUint32TextLen]byte
	out := appendUint64Decimal(buf[:0], uint64(units), scale)
	return string(out)
}

func (Uint32Units) unitLen(units uint32, scale int) int {
	return scaledTextLen(decimalDigits64(uint64(units)), scale)
}

func (Uint32Units) unitCompare(a, b uint32) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func parseUint8[S decimalInput](s S, scale int) (uint8, bool, Error) {
	value, canonical, err := parseUint64(s, scale)
	if err != "" {
		return 0, false, err
	}
	if value > uint64(^uint8(0)) {
		return 0, false, ErrRange
	}
	return uint8(value), canonical, ""
}

func parseUint16[S decimalInput](s S, scale int) (uint16, bool, Error) {
	value, canonical, err := parseUint64(s, scale)
	if err != "" {
		return 0, false, err
	}
	if value > uint64(^uint16(0)) {
		return 0, false, ErrRange
	}
	return uint16(value), canonical, ""
}

func parseUint32[S decimalInput](s S, scale int) (uint32, bool, Error) {
	value, canonical, err := parseUint64(s, scale)
	if err != "" {
		return 0, false, err
	}
	if value > uint64(^uint32(0)) {
		return 0, false, ErrRange
	}
	return uint32(value), canonical, ""
}
