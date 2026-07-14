package sailfish

type decimalInput interface {
	~string | ~[]byte
}

// Decimal is an unsigned fixed-scale decimal stored as one scaled integer.
//
// Numeric value = units / 10^venue-scale.
//
// representation is optional immutable wire text. Numeric mutation clears the
// string header; it never edits string bytes. Clearing the header allocates
// nothing, and strings previously returned by String remain valid.
type Decimal[V Venue[U], U Unit] struct {
	// representation is cache state, not numeric state. The zero-length field
	// intentionally makes Decimal incomparable so callers cannot accidentally
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

// New parses s. It retains s only when s is already canonical fixed-scale
// text. Parsing is strict: no whitespace, signs, exponent notation, or excess
// fractional digits are accepted.
func New[V Venue[U], U Unit](s string) (Decimal[V, U], error) {
	scale, err := checkedScale[V, U]()
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	var venue V
	units, canonical, parseErr := venue.unitParseString(s, scale)
	if parseErr != "" {
		return Decimal[V, U]{}, boxedError(parseErr)
	}
	d := Decimal[V, U]{units: units}
	if canonical {
		d.representation = s
	}
	return d, nil
}

// NewCompact parses s without retaining its backing storage.
func NewCompact[V Venue[U], U Unit](s string) (Decimal[V, U], error) {
	scale, err := checkedScale[V, U]()
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	var venue V
	units, _, parseErr := venue.unitParseString(s, scale)
	if parseErr != "" {
		return Decimal[V, U]{}, boxedError(parseErr)
	}
	return Decimal[V, U]{units: units}, nil
}

// NewBytes parses b without retaining or converting it.
func NewBytes[V Venue[U], U Unit](b []byte) (Decimal[V, U], error) {
	scale, err := checkedScale[V, U]()
	if err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	var venue V
	units, _, parseErr := venue.unitParseBytes(b, scale)
	if parseErr != "" {
		return Decimal[V, U]{}, boxedError(parseErr)
	}
	return Decimal[V, U]{units: units}, nil
}

// NewFromUnits constructs a decimal from already-scaled units.
func NewFromUnits[V Venue[U], U Unit](units U) (Decimal[V, U], error) {
	if _, err := checkedScale[V, U](); err != "" {
		return Decimal[V, U]{}, boxedError(err)
	}
	return Decimal[V, U]{units: units}, nil
}

// Units returns the scaled integer by value. uint256.Int is an inline
// four-limb value, so the returned value owns its storage without allocation.
func (d Decimal[V, U]) Units() U { return d.units }

// SetUnits replaces the scaled integer. A value-changing update invalidates
// cached text without allocation; setting the same value preserves it.
func (d *Decimal[V, U]) SetUnits(units U) {
	if d.units == units {
		return
	}
	d.units = units
	d.representation = ""
}

func (d Decimal[V, U]) IsZero() bool {
	var zero U
	return d.units == zero
}

// HasRepresentation reports whether canonical wire text is currently
// retained.
func (d Decimal[V, U]) HasRepresentation() bool { return d.representation != "" }

// Len returns the exact canonical text length.
func (d Decimal[V, U]) Len() int {
	if d.representation != "" {
		return len(d.representation)
	}
	var venue V
	return venue.unitLen(d.units, mustScale[V, U]())
}

// AppendTo appends canonical fixed-scale text. It allocates only when dst has
// insufficient capacity.
func (d Decimal[V, U]) AppendTo(dst []byte) []byte {
	if d.representation != "" {
		return append(dst, d.representation...)
	}
	var venue V
	return venue.unitAppend(dst, d.units, mustScale[V, U]())
}

// AppendJSON appends a quoted JSON decimal string. Decimal text contains only
// ASCII digits and a decimal point, so no escaping pass is needed.
func (d Decimal[V, U]) AppendJSON(dst []byte) []byte {
	dst = append(dst, '"')
	dst = d.AppendTo(dst)
	return append(dst, '"')
}

// String returns retained text when available. Otherwise it creates exactly
// one result string allocation and does not mutate d.
func (d Decimal[V, U]) String() string {
	if d.representation != "" {
		return d.representation
	}
	var venue V
	return venue.unitString(d.units, mustScale[V, U]())
}

// Canonical returns a copy retaining canonical text. It never mutates shared
// state and is safe to use concurrently with readers of the original value.
func (d Decimal[V, U]) Canonical() Decimal[V, U] {
	if d.representation == "" {
		d.representation = d.String()
	}
	return d
}

// Compare returns -1, 0, or +1.
func (d Decimal[V, U]) Compare(other Decimal[V, U]) int {
	var venue V
	return venue.unitCompare(d.units, other.units)
}

func (d Decimal[V, U]) Cmp(other Decimal[V, U]) int { return d.Compare(other) }

func (d Decimal[V, U]) Equal(other Decimal[V, U]) bool {
	return d.units == other.units
}

func (d Decimal[V, U]) Less(other Decimal[V, U]) bool { return d.Compare(other) < 0 }

// AddOverflow returns the wrapped sum and reports unit overflow.
func (d Decimal[V, U]) AddOverflow(other Decimal[V, U]) (Decimal[V, U], bool) {
	units, overflow := addUnits(d.units, other.units)
	return Decimal[V, U]{units: units}, overflow
}

func (d Decimal[V, U]) Add(other Decimal[V, U]) (Decimal[V, U], error) {
	result, overflow := d.AddOverflow(other)
	if overflow {
		return Decimal[V, U]{}, boxedErrOverflow
	}
	return result, nil
}

// AddAssign leaves d unchanged on overflow. A value-changing success clears
// cached text without allocation; adding zero preserves it.
func (d *Decimal[V, U]) AddAssign(other Decimal[V, U]) (overflow bool) {
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
func (d Decimal[V, U]) SubUnderflow(other Decimal[V, U]) (Decimal[V, U], bool) {
	units, underflow := subUnits(d.units, other.units)
	return Decimal[V, U]{units: units}, underflow
}

func (d Decimal[V, U]) Sub(other Decimal[V, U]) (Decimal[V, U], error) {
	result, underflow := d.SubUnderflow(other)
	if underflow {
		return Decimal[V, U]{}, boxedErrUnderflow
	}
	return result, nil
}

// SubAssign leaves d unchanged on underflow. A value-changing success clears
// cached text without allocation; subtracting zero preserves it.
func (d *Decimal[V, U]) SubAssign(other Decimal[V, U]) (underflow bool) {
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

// Compare compares decimals across scales and unit backends exactly. It does
// not rescale either integer, so comparison cannot overflow.
func Compare[VA Venue[UA], UA Unit, VB Venue[UB], UB Unit](
	a Decimal[VA, UA],
	b Decimal[VB, UB],
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

	as := mustScale[VA, UA]()
	bs := mustScale[VB, UB]()

	var abuf [maxUnitDigits]byte
	var bbuf [maxUnitDigits]byte
	adLen := fillUnitDigits(&abuf, a.units)
	bdLen := fillUnitDigits(&bbuf, b.units)
	ad := abuf[:adLen]
	bd := bbuf[:bdLen]

	// Decimal digit count before the conceptual point determines magnitude.
	aExponent := adLen - as
	bExponent := bdLen - bs
	if aExponent < bExponent {
		return -1
	}
	if aExponent > bExponent {
		return 1
	}

	maxScale := max(bs, as)
	alignedLen := adLen + maxScale - as

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
