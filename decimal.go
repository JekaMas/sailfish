package sailfish

type decimalInput interface {
	string | []byte
}

// FixedDecimal is an unsigned fixed-scale decimal stored as one scaled integer.
//
// Numeric value = units / 10^fractional-decimal-places. For example,
// FixedDecimal[PriceInUint64Units[DecimalPlaces5], uint64] stores one uint64;
// raw units 12_331_232 represent the price 123.31232.
//
// representation is optional immutable wire text. Numeric mutation clears the
// string header; it never edits string bytes. Clearing the header allocates
// nothing, and strings previously returned by String remain valid.
type FixedDecimal[V FixedDecimalFormat[U], U Unit] struct {
	// representation is cache state, not numeric state. The zero-length field
	// intentionally makes FixedDecimal incomparable so callers cannot accidentally
	// include cache state in numeric equality. It stays first because a trailing
	// zero-sized field can increase the enclosing struct size.
	_ [0]func()

	// Keep inline numeric state before the pointer-bearing string header. On
	// 64-bit Go this yields the minimum 24-byte native and 48-byte uint256
	// layouts verified by layout_test.go.
	units U

	// A string is smaller than a []byte slice header and can be returned by
	// String without another conversion or allocation.
	representation string
}

// NewFixedDecimal parses s. It retains s only when s is already canonical
// fixed-decimal text. Parsing is strict: no whitespace, signs, exponent
// notation, or excess fractional digits are accepted.
func NewFixedDecimal[V FixedDecimalFormat[U], U Unit](s string) (FixedDecimal[V, U], error) {
	decimalPlaces, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	var format V
	units, canonical, parseErr := format.unitParseString(s, decimalPlaces)
	if parseErr != "" {
		return FixedDecimal[V, U]{}, boxedError(parseErr)
	}
	d := FixedDecimal[V, U]{units: units}
	if canonical {
		d.representation = s
	}
	return d, nil
}

// NewCompactFixedDecimal parses s without retaining its backing storage.
func NewCompactFixedDecimal[V FixedDecimalFormat[U], U Unit](s string) (FixedDecimal[V, U], error) {
	decimalPlaces, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	var format V
	units, _, parseErr := format.unitParseString(s, decimalPlaces)
	if parseErr != "" {
		return FixedDecimal[V, U]{}, boxedError(parseErr)
	}
	return FixedDecimal[V, U]{units: units}, nil
}

// NewFixedDecimalFromBytes parses b without retaining or converting it.
func NewFixedDecimalFromBytes[V FixedDecimalFormat[U], U Unit](b []byte) (FixedDecimal[V, U], error) {
	decimalPlaces, err := checkedFractionalDecimalPlaces[V, U]()
	if err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	var format V
	units, _, parseErr := format.unitParseBytes(b, decimalPlaces)
	if parseErr != "" {
		return FixedDecimal[V, U]{}, boxedError(parseErr)
	}
	return FixedDecimal[V, U]{units: units}, nil
}

// NewFixedDecimalFromUnits constructs a decimal from already-scaled units.
func NewFixedDecimalFromUnits[V FixedDecimalFormat[U], U Unit](units U) (FixedDecimal[V, U], error) {
	if _, err := checkedFractionalDecimalPlaces[V, U](); err != "" {
		return FixedDecimal[V, U]{}, boxedError(err)
	}
	return FixedDecimal[V, U]{units: units}, nil
}

// Units returns the scaled integer by value. uint256.Int is an inline
// four-limb value, so the returned value owns its storage without allocation.
func (d FixedDecimal[V, U]) Units() U { return d.units }

// SetUnits replaces the scaled integer. A value-changing update invalidates
// cached text without allocation; setting the same value preserves it.
func (d *FixedDecimal[V, U]) SetUnits(units U) {
	if d.units == units {
		return
	}
	d.units = units
	d.representation = ""
}

func (d FixedDecimal[V, U]) IsZero() bool {
	var zero U
	return d.units == zero
}

// HasRepresentation reports whether canonical wire text is currently
// retained.
func (d FixedDecimal[V, U]) HasRepresentation() bool { return d.representation != "" }

// Len returns the exact canonical text length.
func (d FixedDecimal[V, U]) Len() int {
	if d.representation != "" {
		return len(d.representation)
	}
	var format V
	return format.unitLen(d.units, fractionalDecimalPlacesOf[V, U]())
}

// AppendTo appends canonical fixed-decimal text. It allocates only when dst
// has insufficient capacity.
func (d FixedDecimal[V, U]) AppendTo(dst []byte) []byte {
	if d.representation != "" {
		return append(dst, d.representation...)
	}
	var format V
	return format.unitAppend(dst, d.units, fractionalDecimalPlacesOf[V, U]())
}

// AppendJSON appends a quoted JSON decimal string. FixedDecimal text contains only
// ASCII digits and a decimal point, so no escaping pass is needed.
func (d FixedDecimal[V, U]) AppendJSON(dst []byte) []byte {
	dst = append(dst, '"')
	dst = d.AppendTo(dst)
	return append(dst, '"')
}

// String returns retained text when available. Otherwise it creates exactly
// one result string allocation and does not mutate d.
func (d FixedDecimal[V, U]) String() string {
	if d.representation != "" {
		return d.representation
	}
	var format V
	return format.unitString(d.units, fractionalDecimalPlacesOf[V, U]())
}

// Canonical returns a copy retaining canonical text. It never mutates shared
// state and is safe to use concurrently with readers of the original value.
func (d FixedDecimal[V, U]) Canonical() FixedDecimal[V, U] {
	if d.representation == "" {
		d.representation = d.String()
	}
	return d
}

// Compare returns -1, 0, or +1.
func (d FixedDecimal[V, U]) Compare(other FixedDecimal[V, U]) int {
	var format V
	return format.unitCompare(d.units, other.units)
}

func (d FixedDecimal[V, U]) Cmp(other FixedDecimal[V, U]) int { return d.Compare(other) }

func (d FixedDecimal[V, U]) Equal(other FixedDecimal[V, U]) bool {
	return d.units == other.units
}

func (d FixedDecimal[V, U]) Less(other FixedDecimal[V, U]) bool { return d.Compare(other) < 0 }

// AddOverflow returns the wrapped sum and reports unit overflow.
func (d FixedDecimal[V, U]) AddOverflow(other FixedDecimal[V, U]) (FixedDecimal[V, U], bool) {
	units, overflow := addUnits(d.units, other.units)
	return FixedDecimal[V, U]{units: units}, overflow
}

func (d FixedDecimal[V, U]) Add(other FixedDecimal[V, U]) (FixedDecimal[V, U], error) {
	result, overflow := d.AddOverflow(other)
	if overflow {
		return FixedDecimal[V, U]{}, boxedErrOverflow
	}
	return result, nil
}

// AddAssign leaves d unchanged on overflow. A value-changing success clears
// cached text without allocation; adding zero preserves it.
func (d *FixedDecimal[V, U]) AddAssign(other FixedDecimal[V, U]) (overflow bool) {
	units, overflow := addUnits(d.units, other.units)
	if overflow {
		return true
	}
	if units == d.units {
		return false
	}
	d.units = units
	d.representation = ""
	return false
}

// SubUnderflow returns the wrapped difference and reports unit underflow.
func (d FixedDecimal[V, U]) SubUnderflow(other FixedDecimal[V, U]) (FixedDecimal[V, U], bool) {
	units, underflow := subUnits(d.units, other.units)
	return FixedDecimal[V, U]{units: units}, underflow
}

func (d FixedDecimal[V, U]) Sub(other FixedDecimal[V, U]) (FixedDecimal[V, U], error) {
	result, underflow := d.SubUnderflow(other)
	if underflow {
		return FixedDecimal[V, U]{}, boxedErrUnderflow
	}
	return result, nil
}

// SubAssign leaves d unchanged on underflow. A value-changing success clears
// cached text without allocation; subtracting zero preserves it.
func (d *FixedDecimal[V, U]) SubAssign(other FixedDecimal[V, U]) (underflow bool) {
	units, underflow := subUnits(d.units, other.units)
	if underflow {
		return true
	}
	if units == d.units {
		return false
	}
	d.units = units
	d.representation = ""
	return false
}

// Compare compares fixed decimals across fractional decimal-place counts and
// unit backends exactly. It does not rescale either integer, so comparison
// cannot overflow.
func Compare[VA FixedDecimalFormat[UA], UA Unit, VB FixedDecimalFormat[UB], UB Unit](
	a FixedDecimal[VA, UA],
	b FixedDecimal[VB, UB],
) int {
	var azero UA
	var bzero UB
	az := a.units == azero
	bz := b.units == bzero
	if az || bz {
		switch {
		case az && bz:
			return 0
		case az:
			return -1
		default:
			return 1
		}
	}

	as := fractionalDecimalPlacesOf[VA, UA]()
	bs := fractionalDecimalPlacesOf[VB, UB]()

	var abuf [maxUnitDigits]byte
	var bbuf [maxUnitDigits]byte
	adLen := fillUnitDigits(&abuf, a.units)
	bdLen := fillUnitDigits(&bbuf, b.units)
	ad := abuf[:adLen]
	bd := bbuf[:bdLen]

	// FixedDecimal digit count before the conceptual point determines magnitude.
	aExponent := adLen - as
	bExponent := bdLen - bs
	if aExponent < bExponent {
		return -1
	}
	if aExponent > bExponent {
		return 1
	}

	maxFractionalDecimalPlaces := max(bs, as)
	alignedLen := adLen + maxFractionalDecimalPlaces - as

	// Compare scaled integers with conceptual trailing zeroes.
	for i := range alignedLen {
		ac := byte('0')
		if i < len(ad) {
			ac = ad[i]
		}
		bc := byte('0')
		if i < len(bd) {
			bc = bd[i]
		}
		if ac < bc {
			return -1
		}
		if ac > bc {
			return 1
		}
	}
	return 0
}
